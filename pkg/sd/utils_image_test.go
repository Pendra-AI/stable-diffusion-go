package sd

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

// TestToNRGBAKeepsAlphaOfFourChannels feeds a 4-channel (RGBA) buffer and checks
// the alpha survives, i.e. the per-pixel stride honours Channel and the fourth
// channel is not replaced by an opaque one.
func TestToNRGBAKeepsAlphaOfFourChannels(t *testing.T) {
	const w, h = 2, 1
	data := []uint8{10, 20, 30, 40, 50, 60, 70, 80}
	img := &SDImage{Width: w, Height: h, Channel: 4, Data: &data[0]}

	nrgba, err := toNRGBA(img)
	if err != nil {
		t.Fatalf("toNRGBA returned error: %v", err)
	}
	if !bytes.Equal(nrgba.Pix, data) {
		t.Errorf("Pix = %v, want %v", nrgba.Pix, data)
	}
}

// Three channels come out opaque, and so does anything wider than RGBA, where the
// extra channels have no agreed meaning.
func TestToNRGBAOpaqueForOtherChannelCounts(t *testing.T) {
	cases := []struct {
		channel uint32
		data    []uint8
	}{
		{3, []uint8{10, 20, 30, 40, 50, 60}},
		{5, []uint8{10, 20, 30, 99, 99, 40, 50, 60, 99, 99}},
	}
	for _, c := range cases {
		nrgba, err := toNRGBA(&SDImage{Width: 2, Height: 1, Channel: c.channel, Data: &c.data[0]})
		if err != nil {
			t.Fatalf("channel %d: %v", c.channel, err)
		}
		if want := []uint8{10, 20, 30, 255, 40, 50, 60, 255}; !bytes.Equal(nrgba.Pix, want) {
			t.Errorf("channel %d: Pix = %v, want %v", c.channel, nrgba.Pix, want)
		}
		if !nrgba.Opaque() {
			t.Errorf("channel %d: image is not opaque", c.channel)
		}
	}
}

// decodeNRGBA decodes PNG bytes and returns the straight, non-premultiplied
// pixels. It reads Pix because color.Color.RGBA premultiplies, which would report
// a fully transparent pixel as 0,0,0,0 whatever the file holds.
func decodeNRGBA(t *testing.T, encoded []byte) *image.NRGBA {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	nrgba, ok := decoded.(*image.NRGBA)
	if !ok {
		t.Fatalf("decoded image is %T, want *image.NRGBA (a PNG with an alpha channel)", decoded)
	}
	return nrgba
}

// A generated RGBA image reaches the PNG straight: a fully transparent pixel
// keeps its RGB, and a translucent one is not darkened by premultiplying.
func TestEncodePNGPreservesAlpha(t *testing.T) {
	data := []uint8{
		200, 100, 50, 0, // fully transparent, RGB must survive
		200, 100, 50, 128, // translucent
		1, 2, 3, 255, // opaque
	}
	encoded, err := EncodePNG(&SDImage{Width: 3, Height: 1, Channel: 4, Data: &data[0]})
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	if got := decodeNRGBA(t, encoded); !bytes.Equal(got.Pix, data) {
		t.Errorf("decoded Pix = %v, want %v", got.Pix, data)
	}
}

func TestEncodePNGThreeChannelsStayRGB(t *testing.T) {
	data := []uint8{9, 8, 7}
	encoded, err := EncodePNG(&SDImage{Width: 1, Height: 1, Channel: 3, Data: &data[0]})
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	if _, ok := decoded.(*image.RGBA); ok {
		return // RGB PNGs decode to *image.RGBA with opaque alpha
	}
	t.Errorf("decoded image is %T, want an alpha-less PNG", decoded)
}

// SaveImage, and SaveFrames through it, writes the same bytes as EncodePNG.
func TestSaveImageAndFramesPreserveAlpha(t *testing.T) {
	data := []uint8{200, 100, 50, 0, 1, 2, 3, 128}
	img := SDImage{Width: 2, Height: 1, Channel: 4, Data: &data[0]}

	dir := t.TempDir()
	path := filepath.Join(dir, "alpha.png")
	if err := SaveImage(&img, path); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}
	if err := SaveFrames([]SDImage{img}, dir); err != nil {
		t.Fatalf("SaveFrames: %v", err)
	}
	for _, p := range []string{path, filepath.Join(dir, "frame_0001.png")} {
		encoded, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if got := decodeNRGBA(t, encoded); !bytes.Equal(got.Pix, data) {
			t.Errorf("%s: Pix = %v, want %v", p, got.Pix, data)
		}
	}
}

// LoadImage drops the alpha of an input image and keeps its straight RGB, as
// upstream's image loader does.
func TestLoadImageKeepsStraightRGB(t *testing.T) {
	data := []uint8{200, 100, 50, 128}
	path := filepath.Join(t.TempDir(), "translucent.png")
	if err := SaveImage(&SDImage{Width: 1, Height: 1, Channel: 4, Data: &data[0]}, path); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}

	loaded, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}
	if got := unsafe.Slice(loaded.Data, 3); got[0] != 200 || got[1] != 100 || got[2] != 50 {
		t.Errorf("RGB = %v, want [200 100 50]", got)
	}
}

// TestSaveLoadImageRoundTrip writes an SDImage to a PNG via SaveImage and reads
// it back via LoadImage, confirming dimensions, channel count, and pixel values
// survive the lossless round trip.
func TestSaveLoadImageRoundTrip(t *testing.T) {
	const w, h = 4, 3
	src, data := newTestImage(w, h, func(x, y uint32) (uint8, uint8, uint8) {
		return uint8(x * 10), uint8(y * 20), uint8(x + y + 1)
	})
	defer func() { _ = data }()

	path := filepath.Join(t.TempDir(), "round.png")
	if err := SaveImage(src, path); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}

	loaded, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}
	if loaded.Width != w || loaded.Height != h {
		t.Fatalf("dimensions = %dx%d, want %dx%d", loaded.Width, loaded.Height, w, h)
	}
	if loaded.Channel != 3 {
		t.Fatalf("channel = %d, want 3", loaded.Channel)
	}
	if loaded.Data == nil {
		t.Fatal("loaded image has nil data")
	}

	// Spot-check pixel (2,1) survived the round trip.
	nrgba, err := toNRGBA(&loaded)
	if err != nil {
		t.Fatalf("toNRGBA on loaded image: %v", err)
	}
	r, g, b, _ := nrgba.At(2, 1).RGBA()
	got := [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
	want := [3]uint8{2 * 10, 1 * 20, 2 + 1 + 1}
	if got != want {
		t.Fatalf("pixel (2,1) = %v, want %v", got, want)
	}
}

// TestLoadImageMissingFileErrors confirms a missing path is an error, not a panic.
func TestLoadImageMissingFileErrors(t *testing.T) {
	if _, err := LoadImage(filepath.Join(t.TempDir(), "nope.png")); err == nil {
		t.Fatal("LoadImage(missing) = nil error, want error")
	}
}

// TestGenerateImageFromPath covers the empty-path and missing-file cases, both
// of which must yield a zero-value SDImage (nil Data) rather than panicking.
func TestGenerateImageFromPath(t *testing.T) {
	if img := GenerateImageFromPath(""); img.Data != nil {
		t.Error("GenerateImageFromPath(\"\") returned non-nil Data")
	}
	if img := GenerateImageFromPath(filepath.Join(t.TempDir(), "missing.png")); img.Data != nil {
		t.Error("GenerateImageFromPath(missing) returned non-nil Data")
	}
}

// TestGenerateImagesFromPaths covers nil/empty/all-invalid (=> nil) and the
// mixed case where at least one path is valid (=> non-nil).
func TestGenerateImagesFromPaths(t *testing.T) {
	if GenerateImagesFromPaths(nil) != nil {
		t.Error("GenerateImagesFromPaths(nil) != nil")
	}
	if GenerateImagesFromPaths([]string{}) != nil {
		t.Error("GenerateImagesFromPaths(empty) != nil")
	}
	if GenerateImagesFromPaths([]string{"", filepath.Join(t.TempDir(), "x.png")}) != nil {
		t.Error("GenerateImagesFromPaths(all invalid) != nil")
	}

	// One valid image among invalid entries → non-nil result.
	src, data := newTestImage(2, 2, func(x, y uint32) (uint8, uint8, uint8) { return 1, 2, 3 })
	defer func() { _ = data }()
	path := filepath.Join(t.TempDir(), "valid.png")
	if err := SaveImage(src, path); err != nil {
		t.Fatalf("SaveImage: %v", err)
	}
	if got := GenerateImagesFromPaths([]string{"", path}); got == nil {
		t.Error("GenerateImagesFromPaths with one valid path returned nil")
	}
}

// TestSaveFrames writes a frame sequence and checks the expected 1-based,
// zero-padded filenames land in the output directory.
func TestSaveFrames(t *testing.T) {
	f1, d1 := newTestImage(2, 2, func(x, y uint32) (uint8, uint8, uint8) { return 0, 0, 0 })
	f2, d2 := newTestImage(2, 2, func(x, y uint32) (uint8, uint8, uint8) { return 255, 255, 255 })
	defer func() { _, _ = d1, d2 }()

	dir := filepath.Join(t.TempDir(), "frames")
	if err := SaveFrames([]SDImage{*f1, *f2}, dir); err != nil {
		t.Fatalf("SaveFrames: %v", err)
	}
	for _, name := range []string{"frame_0001.png", "frame_0002.png"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

// TestGetCpuAVX confirms the detector returns one of the known variant subdir
// names the Windows loader resolves (avx512/avx2/avx/noavx).
func TestGetCpuAVX(t *testing.T) {
	got := GetCpuAVX()
	switch got {
	case "avx512", "avx2", "avx", "noavx":
	default:
		t.Fatalf("GetCpuAVX() = %q, want one of avx512/avx2/avx/noavx", got)
	}
}
