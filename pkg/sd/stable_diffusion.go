package sd

import (
	"runtime"
	"unsafe"
)

// Define enum types

type RngType int32

const (
	DefaultRNG RngType = iota
	CUDARNG
	CPURNG
	RNGTypeCount
)

type SampleMethod int32

const (
	EulerSampleMethod SampleMethod = iota
	EulerASampleMethod
	HeunSampleMethod
	DPM2SampleMethod
	DPMPP2SASampleMethod
	DPMPP2MSampleMethod
	DPMPP2Mv2SampleMethod
	IPNDMSampleMethod
	IPNDMSampleMethodV
	LCMSampleMethod
	DDIMTrailingSampleMethod
	TCDSampleMethod
	ResMultistepSampleMethod
	Res2SSampleMethod
	ERSDESampleMethod
	EulerCFGPPSampleMethod
	EulerACFGPPSampleMethod
	EulerGESampleMethod
	DPMPP2MSDESampleMethod
	DPMPP2MSDEBTSampleMethod
	LMSSampleMethod
	SampleMethodCount
)

type Scheduler int32

const (
	DiscreteScheduler Scheduler = iota
	KarrasScheduler
	ExponentialScheduler
	AYSScheduler
	GITScheduler
	SGMUniformScheduler
	SimpleScheduler
	SmoothstepScheduler
	KLOptimalScheduler
	LCMScheduler
	BongTangentScheduler
	LTX2Scheduler
	LogitNormalScheduler
	Flux2Scheduler
	FluxScheduler
	BetaScheduler
	LLaDAImageScheduler
	SchedulerCount
)

type Prediction int32

const (
	EPSPred Prediction = iota
	VPred
	EDMVPred
	FlowPred
	FluxFlowPred
	// SefiFlowPred occupies the enum slot that FLUX2_FLOW_PRED held before the
	// master-802 resync: upstream replaced FLUX2_FLOW_PRED with SEFI_FLOW_PRED
	// (same integer value, 5) and appended MINIT2I_FLOW_PRED.
	SefiFlowPred
	Minit2iFlowPred
	SenseNovaU1FlowPred
	Flux3FlowPred
	H3AVFlowPred
	MingImageFlowPred
	PredictionCount
)

type SDType int32

const (
	SDTypeF32 SDType = iota
	SDTypeF16
	SDTypeQ4_0
	SDTypeQ4_1
	// SDTypeQ4_2 = 4, support has been removed
	// SDTypeQ4_3 = 5, support has been removed
	SDTypeQ5_0    = 6
	SDTypeQ5_1    = 7
	SDTypeQ8_0    = 8
	SDTypeQ8_1    = 9
	SDTypeQ2_K    = 10
	SDTypeQ3_K    = 11
	SDTypeQ4_K    = 12
	SDTypeQ5_K    = 13
	SDTypeQ6_K    = 14
	SDTypeQ8_K    = 15
	SDTypeIQ2_XXS = 16
	SDTypeIQ2_XS  = 17
	SDTypeIQ3_XXS = 18
	SDTypeIQ1_S   = 19
	SDTypeIQ4_NL  = 20
	SDTypeIQ3_S   = 21
	SDTypeIQ2_S   = 22
	SDTypeIQ4_XS  = 23
	SDTypeI8      = 24
	SDTypeI16     = 25
	SDTypeI32     = 26
	SDTypeI64     = 27
	SDTypeF64     = 28
	SDTypeIQ1_M   = 29
	SDTypeBF16    = 30
	// SDTypeQ4_0_4_4 = 31, support has been removed from gguf files
	// SDTypeQ4_0_4_8 = 32,
	// SDTypeQ4_0_8_8 = 33,
	SDTypeTQ1_0 = 34
	SDTypeTQ2_0 = 35
	// SDTypeIQ4_NL_4_4 = 36,
	// SDTypeIQ4_NL_4_8 = 37,
	// SDTypeIQ4_NL_8_8 = 38,
	SDTypeMXFP4   = 39
	SDTypeNVFP4   = 40 // NVFP4 (4 blocks, E4M3 scale)
	SDTypeQ1_0    = 41
	SDTypeQ2_0    = 42
	SDTypeF8_E4M3 = 43
	SDTypeF8_E5M2 = 44
	SDTypeCount   = 45
)

// SDLogLevel mirrors enum sd_log_level_t. Compare levels through these constants
// only: SD_LOG_VERBOSE took slot 1 at the master-1031 resync, which moved
// INFO/WARN/ERROR from 1/2/3 to 2/3/4.
type SDLogLevel int32

const (
	SDLogDebug SDLogLevel = iota
	SDLogVerbose
	SDLogInfo
	SDLogWarn
	SDLogError
)

type Preview int32

const (
	PreviewNone Preview = iota
	PreviewProj
	PreviewTAE
	PreviewVAE
	PreviewCount
)

type LoraApplyMode int32

const (
	LoraApplyAuto LoraApplyMode = iota
	LoraApplyImmediately
	LoraApplyAtRuntime
	LoraApplyModeCount
)

type SDCacheMode int32

const (
	SDCacheDisabled SDCacheMode = iota
	SDCacheEasycache
	SDCacheUcache
	SDCacheDbcache
	SDCacheTaylorseer
	SDCacheCacheDit
	SDCacheSpectrum
)

type SDVAEFormat int32

const (
	VAEFormatAuto  SDVAEFormat = -1
	FluxVAEFormat  SDVAEFormat = 0
	SD3VAEFormat   SDVAEFormat = 1
	Flux2VAEFormat SDVAEFormat = 2
	WanVAEFormat   SDVAEFormat = 3
	VAEFormatCount SDVAEFormat = 4
)

// SDCancelMode mirrors enum sd_cancel_mode_t (added upstream in the master-802
// resync) and selects how sd_cancel_generation interrupts an in-flight
// generation.
type SDCancelMode int32

const (
	// CancelAll stops the current generation as soon as possible.
	CancelAll SDCancelMode = iota
	// CancelNewLatents finishes the current image sample, then skips
	// additional batch latents and returns the completed images.
	CancelNewLatents
	// CancelReset clears a pending cancellation request.
	CancelReset
)

type SDHiresUpscaler int32

const (
	HiresUpscalerNone SDHiresUpscaler = iota
	HiresUpscalerLatent
	HiresUpscalerLatentNearest
	HiresUpscalerLatentNearestExact
	HiresUpscalerLatentAntialiased
	HiresUpscalerLatentBicubic
	HiresUpscalerLatentBicubicAntialiased
	HiresUpscalerLanczos
	HiresUpscalerNearest
	HiresUpscalerModel
	HiresUpscalerCount
)

// Define structs

// SDTilingParams mirrors sd_tiling_params_t. The master-1031 resync renamed the
// X/Y fields to W/H without changing the layout.
type SDTilingParams struct {
	Enabled        bool
	TemporalTiling bool
	// TileSizeW and TileSizeH are the spatial tile dimensions in image pixels,
	// for both encode and decode; 0 uses 256.
	TileSizeW     int32
	TileSizeH     int32
	TargetOverlap float32
	// RelSizeW and RelSizeH override the tile size when positive: <= 1 is a
	// fraction of the dimension, > 1 a target tile count.
	RelSizeW        float32
	RelSizeH        float32
	ExtraTilingArgs *uint8
}

type SDEmbedding struct {
	Name *uint8
	Path *uint8
}

// SDContextParams mirrors sd_ctx_params_t at upstream master-1031-c6dfa9d.
// Field order, types and implicit padding MUST match the C struct exactly —
// this layout is what crosses the FFI boundary, and TestStructLayoutMatchesHeader
// pins it. Notable drift from the previous pin (master-802-e92e86f):
//   - added: AudioEncoderPath (after AudioVAEPath); DisableSegmentedCompute,
//     BatchedCFG, LinearScale, AttnScale, Tokenizer, SageAttn,
//     ConditioningCacheSize, FlashAttnMode, ConvDirectMode (after ModelArgs)
//   - replaced: StreamLayers became DisablePrefetch in the same slot, with the
//     opposite sense. Segmented streaming is automatic now and prefetch is on
//     unless disabled, so there is nothing left to opt into.
//   - library defaults changed: AutoFit is true (was false), BatchedCFG is true
//     and ConditioningCacheSize is 4, so set them after ContextParamsInit, not
//     from a zero value.
type SDContextParams struct {
	ModelPath                   *uint8
	ClipLPath                   *uint8
	ClipGPath                   *uint8
	ClipVisionPath              *uint8
	T5XXLPath                   *uint8
	LLMPath                     *uint8
	LLMVisionPath               *uint8
	DiffusionModelPath          *uint8
	HighNoiseDiffusionModelPath *uint8
	UncondDiffusionModelPath    *uint8
	EmbeddingsConnectorsPath    *uint8
	VAEPath                     *uint8
	AudioVAEPath                *uint8
	AudioEncoderPath            *uint8
	TAESDPath                   *uint8
	ControlNetPath              *uint8
	IPAdapterPath               *uint8
	MotionModulePath            *uint8
	Embeddings                  *SDEmbedding
	EmbeddingCount              uint32
	PhotoMakerPath              *uint8
	PulidWeightsPath            *uint8
	TensorTypeRules             *uint8
	NThreads                    int32
	WType                       SDType
	RNGType                     RngType
	SamplerRNGType              RngType
	Prediction                  Prediction
	LoraApplyMode               LoraApplyMode
	EnableMmap                  bool
	FlashAttn                   bool
	DiffusionFlashAttn          bool
	TAEPreviewOnly              bool
	DiffusionConvDirect         bool
	VAEConvDirect               bool
	ForceSDXLVAConvScale        bool
	VAEFormat                   SDVAEFormat
	MaxVRAM                     *uint8
	DisablePrefetch             bool // disable asynchronous next-segment weight prefetch
	EagerLoad                   bool
	Backend                     *uint8
	ParamsBackend               *uint8
	SplitMode                   *uint8
	AutoFit                     bool
	RPCServers                  *uint8
	ModelArgs                   *uint8
	DisableSegmentedCompute     bool    // force monolithic graph execution
	BatchedCFG                  bool    // run the CFG branches in one batched forward when supported
	LinearScale                 float32 // override linear input scaling; 0 keeps the model default
	AttnScale                   float32 // override flash-attention K/V scaling; 0 keeps the model default
	Tokenizer                   *uint8  // tokenizer.json path or main=FILE,clip-l=FILE,clip-g=FILE assignments; required for PiD and Lens
	SageAttn                    bool
	ConditioningCacheSize       int32  // cached conditioning entries per context; 0 disables caching
	FlashAttnMode               *uint8 // on/off/auto or module assignments; nil or empty honours the legacy flags
	ConvDirectMode              *uint8 // on/off/auto or diffusion/vae/controlnet assignments; nil or empty honours the legacy flags
}

type SDImage struct {
	Width   uint32
	Height  uint32
	Channel uint32
	Data    *uint8
}

// SDImageResult mirrors sd_image_result_t, one item of a generation result.
// GenerationInfo is owned UTF-8 JSON using the sdcpp.generation/v1 schema.
type SDImageResult struct {
	Image          SDImage
	GenerationInfo *uint8
}

// SDImagePreprocessParams mirrors sd_image_preprocess_params_t. Rules is a list
// of semicolon-separated target=...,key=value rules; nil preserves the defaults.
type SDImagePreprocessParams struct {
	Rules *uint8
}

// SDRefVideo mirrors sd_ref_video_t, a reference video for video generation.
type SDRefVideo struct {
	Frames     *SDImage
	FrameCount int32
	Fps        int32
	Audio      SDAudio
}

// SDPreviewInfo mirrors sd_preview_info_t. SamplePass counts sampling passes from
// 1 since the last SetPreviewCallback; TotalSteps is the step count of that pass.
// Both are zero before sampling starts.
type SDPreviewInfo struct {
	SamplePass int32
	TotalSteps int32
}

// cImageGenerationResult mirrors sd_image_generation_result_t. Items is an array
// of Count SDImageResult, not of SDImage. Callers use SDImageGenerationResult.
type cImageGenerationResult struct {
	Items *SDImageResult
	Count int32
}

// cVideoGenerationResult mirrors sd_video_generation_result_t. Callers use
// SDVideoGenerationResult.
type cVideoGenerationResult struct {
	Frames         *SDImage
	FrameCount     int32
	Audio          *SDAudio
	Fps            int32
	GenerationInfo *uint8
}

type SDSLGParams struct {
	Layers     *int32
	LayerCount uintptr
	LayerStart float32
	LayerEnd   float32
	Scale      float32
}

type SDGuidanceParams struct {
	TxtCfg            float32
	ImgCfg            float32
	DistilledGuidance float32
	SLG               SDSLGParams
}

type SDSampleParams struct {
	Guidance          SDGuidanceParams
	Scheduler         Scheduler
	SampleMethod      SampleMethod
	SampleSteps       int32
	Eta               float32
	ShiftedTimestep   int32
	CustomSigmas      *float32
	CustomSigmasCount int32
	FlowShift         float32
	ExtraSampleArgs   *uint8
}

type SDPMParams struct {
	IDImages      *SDImage
	IDImagesCount int32
	IDEmbedPath   *uint8
	StyleStrength float32
}

// SDPulidParams mirrors sd_pulid_params_t (added upstream in the master-802
// resync).
type SDPulidParams struct {
	IDEmbeddingPath *uint8
	IDWeight        float32
}

type SDCacheParams struct {
	Mode                     SDCacheMode
	ReuseThreshold           float32
	StartPercent             float32
	EndPercent               float32
	ErrorDecayRate           float32
	UseRelativeThreshold     bool
	ResetErrorOnCompute      bool
	FnComputeBlocks          int32
	BnComputeBlocks          int32
	ResidualDiffThreshold    float32
	MaxWarmupSteps           int32
	MaxCachedSteps           int32
	MaxContinuousCachedSteps int32
	TaylorseerNDerivatives   int32
	TaylorseerSkipInterval   int32
	ScmMask                  *uint8
	ScmPolicyDynamic         bool
	SpectrumW                float32
	SpectrumM                int32
	SpectrumLam              float32
	SpectrumWindowSize       int32
	SpectrumFlexWindow       float32
	SpectrumWarmupSteps      int32
	SpectrumStopPercent      float32
}

type SDLora struct {
	IsHighNoise bool
	Multiplier  float32
	Path        *uint8
}

type SDHiresParams struct {
	Enabled           bool
	Upscaler          SDHiresUpscaler
	ModelPath         *uint8
	Scale             float32
	TargetWidth       int32
	TargetHeight      int32
	Steps             int32
	DenoisingStrength float32
	UpscaleTileSize   int32
	CustomSigmas      *float32
	CustomSigmasCount int32
}

// SDAudio mirrors sd_audio_t. It is produced by audio-capable video models (e.g.
// LTX2) and accepted as a reference by generate_video.
type SDAudio struct {
	SampleRate  uint32
	Channels    uint32
	SampleCount uint64
	Data        *float32
}

// SDImgGenParams mirrors sd_img_gen_params_t at upstream master-1031-c6dfa9d.
// Drift from the previous pin: ImagePreprocess was appended.
type SDImgGenParams struct {
	Loras             *SDLora
	LoraCount         uint32
	Prompt            *uint8
	NegativePrompt    *uint8
	ClipSkip          int32
	InitImage         SDImage
	RefImages         *SDImage
	RefImagesCount    int32
	RefImageArgs      *uint8
	MaskImage         SDImage
	Width             int32
	Height            int32
	SampleParams      SDSampleParams
	Strength          float32
	Seed              int64
	BatchCount        int32
	ControlImage      SDImage
	ControlStrength   float32
	IPAdapterImage    SDImage
	IPAdapterStrength float32
	PMParams          SDPMParams
	PulidParams       SDPulidParams
	VAETilingParams   SDTilingParams
	Cache             SDCacheParams
	Hires             SDHiresParams
	QwenImageLayers   int32
	CircularX         bool
	CircularY         bool
	ImagePreprocess   SDImagePreprocessParams
}

// SDVidGenParams mirrors sd_vid_gen_params_t at upstream master-1031-c6dfa9d.
// Drift from the previous pin: RefImages, RefImagesCount, RefVideos,
// RefVideosCount, RefAudios and RefAudiosCount were inserted before
// ControlFrames, and ImagePreprocess and RefImageArgs were appended.
type SDVidGenParams struct {
	Loras                 *SDLora
	LoraCount             uint32
	Prompt                *uint8
	NegativePrompt        *uint8
	ClipSkip              int32
	InitImage             SDImage
	EndImage              SDImage
	RefImages             *SDImage
	RefImagesCount        int32
	RefVideos             *SDRefVideo
	RefVideosCount        int32
	RefAudios             *SDAudio
	RefAudiosCount        int32
	ControlFrames         *SDImage
	ControlFramesSize     int32
	Width                 int32
	Height                int32
	SampleParams          SDSampleParams
	HighNoiseSampleParams SDSampleParams
	MOEBoundary           float32
	Strength              float32
	Seed                  int64
	VideoFrames           int32
	Fps                   int32
	VaceStrength          float32
	VAETilingParams       SDTilingParams
	Cache                 SDCacheParams
	Hires                 SDHiresParams
	CircularX             bool
	CircularY             bool
	ImagePreprocess       SDImagePreprocessParams
	RefImageArgs          *uint8
}

// Define context types
type SDContext struct {
	ptr unsafe.Pointer
}

type UpscalerContext struct {
	ptr unsafe.Pointer
}

// Define callback function types.
//
// These mirror the C signatures, which are `void (*)(...)`. They are
// nonetheless declared to return uintptr because purego turns a Go func
// passed as a C callback into a machine callback via purego.NewCallback,
// and on Windows purego delegates to syscall.NewCallback ->
// runtime.compileCallback, which REQUIRES the callback return exactly one
// uintptr-sized result. A void Go callback panics there with
// "compileCallback: expected function with one uintptr-sized result"
// (the symptom seen during stable-diffusion image generation on Windows).
// The C caller declares these as void and simply ignores the returned
// register, so always returning 0 is safe on every platform. (This is the
// same shape used for ggml's log callback shim elsewhere in the ecosystem.)
type SDLogCallback func(level SDLogLevel, text *uint8, data unsafe.Pointer) uintptr

// SDProgressCallback NOTE (Windows): this callback takes a float32 argument.
// purego's callback support on Windows is syscall.NewCallback, which supports
// neither float arguments nor float returns — so this callback cannot be
// wrapped on Windows. SetProgressCallback handles that gracefully (it skips
// registration on Windows rather than panicking), so progress reporting is
// simply unavailable there. The log and preview callbacks have no float in
// their signatures and work on every platform. Lifting this would require
// removing the float from the C-ABI signature.
type SDProgressCallback func(step int32, steps int32, time float32, data unsafe.Pointer) uintptr
type SDPreviewCallback func(step int32, frameCount int32, frames *SDImage, isNoisy bool, data unsafe.Pointer) uintptr

// Dynamic library function declarations
var (
	sdSetLogCallback         func(cb SDLogCallback, data unsafe.Pointer)
	sdSetProgressCallback    func(cb SDProgressCallback, data unsafe.Pointer)
	sdSetPreviewCallback     func(cb SDPreviewCallback, mode Preview, interval int32, denoised bool, noisy bool, data unsafe.Pointer)
	sdGetNumPhysicalCores    func() int32
	sdGetSystemInfo          func() *uint8
	sdTypeName               func(typ SDType) *uint8
	strToSDType              func(str *uint8) SDType
	sdRngTypeName            func(rngType RngType) *uint8
	strToRngType             func(str *uint8) RngType
	sdSampleMethodName       func(method SampleMethod) *uint8
	strToSampleMethod        func(str *uint8) SampleMethod
	sdSchedulerName          func(scheduler Scheduler) *uint8
	strToScheduler           func(str *uint8) Scheduler
	sdPredictionName         func(prediction Prediction) *uint8
	strToPrediction          func(str *uint8) Prediction
	sdPreviewName            func(preview Preview) *uint8
	strToPreview             func(str *uint8) Preview
	sdLoraApplyModeName      func(mode LoraApplyMode) *uint8
	strToLoraApplyMode       func(str *uint8) LoraApplyMode
	sdCacheParamsInit        func(params *SDCacheParams)
	sdContextParamsInit      func(params *SDContextParams)
	sdContextParamsToStr     func(params *SDContextParams) *uint8
	newSDContext             func(params *SDContextParams) unsafe.Pointer
	freeSDContext            func(ctx unsafe.Pointer)
	sdSampleParamsInit       func(params *SDSampleParams)
	sdSampleParamsToStr      func(params *SDSampleParams) *uint8
	sdGetDefaultSampleMethod func(ctx unsafe.Pointer) SampleMethod
	sdGetDefaultScheduler    func(ctx unsafe.Pointer, sampleMethod SampleMethod) Scheduler
	sdImgGenParamsInit       func(params *SDImgGenParams)
	sdImgGenParamsToStr      func(params *SDImgGenParams) *uint8
	generateImage            func(ctx unsafe.Pointer, params *SDImgGenParams, imagesOut **SDImage, numImagesOut *int32) bool
	sdCancelGeneration       func(ctx unsafe.Pointer, mode SDCancelMode)
	sdVidGenParamsInit       func(params *SDVidGenParams)
	generateVideo            func(ctx unsafe.Pointer, params *SDVidGenParams, framesOut **SDImage, numFramesOut *int32, audioOut **SDAudio) bool
	newUpscalerContext       func(esrganPath *uint8, direct bool, nThreads int32, tileSize int32, backend *uint8, paramsBackend *uint8) unsafe.Pointer
	freeUpscalerContext      func(ctx unsafe.Pointer)
	upscale                  func(ctx unsafe.Pointer, inputImage *SDImage, upscaleFactor uint32, imagesOut **SDImage, numImagesOut *int32) bool
	getUpscaleFactor         func(ctx unsafe.Pointer) int32
	convert                  func(inputPath *uint8, vaePath *uint8, outputPath *uint8, outputType SDType, tensorTypeRules *uint8, convertName bool) bool
	preprocessCanny          func(image *SDImage, highThreshold float32, lowThreshold float32, weak float32, strong float32, inverse bool) bool
	sdCommit                 func() *uint8
	sdVersion                func() *uint8

	sdCtxSupportsImageGeneration func(ctx unsafe.Pointer) bool
	sdCtxSupportsVideoGeneration func(ctx unsafe.Pointer) bool
	sdHiresUpscalerName          func(upscaler SDHiresUpscaler) *uint8
	strToSDHiresUpscaler         func(str *uint8) SDHiresUpscaler
	sdHiresParamsInit            func(params *SDHiresParams)
	freeSDAudio                  func(audio *SDAudio)
	freeSDImages                 func(images *SDImage, numImages int32)
)

// The shared library is loaded lazily via Load (see load.go); importing this
// package performs no filesystem access and no dlopen.

// Wrapper functions
type SDLogLevelType int32

type SDLogCallbackType func(level SDLogLevelType, text string, data interface{})

// SetLogCallback sets log callback
func SetLogCallback(cb SDLogCallbackType, data interface{}) {
	if cb == nil {
		sdSetLogCallback(nil, nil)
		return
	}

	// Create a closure to convert Go callback to C callback. It returns
	// uintptr (always 0) so purego.NewCallback accepts it on Windows; see
	// the SDLogCallback type comment.
	cCallback := func(level SDLogLevel, text *uint8, cData unsafe.Pointer) uintptr {
		cb(SDLogLevelType(level), CGoString(text), data)
		return 0
	}

	sdSetLogCallback(cCallback, nil)
}

// SetProgressCallback sets progress callback.
//
// Windows limitation (handled gracefully): the C progress callback takes a
// `float` argument, and Windows' syscall.NewCallback — which purego uses to
// turn a Go func into a C callback there — supports neither float arguments
// nor float returns. Registering a progress callback on Windows would panic
// ("float arguments not supported"). Rather than leave that as a latent crash,
// SetProgressCallback is a no-op for a non-nil cb on Windows: progress
// reporting is simply unavailable there. Clearing (cb == nil) still works on
// every platform. The log and preview callbacks have no float in their
// signatures and work everywhere.
func SetProgressCallback(cb func(step int, steps int, time float32, data interface{}), data interface{}) {
	if cb == nil {
		sdSetProgressCallback(nil, nil)
		return
	}
	if runtime.GOOS == "windows" {
		// Can't wrap a float-arg callback on Windows; skip registration
		// instead of panicking. (See the doc comment above.)
		return
	}

	cCallback := func(step int32, steps int32, time float32, cData unsafe.Pointer) uintptr {
		cb(int(step), int(steps), time, data)
		return 0
	}

	sdSetProgressCallback(cCallback, nil)
}

// SetPreviewCallback sets preview callback
func SetPreviewCallback(cb func(step int, frameCount int, frames []SDImage, isNoisy bool, data interface{}), mode Preview, interval int, denoised bool, noisy bool, data interface{}) {
	if cb == nil {
		sdSetPreviewCallback(nil, mode, int32(interval), denoised, noisy, nil)
		return
	}

	cCallback := func(step int32, frameCount int32, cFrames *SDImage, isNoisy bool, cData unsafe.Pointer) uintptr {
		// Convert C pointer to Go slice
		frames := make([]SDImage, frameCount)
		for i := range frames {
			frames[i] = *(*SDImage)(unsafe.Add(unsafe.Pointer(cFrames), uintptr(i)*unsafe.Sizeof(SDImage{})))
		}
		cb(int(step), int(frameCount), frames, isNoisy, data)
		return 0
	}

	sdSetPreviewCallback(cCallback, mode, int32(interval), denoised, noisy, nil)
}

// GetNumPhysicalCores gets the number of physical cores
func GetNumPhysicalCores() int {
	return int(sdGetNumPhysicalCores())
}

// GetSystemInfo gets system information
func GetSystemInfo() string {
	return CGoString(sdGetSystemInfo())
}

// SDTypeName gets SD type name
func SDTypeName(typ SDType) string {
	return CGoString(sdTypeName(typ))
}

// StrToSDType converts string to SD type
func StrToSDType(str string) SDType {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToSDType(cStr)
}

// RNGTypeName gets RNG type name
func RNGTypeName(rngType RngType) string {
	return CGoString(sdRngTypeName(rngType))
}

// StrToRNGType converts string to RNG type
func StrToRNGType(str string) RngType {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToRngType(cStr)
}

// SampleMethodName gets sample method name
func SampleMethodName(method SampleMethod) string {
	return CGoString(sdSampleMethodName(method))
}

// StrToSampleMethod converts string to sample method
func StrToSampleMethod(str string) SampleMethod {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToSampleMethod(cStr)
}

// SchedulerName gets scheduler name
func SchedulerName(scheduler Scheduler) string {
	return CGoString(sdSchedulerName(scheduler))
}

// StrToScheduler converts string to scheduler
func StrToScheduler(str string) Scheduler {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToScheduler(cStr)
}

// PredictionName gets prediction type name
func PredictionName(prediction Prediction) string {
	return CGoString(sdPredictionName(prediction))
}

// StrToPrediction converts string to prediction type
func StrToPrediction(str string) Prediction {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToPrediction(cStr)
}

// PreviewName gets preview type name
func PreviewName(preview Preview) string {
	return CGoString(sdPreviewName(preview))
}

// StrToPreview converts string to preview type
func StrToPreview(str string) Preview {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToPreview(cStr)
}

// LoraApplyModeName gets LoRA apply mode name
func LoraApplyModeName(mode LoraApplyMode) string {
	return CGoString(sdLoraApplyModeName(mode))
}

// StrToLoraApplyMode converts string to LoRA apply mode
func StrToLoraApplyMode(str string) LoraApplyMode {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToLoraApplyMode(cStr)
}

// CacheParamsInit initializes cache parameters
func CacheParamsInit(params *SDCacheParams) {
	sdCacheParamsInit(params)
}

// ContextParamsInit initializes context parameters
func ContextParamsInit(params *SDContextParams) {
	sdContextParamsInit(params)
}

// ContextParamsToStr converts context parameters to string
func ContextParamsToStr(params *SDContextParams) string {
	return CGoString(sdContextParamsToStr(params))
}

// NewContext creates a new context
func NewContext(params *SDContextParams) *SDContext {
	ptr := newSDContext(params)
	return &SDContext{ptr: ptr}
}

// FreeContext frees context
func (ctx *SDContext) Free() {
	if ctx.ptr != nil {
		freeSDContext(ctx.ptr)
		ctx.ptr = nil
	}
}

// SampleParamsInit initializes sample parameters
func SampleParamsInit(params *SDSampleParams) {
	sdSampleParamsInit(params)
}

// SampleParamsToStr converts sample parameters to string
func SampleParamsToStr(params *SDSampleParams) string {
	return CGoString(sdSampleParamsToStr(params))
}

// GetDefaultSampleMethod gets default sample method
func (ctx *SDContext) GetDefaultSampleMethod() SampleMethod {
	return sdGetDefaultSampleMethod(ctx.ptr)
}

// GetDefaultScheduler gets default scheduler
func (ctx *SDContext) GetDefaultScheduler(sampleMethod SampleMethod) Scheduler {
	return sdGetDefaultScheduler(ctx.ptr, sampleMethod)
}

// ImgGenParamsInit initializes image generation parameters
func ImgGenParamsInit(params *SDImgGenParams) {
	sdImgGenParamsInit(params)
}

// ImgGenParamsToStr converts image generation parameters to string
func ImgGenParamsToStr(params *SDImgGenParams) string {
	return CGoString(sdImgGenParamsToStr(params))
}

// GenerateImage generates images. As of the master-802 upstream resync,
// generate_image returns a success bool and yields a malloc'd array of images
// through out-parameters (instead of returning the array directly), so this
// wrapper returns the native array pointer plus the number of images in it.
// On failure it returns (nil, 0). The caller owns the returned array and must
// release it with FreeImages(images, count) when done.
func (ctx *SDContext) GenerateImage(params *SDImgGenParams) (*SDImage, int) {
	var (
		imagesPtr *SDImage
		numImages int32
	)
	ok := generateImage(ctx.ptr, params, &imagesPtr, &numImages)
	if !ok || imagesPtr == nil {
		return nil, 0
	}
	return imagesPtr, int(numImages)
}

// CancelGeneration requests cancellation of the in-flight generation on this
// context (or clears a pending request, with CancelReset). It is safe to call
// from another goroutine while GenerateImage/GenerateVideo is running — that
// is its purpose.
func (ctx *SDContext) CancelGeneration(mode SDCancelMode) {
	sdCancelGeneration(ctx.ptr, mode)
}

// VidGenParamsInit initializes video generation parameters
func VidGenParamsInit(params *SDVidGenParams) {
	sdVidGenParamsInit(params)
}

// GenerateVideo generates video. As of the upstream resync, generate_video
// returns a success bool and also yields an audio buffer for audio-capable
// models (e.g. LTX2). This binding does not surface audio, so the buffer is
// freed immediately; the public ([]SDImage, int) signature is unchanged so
// existing callers keep working.
func (ctx *SDContext) GenerateVideo(params *SDVidGenParams) ([]SDImage, int) {
	var (
		framesPtr *SDImage
		numFrames int32
		audioPtr  *SDAudio
	)

	ok := generateVideo(ctx.ptr, params, &framesPtr, &numFrames, &audioPtr)
	if audioPtr != nil && freeSDAudio != nil {
		freeSDAudio(audioPtr)
	}
	if !ok || framesPtr == nil {
		return nil, 0
	}

	frames := make([]SDImage, numFrames)
	for i := range frames {
		frames[i] = *(*SDImage)(unsafe.Add(unsafe.Pointer(framesPtr), uintptr(i)*unsafe.Sizeof(SDImage{})))
	}

	return frames, int(numFrames)
}

// NewUpscalerContext creates a new upscaler context. backend and paramsBackend
// select the runtime backend; empty strings fall back to the library default.
// (Upstream removed the offload_params_to_cpu argument in the master-802
// resync; CPU offload is now expressed as a params-backend assignment spec,
// e.g. paramsBackend = "*=cpu".)
func NewUpscalerContext(esrganPath string, direct bool, nThreads int, tileSize int, backend string, paramsBackend string) *UpscalerContext {
	cPath := CString(esrganPath)
	cBackend := CString(backend)
	cParamsBackend := CString(paramsBackend)
	defer FreeCString(cPath)
	defer FreeCString(cBackend)
	defer FreeCString(cParamsBackend)

	ptr := newUpscalerContext(cPath, direct, int32(nThreads), int32(tileSize), cBackend, cParamsBackend)
	return &UpscalerContext{ptr: ptr}
}

// FreeUpscalerContext frees upscaler context
func (ctx *UpscalerContext) Free() {
	if ctx.ptr != nil {
		freeUpscalerContext(ctx.ptr)
		ctx.ptr = nil
	}
}

// Upscale upscales an image. As of the master-802 upstream resync, upscale
// returns a success bool and yields a malloc'd image array through
// out-parameters, so this wrapper returns the native array pointer plus the
// number of images in it ((nil, 0) on failure). The caller owns the returned
// array and must release it with FreeImages(images, count) when done.
func (ctx *UpscalerContext) Upscale(inputImage SDImage, upscaleFactor uint32) (*SDImage, int) {
	var (
		imagesPtr *SDImage
		numImages int32
	)
	ok := upscale(ctx.ptr, &inputImage, upscaleFactor, &imagesPtr, &numImages)
	if !ok || imagesPtr == nil {
		return nil, 0
	}
	return imagesPtr, int(numImages)
}

// GetUpscaleFactor gets upscale factor
func (ctx *UpscalerContext) GetUpscaleFactor() int {
	return int(getUpscaleFactor(ctx.ptr))
}

// Convert converts model
func Convert(inputPath, vaePath, outputPath string, outputType SDType, tensorTypeRules string, convertName bool) bool {
	cInputPath := CString(inputPath)
	cVaePath := CString(vaePath)
	cOutputPath := CString(outputPath)
	cTensorTypeRules := CString(tensorTypeRules)

	defer func() {
		FreeCString(cInputPath)
		FreeCString(cVaePath)
		FreeCString(cOutputPath)
		FreeCString(cTensorTypeRules)
	}()

	return convert(cInputPath, cVaePath, cOutputPath, outputType, cTensorTypeRules, convertName)
}

// PreprocessCanny preprocesses Canny edge detection
func PreprocessCanny(image SDImage, highThreshold, lowThreshold, weak, strong float32, inverse bool) bool {
	return preprocessCanny(&image, highThreshold, lowThreshold, weak, strong, inverse)
}

// Commit gets commit information
func Commit() string {
	return CGoString(sdCommit())
}

// Version gets version information
func Version() string {
	return CGoString(sdVersion())
}

// HiresParamsInit initializes hi-res fix parameters with library defaults.
func HiresParamsInit(params *SDHiresParams) {
	sdHiresParamsInit(params)
}

// SDHiresUpscalerName gets the hi-res upscaler name.
func SDHiresUpscalerName(upscaler SDHiresUpscaler) string {
	return CGoString(sdHiresUpscalerName(upscaler))
}

// StrToSDHiresUpscaler converts a string to a hi-res upscaler.
func StrToSDHiresUpscaler(str string) SDHiresUpscaler {
	cStr := CString(str)
	defer FreeCString(cStr)
	return strToSDHiresUpscaler(cStr)
}

// SupportsImageGeneration reports whether the loaded model can generate images.
func (ctx *SDContext) SupportsImageGeneration() bool {
	return sdCtxSupportsImageGeneration(ctx.ptr)
}

// SupportsVideoGeneration reports whether the loaded model can generate video.
func (ctx *SDContext) SupportsVideoGeneration() bool {
	return sdCtxSupportsVideoGeneration(ctx.ptr)
}

// FreeAudio releases a native audio buffer returned by generate_video. It is
// safe to call with a nil audio or when the native binding is unavailable (the
// library has not been loaded / the symbol was not registered), in which case
// it does nothing — consistent with FreeImage/FreeImages.
func FreeAudio(audio *SDAudio) {
	if audio != nil && freeSDAudio != nil {
		freeSDAudio(audio)
	}
}

// Helper function: Convert C string to Go string
func CGoString(cStr *uint8) string {
	if cStr == nil {
		return ""
	}

	// Calculate string length
	var len int
	for p := cStr; *p != 0; p = (*uint8)(unsafe.Add(unsafe.Pointer(p), 1)) {
		len++
	}

	// Convert to Go string
	return string(unsafe.Slice(cStr, len))
}

// Helper function: Convert Go string to C string
func CString(str string) *uint8 {
	if str == "" {
		return nil
	}

	// Allocate memory, including the NULL terminator
	buf := make([]uint8, len(str)+1)
	copy(buf, str)
	buf[len(str)] = 0

	// Return pointer to buffer
	return &buf[0]
}

// Helper function: Free C string
func FreeCString(cStr *uint8) {
	// In Go, we use slices to manage memory, so no need to free
	// This function is just to maintain API consistency
}
