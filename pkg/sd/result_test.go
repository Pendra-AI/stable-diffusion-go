package sd

import (
	"bytes"
	"image/png"
	"testing"
	"unsafe"
)

func cstr(s string) *uint8 {
	b := append([]byte(s), 0)
	return &b[0]
}

// stubFreeFuncs replaces the native free functions with counters for the length
// of a test. free_sd_images is included so a result freed through it fails the
// test: its items are not an SDImage array.
func stubFreeFuncs(t *testing.T) (imageFrees, videoFrees *int) {
	t.Helper()
	saveImage, saveVideo, saveImages := freeSDImageGenerationResult, freeSDVideoGenerationResult, freeSDImages
	t.Cleanup(func() {
		freeSDImageGenerationResult, freeSDVideoGenerationResult, freeSDImages = saveImage, saveVideo, saveImages
	})
	imageFrees, videoFrees = new(int), new(int)
	freeSDImageGenerationResult = func(*cImageGenerationResult) { *imageFrees++ }
	freeSDVideoGenerationResult = func(*cVideoGenerationResult) { *videoFrees++ }
	freeSDImages = func(*SDImage, int32) { t.Error("free_sd_images must never be called on a result") }
	return imageFrees, videoFrees
}

// fakeImageResult builds a result laid out like the native one, holding a 1x1 RGB
// image per colour.
func fakeImageResult(colours ...[3]uint8) *cImageGenerationResult {
	items := make([]SDImageResult, len(colours))
	for i, c := range colours {
		pixels := []uint8{c[0], c[1], c[2]}
		items[i] = SDImageResult{
			Image:          SDImage{Width: 1, Height: 1, Channel: 3, Data: &pixels[0]},
			GenerationInfo: cstr(`{"index":` + string(rune('0'+i)) + `}`),
		}
	}
	return &cImageGenerationResult{Items: &items[0], Count: int32(len(items))}
}

func fakeVideoResult() *cVideoGenerationResult {
	a, b := []uint8{1, 2, 3}, []uint8{4, 5, 6}
	frames := []SDImage{
		{Width: 1, Height: 1, Channel: 3, Data: &a[0]},
		{Width: 1, Height: 1, Channel: 3, Data: &b[0]},
	}
	return &cVideoGenerationResult{
		Frames:         &frames[0],
		FrameCount:     2,
		Audio:          &SDAudio{SampleRate: 48000, Channels: 2},
		Fps:            16,
		GenerationInfo: cstr(`{"fps":16}`),
	}
}

// The items are 32 bytes (sd_image_result_t), not 24 (sd_image_t), so reading the
// second item checks the stride as well as the accessors.
func TestImageResultAccessors(t *testing.T) {
	stubFreeFuncs(t)
	res := newImageGenerationResult(true, fakeImageResult([3]uint8{1, 2, 3}, [3]uint8{4, 5, 6}))
	defer res.Free()

	if got := res.Count(); got != 2 {
		t.Fatalf("Count() = %d, want 2", got)
	}
	second := res.Image(1)
	if second == nil {
		t.Fatal("Image(1) = nil")
	}
	if got := unsafe.Slice(second.Data, 3); got[0] != 4 || got[1] != 5 || got[2] != 6 {
		t.Errorf("Image(1) pixels = %v, want [4 5 6]", got)
	}
	if got := res.GenerationInfo(1); got != `{"index":1}` {
		t.Errorf("GenerationInfo(1) = %q", got)
	}
	for _, i := range []int{-1, 2, 1 << 30} {
		if res.Image(i) != nil {
			t.Errorf("Image(%d) != nil", i)
		}
		if got := res.GenerationInfo(i); got != "" {
			t.Errorf("GenerationInfo(%d) = %q, want empty", i, got)
		}
	}
}

// The caller reads the pixels through the handle before freeing it.
func TestImageResultEncodesBeforeFree(t *testing.T) {
	imageFrees, _ := stubFreeFuncs(t)
	res := newImageGenerationResult(true, fakeImageResult([3]uint8{10, 20, 30}))

	encoded, err := EncodePNG(res.Image(0))
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	res.Free()
	if *imageFrees != 1 {
		t.Errorf("free calls = %d, want 1", *imageFrees)
	}

	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	if r, g, b, _ := decoded.At(0, 0).RGBA(); r>>8 != 10 || g>>8 != 20 || b>>8 != 30 {
		t.Errorf("decoded pixel = (%d,%d,%d), want (10,20,30)", r>>8, g>>8, b>>8)
	}
}

func TestImageResultSlotWithoutPixels(t *testing.T) {
	stubFreeFuncs(t)
	raw := fakeImageResult([3]uint8{1, 2, 3})
	raw.Items.Image.Data = nil
	res := newImageGenerationResult(true, raw)
	defer res.Free()

	if res.Image(0) != nil {
		t.Error("Image(0) != nil for a slot without pixels")
	}
	if got := res.GenerationInfo(0); got == "" {
		t.Error("GenerationInfo(0) lost its value")
	}
}

func TestImageResultFreeIsIdempotent(t *testing.T) {
	imageFrees, _ := stubFreeFuncs(t)
	res := newImageGenerationResult(true, fakeImageResult([3]uint8{1, 2, 3}))
	res.Free()
	res.Free()
	if *imageFrees != 1 {
		t.Errorf("free calls = %d, want 1", *imageFrees)
	}
	if res.Count() != 0 || res.Image(0) != nil || res.GenerationInfo(0) != "" {
		t.Error("a freed result still exposes native memory")
	}
}

func TestResultsAreNilSafe(t *testing.T) {
	var img *SDImageGenerationResult
	if img.Count() != 0 || img.Image(0) != nil || img.GenerationInfo(0) != "" {
		t.Error("nil image result is not empty")
	}
	img.Free()

	var vid *SDVideoGenerationResult
	if vid.FrameCount() != 0 || vid.Frame(0) != nil || vid.Frames() != nil || vid.Audio() != nil ||
		vid.Fps() != 0 || vid.GenerationInfo() != "" {
		t.Error("nil video result is not empty")
	}
	vid.Free()
}

// Free must not panic when the free binding is missing, as it is before Load.
func TestResultFreeWithoutBinding(t *testing.T) {
	saveImage, saveVideo := freeSDImageGenerationResult, freeSDVideoGenerationResult
	t.Cleanup(func() { freeSDImageGenerationResult, freeSDVideoGenerationResult = saveImage, saveVideo })
	freeSDImageGenerationResult, freeSDVideoGenerationResult = nil, nil

	newImageGenerationResult(true, fakeImageResult([3]uint8{1, 2, 3})).Free()
	newVideoGenerationResult(true, fakeVideoResult()).Free()
}

func TestGenerateImageOutcomes(t *testing.T) {
	saveGenerate := generateImage
	t.Cleanup(func() { generateImage = saveGenerate })

	emptyItems := fakeImageResult([3]uint8{1, 2, 3})
	emptyItems.Items = nil
	noCount := fakeImageResult([3]uint8{1, 2, 3})
	noCount.Count = 0
	negativeCount := fakeImageResult([3]uint8{1, 2, 3})
	negativeCount.Count = -1

	cases := []struct {
		name      string
		ok        bool
		raw       *cImageGenerationResult
		wantNil   bool
		wantFrees int
	}{
		{"failure sets no result", false, nil, true, 0},
		{"success without a result", true, nil, true, 0},
		{"failure that still set a result is released", false, fakeImageResult([3]uint8{1, 2, 3}), true, 1},
		{"no items", true, emptyItems, true, 1},
		{"zero count", true, noCount, true, 1},
		{"negative count", true, negativeCount, true, 1},
		{"success", true, fakeImageResult([3]uint8{1, 2, 3}), false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			imageFrees, _ := stubFreeFuncs(t)
			generateImage = func(_ unsafe.Pointer, _ *SDImgGenParams, out **cImageGenerationResult) bool {
				*out = c.raw
				return c.ok
			}
			res := (&SDContext{}).GenerateImage(&SDImgGenParams{})
			if (res == nil) != c.wantNil {
				t.Fatalf("GenerateImage() = %v, wantNil %v", res, c.wantNil)
			}
			if *imageFrees != c.wantFrees {
				t.Errorf("free calls = %d, want %d", *imageFrees, c.wantFrees)
			}
		})
	}
}

// Generating without a loaded library or a context returns nil, not a panic.
func TestGenerateWithoutLibrary(t *testing.T) {
	saveImage, saveVideo, saveUpscale := generateImage, generateVideo, upscale
	t.Cleanup(func() { generateImage, generateVideo, upscale = saveImage, saveVideo, saveUpscale })
	generateImage, generateVideo, upscale = nil, nil, nil

	if (&SDContext{}).GenerateImage(nil) != nil {
		t.Error("GenerateImage without a library != nil")
	}
	if (&SDContext{}).GenerateVideo(nil) != nil {
		t.Error("GenerateVideo without a library != nil")
	}
	if (&UpscalerContext{}).Upscale(SDImage{}, 2) != nil {
		t.Error("Upscale without a library != nil")
	}

	var ctx *SDContext
	if ctx.GenerateImage(nil) != nil || ctx.GenerateVideo(nil) != nil {
		t.Error("generating on a nil context != nil")
	}
	var up *UpscalerContext
	if up.Upscale(SDImage{}, 2) != nil {
		t.Error("Upscale on a nil context != nil")
	}
}

func TestUpscaleWrapsResult(t *testing.T) {
	saveUpscale, saveSupported := upscale, byValueImageArgsSupported
	t.Cleanup(func() { upscale, byValueImageArgsSupported = saveUpscale, saveSupported })
	byValueImageArgsSupported = true
	imageFrees, _ := stubFreeFuncs(t)

	var gotFactor uint32
	upscale = func(_ unsafe.Pointer, _ *SDImage, factor uint32, out **cImageGenerationResult) bool {
		gotFactor = factor
		*out = fakeImageResult([3]uint8{7, 8, 9})
		return true
	}
	res := (&UpscalerContext{}).Upscale(SDImage{}, 4)
	if res.Count() != 1 || gotFactor != 4 {
		t.Fatalf("Count() = %d, factor = %d", res.Count(), gotFactor)
	}
	res.Free()
	if *imageFrees != 1 {
		t.Errorf("free calls = %d, want 1", *imageFrees)
	}

	upscale = func(unsafe.Pointer, *SDImage, uint32, **cImageGenerationResult) bool { return false }
	if (&UpscalerContext{}).Upscale(SDImage{}, 4) != nil {
		t.Error("failed Upscale != nil")
	}
}

// Where sd_image_t cannot be passed by value, the native call would crash, so it
// must not be made at all.
func TestByValueImageCallsAreRefusedWhereUnsupported(t *testing.T) {
	saveUpscale, saveCanny, saveSupported := upscale, preprocessCanny, byValueImageArgsSupported
	t.Cleanup(func() { upscale, preprocessCanny, byValueImageArgsSupported = saveUpscale, saveCanny, saveSupported })
	byValueImageArgsSupported = false

	upscale = func(unsafe.Pointer, *SDImage, uint32, **cImageGenerationResult) bool {
		t.Error("upscale was called")
		return true
	}
	preprocessCanny = func(*SDImage, float32, float32, float32, float32, bool) bool {
		t.Error("preprocess_canny was called")
		return true
	}
	if (&UpscalerContext{}).Upscale(SDImage{}, 2) != nil {
		t.Error("Upscale != nil")
	}
	if PreprocessCanny(SDImage{}, 0, 0, 0, 0, false) {
		t.Error("PreprocessCanny = true")
	}
}

func TestVideoResultAccessors(t *testing.T) {
	_, videoFrees := stubFreeFuncs(t)
	res := newVideoGenerationResult(true, fakeVideoResult())

	if res.FrameCount() != 2 || res.Fps() != 16 || res.GenerationInfo() != `{"fps":16}` {
		t.Errorf("FrameCount=%d Fps=%d GenerationInfo=%q", res.FrameCount(), res.Fps(), res.GenerationInfo())
	}
	if res.Audio() == nil || res.Audio().SampleRate != 48000 {
		t.Errorf("Audio() = %+v", res.Audio())
	}
	second := res.Frame(1)
	if second == nil || *second.Data != 4 {
		t.Fatalf("Frame(1) = %+v", second)
	}
	frames := res.Frames()
	if len(frames) != 2 || *frames[1].Data != 4 {
		t.Fatalf("Frames() = %+v", frames)
	}
	if res.Frame(-1) != nil || res.Frame(2) != nil {
		t.Error("out-of-range Frame != nil")
	}

	res.Free()
	res.Free()
	if *videoFrees != 1 {
		t.Errorf("free calls = %d, want 1", *videoFrees)
	}
	if res.FrameCount() != 0 || res.Frames() != nil || res.Audio() != nil || res.Fps() != 0 {
		t.Error("a freed result still exposes native memory")
	}
}

func TestGenerateVideoOutcomes(t *testing.T) {
	saveGenerate := generateVideo
	t.Cleanup(func() { generateVideo = saveGenerate })

	noFrames := fakeVideoResult()
	noFrames.Frames = nil
	noCount := fakeVideoResult()
	noCount.FrameCount = 0

	cases := []struct {
		name      string
		ok        bool
		raw       *cVideoGenerationResult
		wantNil   bool
		wantFrees int
	}{
		{"failure sets no result", false, nil, true, 0},
		{"failure that still set a result is released", false, fakeVideoResult(), true, 1},
		{"no frames", true, noFrames, true, 1},
		{"zero count", true, noCount, true, 1},
		{"success", true, fakeVideoResult(), false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, videoFrees := stubFreeFuncs(t)
			generateVideo = func(_ unsafe.Pointer, _ *SDVidGenParams, out **cVideoGenerationResult) bool {
				*out = c.raw
				return c.ok
			}
			res := (&SDContext{}).GenerateVideo(&SDVidGenParams{})
			if (res == nil) != c.wantNil {
				t.Fatalf("GenerateVideo() = %v, wantNil %v", res, c.wantNil)
			}
			if *videoFrees != c.wantFrees {
				t.Errorf("free calls = %d, want %d", *videoFrees, c.wantFrees)
			}
		})
	}
}

func TestPreviewInfoFromRegister(t *testing.T) {
	got := previewInfoFromRegister(20<<32 | 3)
	if want := (SDPreviewInfo{SamplePass: 3, TotalSteps: 20}); got != want {
		t.Errorf("previewInfoFromRegister = %+v, want %+v", got, want)
	}
}

func TestContextQueriesWithoutLibrary(t *testing.T) {
	savePreview, saveName := sdGetPreviewInfo, sdGetModelVersionName
	t.Cleanup(func() { sdGetPreviewInfo, sdGetModelVersionName = savePreview, saveName })
	sdGetPreviewInfo, sdGetModelVersionName = nil, nil

	if got := GetPreviewInfo(); got != (SDPreviewInfo{}) {
		t.Errorf("GetPreviewInfo() = %+v, want zero", got)
	}
	if got := (&SDContext{}).ModelVersionName(); got != "" {
		t.Errorf("ModelVersionName() = %q, want empty", got)
	}
}
