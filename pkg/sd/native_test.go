package sd

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// These tests hold the binding to a real library instead of to the header. They
// need no model and run only when SD_TEST_LIB_DIR names a directory holding one:
//
//	SD_TEST_LIB_DIR=/path/to/build/bin go test ./pkg/sd -run TestNative
func loadNativeLibrary(t *testing.T) {
	t.Helper()
	dir := os.Getenv("SD_TEST_LIB_DIR")
	if dir == "" {
		t.Skip("SD_TEST_LIB_DIR is not set")
	}
	if err := Load(dir); err != nil {
		t.Fatalf("Load(%q): %v", dir, err)
	}
}

// guarded surrounds params with canaries, so a library writing past the end of
// the Go struct, because its C struct is larger, fails the test instead of
// corrupting memory quietly.
type guarded[T any] struct {
	params T
	canary [256]byte
}

func newGuarded[T any]() *guarded[T] {
	g := new(guarded[T])
	for i := range g.canary {
		g.canary[i] = 0xA5
	}
	return g
}

func (g *guarded[T]) check(t *testing.T) {
	t.Helper()
	for i, b := range g.canary {
		if b != 0xA5 {
			t.Fatalf("library wrote %d bytes past the end of the Go struct", i+1)
		}
	}
}

func TestNativeVersion(t *testing.T) {
	loadNativeLibrary(t)
	pin, err := os.ReadFile("../../lib/version.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(pin))
	if got := Version(); !strings.HasPrefix(got, want) {
		t.Errorf("Version() = %q, want prefix %q", got, want)
	}
	if got := Commit(); !strings.HasSuffix(want, got) {
		t.Errorf("Commit() = %q is not the commit of %q", got, want)
	}
}

// The defaults sit at every region of each struct, so reading them through the Go
// fields shows the library and the binding agree on where those fields are.
func TestNativeContextParamsInit(t *testing.T) {
	loadNativeLibrary(t)
	g := newGuarded[SDContextParams]()
	ContextParamsInit(&g.params)
	g.check(t)
	p := &g.params

	if p.NThreads <= 0 || p.WType != SDTypeCount || p.RNGType != CUDARNG || p.SamplerRNGType != RNGTypeCount ||
		p.Prediction != PredictionCount || p.LoraApplyMode != LoraApplyAuto || p.VAEFormat != VAEFormatAuto {
		t.Errorf("scalar defaults are wrong: %+v", p)
	}
	if !p.AutoFit || !p.BatchedCFG || p.DisablePrefetch || p.DisableSegmentedCompute || p.SageAttn {
		t.Errorf("bool defaults are wrong: %+v", p)
	}
	if p.ConditioningCacheSize != 4 || p.LinearScale != 0 || p.AttnScale != 0 {
		t.Errorf("tail defaults are wrong: %+v", p)
	}
}

// sd_ctx_params_to_str reads each field through the C layout and prints it under
// its own name, so a unique value per field catches any field landing elsewhere.
func TestNativeContextParamsRoundTrip(t *testing.T) {
	loadNativeLibrary(t)
	var p SDContextParams
	ContextParamsInit(&p)

	strs := map[string]**uint8{
		"model_path":                      &p.ModelPath,
		"clip_l_path":                     &p.ClipLPath,
		"clip_g_path":                     &p.ClipGPath,
		"clip_vision_path":                &p.ClipVisionPath,
		"t5xxl_path":                      &p.T5XXLPath,
		"llm_path":                        &p.LLMPath,
		"llm_vision_path":                 &p.LLMVisionPath,
		"tokenizer":                       &p.Tokenizer,
		"diffusion_model_path":            &p.DiffusionModelPath,
		"high_noise_diffusion_model_path": &p.HighNoiseDiffusionModelPath,
		"uncond_diffusion_model_path":     &p.UncondDiffusionModelPath,
		"embeddings_connectors_path":      &p.EmbeddingsConnectorsPath,
		"vae_path":                        &p.VAEPath,
		"audio_vae_path":                  &p.AudioVAEPath,
		"audio_encoder_path":              &p.AudioEncoderPath,
		"taesd_path":                      &p.TAESDPath,
		"control_net_path":                &p.ControlNetPath,
		"photo_maker_path":                &p.PhotoMakerPath,
		"pulid_weights_path":              &p.PulidWeightsPath,
		"tensor_type_rules":               &p.TensorTypeRules,
		"max_vram":                        &p.MaxVRAM,
		"backend":                         &p.Backend,
		"params_backend":                  &p.ParamsBackend,
		"split_mode":                      &p.SplitMode,
		"model_args":                      &p.ModelArgs,
		"flash_attn_mode":                 &p.FlashAttnMode,
		"conv_direct_mode":                &p.ConvDirectMode,
	}
	for name, field := range strs {
		*field = CString("v-" + name)
	}
	p.NThreads, p.ConditioningCacheSize = 3, 7
	p.DisablePrefetch, p.DisableSegmentedCompute, p.EagerLoad = true, false, true
	p.AutoFit, p.FlashAttn, p.DiffusionFlashAttn, p.SageAttn = false, true, false, true
	p.LinearScale, p.AttnScale = 1.5, 2.5
	p.Prediction = MingImageFlowPred

	dump := ContextParamsToStr(&p)
	want := []string{
		"n_threads: 3", "conditioning_cache_size: 7", "prediction: ming_image_flow",
		"disable_prefetch: true", "disable_segmented_compute: false", "eager_load: true",
		"auto_fit: false", "flash_attn: true", "diffusion_flash_attn: false", "sage_attn: true",
		"linear_scale: 1.5", "attn_scale: 2.5",
	}
	for name := range strs {
		want = append(want, fmt.Sprintf("%s: v-%s", name, name))
	}
	lines := map[string]bool{}
	for _, line := range strings.Split(dump, "\n") {
		lines[line] = true
	}
	for _, w := range want {
		if !lines[w] {
			t.Errorf("missing line %q in:\n%s", w, dump)
		}
	}
}

func TestNativeImgGenParams(t *testing.T) {
	loadNativeLibrary(t)
	g := newGuarded[SDImgGenParams]()
	ImgGenParamsInit(&g.params)
	g.check(t)
	p := &g.params

	if p.ClipSkip != -1 || p.Width != 512 || p.Height != 512 || p.Seed != -1 || p.BatchCount != 1 ||
		p.Strength != 0.75 || p.QwenImageLayers != 3 {
		t.Errorf("scalar defaults are wrong: %+v", p)
	}
	if p.ControlStrength != 0.9 || p.IPAdapterStrength != 1 || p.PMParams.StyleStrength != 20 || p.PulidParams.IDWeight != 1 {
		t.Errorf("strength defaults are wrong: %+v", p)
	}
	if p.SampleParams.SampleSteps != 20 || p.SampleParams.Guidance.TxtCfg != 7 || p.SampleParams.Scheduler != SchedulerCount ||
		p.SampleParams.SampleMethod != SampleMethodCount {
		t.Errorf("sample defaults are wrong: %+v", p.SampleParams)
	}
	if p.VAETilingParams.TargetOverlap != 0.5 || p.VAETilingParams.TileSizeW != 0 {
		t.Errorf("tiling defaults are wrong: %+v", p.VAETilingParams)
	}
	if p.Hires.UpscaleTileSize != 128 || p.Hires.Upscaler == HiresUpscalerCount {
		t.Errorf("hires defaults are wrong: %+v", p.Hires)
	}
	if p.ImagePreprocess.Rules != nil {
		t.Errorf("ImagePreprocess.Rules = %v, want nil", p.ImagePreprocess.Rules)
	}

	p.Prompt = CString("a prompt")
	p.Seed = 1234567890123
	p.CircularY = true
	p.VAETilingParams.Enabled = true
	p.VAETilingParams.ExtraTilingArgs = CString("extra-tiling")
	p.Hires.ModelPath = CString("hires-model")
	dump := ImgGenParamsToStr(p)
	for _, want := range []string{
		"prompt: a prompt", "seed: 1234567890123", "batch_count: 1", "circular_y: true",
		"VAE tiling: true (temporal=false, extra_tiling_args=extra-tiling)", "model_path=hires-model",
	} {
		if !strings.Contains(dump, want) {
			t.Errorf("missing %q in:\n%s", want, dump)
		}
	}
}

func TestNativeVidGenParamsInit(t *testing.T) {
	loadNativeLibrary(t)
	g := newGuarded[SDVidGenParams]()
	VidGenParamsInit(&g.params)
	g.check(t)
	p := &g.params

	if p.Width != 512 || p.Height != 512 || p.Seed != -1 || p.VideoFrames != 6 || p.Fps != 16 ||
		p.MOEBoundary != 0.875 || p.VaceStrength != 1 || p.Strength != 0.75 {
		t.Errorf("scalar defaults are wrong: %+v", p)
	}
	if p.HighNoiseSampleParams.SampleSteps != -1 || p.SampleParams.SampleSteps != 20 {
		t.Errorf("sample defaults are wrong: %+v / %+v", p.SampleParams, p.HighNoiseSampleParams)
	}
	if p.Hires.UpscaleTileSize != 128 || p.Hires.DenoisingStrength != 0.7 {
		t.Errorf("hires defaults are wrong: %+v", p.Hires)
	}
	// The last field of the struct is set to a static empty string.
	if p.RefImageArgs == nil || *p.RefImageArgs != 0 {
		t.Errorf("RefImageArgs = %v, want an empty string", p.RefImageArgs)
	}
	if p.RefImages != nil || p.RefVideos != nil || p.RefAudios != nil || p.ImagePreprocess.Rules != nil {
		t.Errorf("reference inputs are not empty: %+v", p)
	}
}

func TestNativeEnumNames(t *testing.T) {
	loadNativeLibrary(t)
	cases := []struct{ got, want string }{
		{SampleMethodName(LMSSampleMethod), "lms"},
		{SchedulerName(LLaDAImageScheduler), "llada_image"},
		{PredictionName(SenseNovaU1FlowPred), "sensenova_u1_flow"},
		{PredictionName(H3AVFlowPred), "h3_av_flow"},
		{SDTypeName(SDTypeF8_E4M3), "f8_e4m3"},
		{SDTypeName(SDTypeF8_E5M2), "f8_e5m2"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("name = %q, want %q", c.got, c.want)
		}
	}
	if got := StrToSampleMethod("lms"); got != LMSSampleMethod {
		t.Errorf("StrToSampleMethod(lms) = %d", got)
	}
}

// Failures on the paths a result takes must come back as nil, and the free
// functions must accept nil as the header promises.
func TestNativeResultsWithoutContext(t *testing.T) {
	loadNativeLibrary(t)

	var img SDImgGenParams
	ImgGenParamsInit(&img)
	if (&SDContext{}).GenerateImage(&img) != nil {
		t.Error("GenerateImage without a context != nil")
	}
	var vid SDVidGenParams
	VidGenParamsInit(&vid)
	if (&SDContext{}).GenerateVideo(&vid) != nil {
		t.Error("GenerateVideo without a context != nil")
	}
	freeSDImageGenerationResult(nil)
	freeSDVideoGenerationResult(nil)

	if got := (&SDContext{}).ModelVersionName(); got != "Unknown" {
		t.Errorf("ModelVersionName() = %q, want Unknown", got)
	}
	if got := GetPreviewInfo(); got != (SDPreviewInfo{}) {
		t.Errorf("GetPreviewInfo() = %+v, want zero", got)
	}
}

// SD_LOG_ERROR is 4 since master-1031. A context that cannot find its model logs
// an error, so the level reaching Go shows the shifted constants are the real ones.
func TestNativeLogLevelShift(t *testing.T) {
	loadNativeLibrary(t)

	levels := map[SDLogLevel]int{}
	SetLogCallback(func(level SDLogLevelType, text string, _ interface{}) {
		levels[SDLogLevel(level)]++
	}, nil)
	defer SetLogCallback(nil, nil)

	var p SDContextParams
	ContextParamsInit(&p)
	p.ModelPath = CString("/nonexistent/model.gguf")
	ctx := NewContext(&p)
	defer ctx.Free()

	if levels[SDLogError] == 0 {
		t.Errorf("no error reached the log callback at level %d: %v", SDLogError, levels)
	}
	for level := range levels {
		if level < SDLogDebug || level > SDLogError {
			t.Errorf("log level %d is out of range", level)
		}
	}
}

func TestNativeDeviceMemory(t *testing.T) {
	loadNativeLibrary(t)
	if !devMemAvailable {
		t.Fatal("the sd_backend_dev_* wrappers did not bind")
	}
	if devices, ok := GpuDevices(); !ok || len(devices) == 0 {
		t.Errorf("GpuDevices() = %+v, %v", devices, ok)
	}
}

// preprocess_canny takes sd_image_t by value, see byValueImageArgsSupported.
func TestNativePreprocessCanny(t *testing.T) {
	loadNativeLibrary(t)
	if !byValueImageArgsSupported {
		t.Skip("by-value sd_image_t is not supported on this target")
	}
	const size = 16
	data := make([]uint8, size*size*3)
	for i := range data {
		if (i/3)%size >= size/2 {
			data[i] = 255 // a vertical edge
		}
	}
	before := append([]uint8(nil), data...)
	ok := PreprocessCanny(SDImage{Width: size, Height: size, Channel: 3, Data: &data[0]}, 0.08, 0.08, 0.8, 1, false)
	if !ok {
		t.Fatal("PreprocessCanny returned false")
	}
	if string(data) == string(before) {
		t.Error("PreprocessCanny left the image unchanged")
	}
}
