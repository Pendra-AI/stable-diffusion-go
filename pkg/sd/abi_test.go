package sd

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

// The C side of these tests is pkg/sd/testdata/abi_layout.txt, printed by
// csrc/test/abi_layout.c compiled against include/stable-diffusion.h at the pin
// in lib/version.txt. The compiler decided every number in it, and
// scripts/check-abi-layout.sh fails CI if the header stops agreeing with it.
//
// Everything here assumes 64-bit pointers. The libraries are only built for
// 64-bit targets, and for these structs (no long, no bitfields) the layouts of
// linux/amd64, linux/arm64, darwin/arm64 and windows/amd64 are identical, so the
// file holds for each of them.

type cField struct {
	name         string
	offset, size uintptr
	kind         string
}

type cStruct struct {
	size   uintptr
	fields []cField
}

// readABILayout parses abi_layout.txt into its structs and enumerators.
func readABILayout(t *testing.T) (map[string]*cStruct, map[string]int32) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "abi_layout.txt"))
	if err != nil {
		t.Fatalf("open abi_layout.txt: %v", err)
	}
	defer f.Close()

	number := func(s string) uint64 {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			t.Fatalf("bad number %q in abi_layout.txt: %v", s, err)
		}
		return n
	}

	structs := map[string]*cStruct{}
	enums := map[string]int32{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.Fields(sc.Text())
		switch {
		case len(w) == 3 && w[0] == "struct":
			structs[w[1]] = &cStruct{size: uintptr(number(w[2]))}
		case len(w) == 6 && w[0] == "field":
			s := structs[w[1]]
			if s == nil {
				t.Fatalf("field line before its struct line: %q", sc.Text())
			}
			s.fields = append(s.fields, cField{w[2], uintptr(number(w[3])), uintptr(number(w[4])), w[5]})
		case len(w) == 4 && w[0] == "enum":
			v, err := strconv.ParseInt(w[3], 10, 32)
			if err != nil {
				t.Fatalf("bad enum value in %q: %v", sc.Text(), err)
			}
			enums[w[2]] = int32(v)
		default:
			t.Fatalf("unrecognised line in abi_layout.txt: %q", sc.Text())
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read abi_layout.txt: %v", err)
	}
	return structs, enums
}

// goStructs maps every struct in the header to the Go type that mirrors it.
var goStructs = map[string]reflect.Type{
	"sd_tiling_params_t":           reflect.TypeOf(SDTilingParams{}),
	"sd_embedding_t":               reflect.TypeOf(SDEmbedding{}),
	"sd_ctx_params_t":              reflect.TypeOf(SDContextParams{}),
	"sd_audio_t":                   reflect.TypeOf(SDAudio{}),
	"sd_image_t":                   reflect.TypeOf(SDImage{}),
	"sd_image_result_t":            reflect.TypeOf(SDImageResult{}),
	"sd_image_generation_result_t": reflect.TypeOf(cImageGenerationResult{}),
	"sd_video_generation_result_t": reflect.TypeOf(cVideoGenerationResult{}),
	"sd_image_preprocess_params_t": reflect.TypeOf(SDImagePreprocessParams{}),
	"sd_ref_video_t":               reflect.TypeOf(SDRefVideo{}),
	"sd_slg_params_t":              reflect.TypeOf(SDSLGParams{}),
	"sd_guidance_params_t":         reflect.TypeOf(SDGuidanceParams{}),
	"sd_sample_params_t":           reflect.TypeOf(SDSampleParams{}),
	"sd_pm_params_t":               reflect.TypeOf(SDPMParams{}),
	"sd_pulid_params_t":            reflect.TypeOf(SDPulidParams{}),
	"sd_cache_params_t":            reflect.TypeOf(SDCacheParams{}),
	"sd_lora_t":                    reflect.TypeOf(SDLora{}),
	"sd_hires_params_t":            reflect.TypeOf(SDHiresParams{}),
	"sd_img_gen_params_t":          reflect.TypeOf(SDImgGenParams{}),
	"sd_vid_gen_params_t":          reflect.TypeOf(SDVidGenParams{}),
	"sd_preview_info_t":            reflect.TypeOf(SDPreviewInfo{}),
}

// goFieldAliases lists the Go fields whose name is not the C name with the
// underscores and case removed.
var goFieldAliases = map[string]string{
	"forcesdxlvaeconvscale": "forcesdxlvaconvscale",
}

func normaliseFieldName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

// goKind classifies a Go field the way KIND in abi_layout.c classifies a C one.
func goKind(k reflect.Kind) string {
	switch k {
	case reflect.Bool:
		return "bool"
	case reflect.Int32, reflect.Uint32:
		return "int32"
	case reflect.Int64, reflect.Uint64, reflect.Uintptr:
		return "int64"
	case reflect.Float32:
		return "float32"
	case reflect.Pointer, reflect.UnsafePointer, reflect.Struct:
		return "other"
	}
	return "unsupported " + k.String()
}

// TestStructLayoutMatchesHeader is the main proof of the ABI port. For every
// struct crossing the boundary it checks the size and, field by field and in
// order, the name, offset, size and kind against what the C compiler printed.
// The name check is what catches two same-typed fields being transposed or a
// field being replaced in place, which offsets alone cannot see.
func TestStructLayoutMatchesHeader(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("abi_layout.txt assumes 64-bit pointers")
	}
	structs, _ := readABILayout(t)

	for cName := range structs {
		if _, ok := goStructs[cName]; !ok {
			t.Errorf("%s has no Go mirror in goStructs", cName)
		}
	}
	for cName, typ := range goStructs {
		c, ok := structs[cName]
		if !ok {
			t.Errorf("%s (%s) is not in abi_layout.txt", cName, typ.Name())
			continue
		}
		t.Run(cName, func(t *testing.T) {
			if typ.Size() != c.size {
				t.Errorf("sizeof %s = %d, Go %s = %d", cName, c.size, typ.Name(), typ.Size())
			}
			if typ.NumField() != len(c.fields) {
				t.Fatalf("%s has %d fields, Go %s has %d", cName, len(c.fields), typ.Name(), typ.NumField())
			}
			for i, cf := range c.fields {
				gf := typ.Field(i)
				want := normaliseFieldName(cf.name)
				if alias, ok := goFieldAliases[want]; ok {
					want = alias
				}
				if got := normaliseFieldName(gf.Name); got != want {
					t.Errorf("field %d: C %s, Go %s", i, cf.name, gf.Name)
				}
				if gf.Offset != cf.offset {
					t.Errorf("%s: C offset %d, Go offset %d", cf.name, cf.offset, gf.Offset)
				}
				if gf.Type.Size() != cf.size {
					t.Errorf("%s: C size %d, Go size %d", cf.name, cf.size, gf.Type.Size())
				}
				if got := goKind(gf.Type.Kind()); got != cf.kind {
					t.Errorf("%s: C kind %s, Go kind %s", cf.name, cf.kind, got)
				}
			}
		})
	}
}

// TestStructLayoutKeyOffsets repeats the offsets a port is most likely to get
// wrong as plain unsafe.Sizeof/Offsetof constants, printed by the same C program,
// so a change to any of them shows up in review without opening the data file.
func TestStructLayoutKeyOffsets(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("these offsets assume 64-bit pointers")
	}
	var (
		ctx     SDContextParams
		img     SDImgGenParams
		vid     SDVidGenParams
		tiling  SDTilingParams
		item    SDImageResult
		imgRes  cImageGenerationResult
		vidRes  cVideoGenerationResult
		refVid  SDRefVideo
		preview SDPreviewInfo
	)
	cases := []struct {
		name      string
		got, want uintptr
	}{
		{"sizeof sd_ctx_params_t", unsafe.Sizeof(ctx), 336},
		{"sd_ctx_params_t.audio_encoder_path", unsafe.Offsetof(ctx.AudioEncoderPath), 104},
		{"sd_ctx_params_t.taesd_path", unsafe.Offsetof(ctx.TAESDPath), 112},
		{"sd_ctx_params_t.max_vram", unsafe.Offsetof(ctx.MaxVRAM), 224},
		{"sd_ctx_params_t.disable_prefetch", unsafe.Offsetof(ctx.DisablePrefetch), 232},
		{"sd_ctx_params_t.eager_load", unsafe.Offsetof(ctx.EagerLoad), 233},
		{"sd_ctx_params_t.model_args", unsafe.Offsetof(ctx.ModelArgs), 280},
		{"sd_ctx_params_t.disable_segmented_compute", unsafe.Offsetof(ctx.DisableSegmentedCompute), 288},
		{"sd_ctx_params_t.linear_scale", unsafe.Offsetof(ctx.LinearScale), 292},
		{"sd_ctx_params_t.tokenizer", unsafe.Offsetof(ctx.Tokenizer), 304},
		{"sd_ctx_params_t.conv_direct_mode", unsafe.Offsetof(ctx.ConvDirectMode), 328},

		{"sizeof sd_tiling_params_t", unsafe.Sizeof(tiling), 32},
		{"sd_tiling_params_t.tile_size_w", unsafe.Offsetof(tiling.TileSizeW), 4},
		{"sd_tiling_params_t.target_overlap", unsafe.Offsetof(tiling.TargetOverlap), 12},
		{"sd_tiling_params_t.extra_tiling_args", unsafe.Offsetof(tiling.ExtraTilingArgs), 24},

		{"sizeof sd_img_gen_params_t", unsafe.Sizeof(img), 552},
		{"sd_img_gen_params_t.sample_params", unsafe.Offsetof(img.SampleParams), 120},
		{"sd_img_gen_params_t.seed", unsafe.Offsetof(img.Seed), 224},
		{"sd_img_gen_params_t.vae_tiling_params", unsafe.Offsetof(img.VAETilingParams), 352},
		{"sd_img_gen_params_t.hires", unsafe.Offsetof(img.Hires), 480},
		{"sd_img_gen_params_t.image_preprocess", unsafe.Offsetof(img.ImagePreprocess), 544},

		{"sizeof sd_vid_gen_params_t", unsafe.Sizeof(vid), 592},
		{"sd_vid_gen_params_t.ref_images", unsafe.Offsetof(vid.RefImages), 88},
		{"sd_vid_gen_params_t.ref_videos", unsafe.Offsetof(vid.RefVideos), 104},
		{"sd_vid_gen_params_t.ref_audios", unsafe.Offsetof(vid.RefAudios), 120},
		{"sd_vid_gen_params_t.control_frames", unsafe.Offsetof(vid.ControlFrames), 136},
		{"sd_vid_gen_params_t.seed", unsafe.Offsetof(vid.Seed), 360},
		{"sd_vid_gen_params_t.image_preprocess", unsafe.Offsetof(vid.ImagePreprocess), 576},
		{"sd_vid_gen_params_t.ref_image_args", unsafe.Offsetof(vid.RefImageArgs), 584},

		{"sizeof sd_image_result_t", unsafe.Sizeof(item), 32},
		{"sd_image_result_t.generation_info", unsafe.Offsetof(item.GenerationInfo), 24},
		{"sizeof sd_image_generation_result_t", unsafe.Sizeof(imgRes), 16},
		{"sd_image_generation_result_t.count", unsafe.Offsetof(imgRes.Count), 8},
		{"sizeof sd_video_generation_result_t", unsafe.Sizeof(vidRes), 40},
		{"sd_video_generation_result_t.frame_count", unsafe.Offsetof(vidRes.FrameCount), 8},
		{"sd_video_generation_result_t.audio", unsafe.Offsetof(vidRes.Audio), 16},
		{"sd_video_generation_result_t.fps", unsafe.Offsetof(vidRes.Fps), 24},
		{"sd_video_generation_result_t.generation_info", unsafe.Offsetof(vidRes.GenerationInfo), 32},
		{"sizeof sd_ref_video_t", unsafe.Sizeof(refVid), 40},
		{"sd_ref_video_t.audio", unsafe.Offsetof(refVid.Audio), 16},
		{"sizeof sd_preview_info_t", unsafe.Sizeof(preview), 8},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

// goEnums lists every enumerator in the header with the Go constant mirroring it.
var goEnums = map[string]int32{
	"STD_DEFAULT_RNG": int32(DefaultRNG),
	"CUDA_RNG":        int32(CUDARNG),
	"CPU_RNG":         int32(CPURNG),
	"RNG_TYPE_COUNT":  int32(RNGTypeCount),

	"EULER_SAMPLE_METHOD":          int32(EulerSampleMethod),
	"EULER_A_SAMPLE_METHOD":        int32(EulerASampleMethod),
	"HEUN_SAMPLE_METHOD":           int32(HeunSampleMethod),
	"DPM2_SAMPLE_METHOD":           int32(DPM2SampleMethod),
	"DPMPP2S_A_SAMPLE_METHOD":      int32(DPMPP2SASampleMethod),
	"DPMPP2M_SAMPLE_METHOD":        int32(DPMPP2MSampleMethod),
	"DPMPP2Mv2_SAMPLE_METHOD":      int32(DPMPP2Mv2SampleMethod),
	"IPNDM_SAMPLE_METHOD":          int32(IPNDMSampleMethod),
	"IPNDM_V_SAMPLE_METHOD":        int32(IPNDMSampleMethodV),
	"LCM_SAMPLE_METHOD":            int32(LCMSampleMethod),
	"DDIM_TRAILING_SAMPLE_METHOD":  int32(DDIMTrailingSampleMethod),
	"TCD_SAMPLE_METHOD":            int32(TCDSampleMethod),
	"RES_MULTISTEP_SAMPLE_METHOD":  int32(ResMultistepSampleMethod),
	"RES_2S_SAMPLE_METHOD":         int32(Res2SSampleMethod),
	"ER_SDE_SAMPLE_METHOD":         int32(ERSDESampleMethod),
	"EULER_CFG_PP_SAMPLE_METHOD":   int32(EulerCFGPPSampleMethod),
	"EULER_A_CFG_PP_SAMPLE_METHOD": int32(EulerACFGPPSampleMethod),
	"EULER_GE_SAMPLE_METHOD":       int32(EulerGESampleMethod),
	"DPMPP2M_SDE_SAMPLE_METHOD":    int32(DPMPP2MSDESampleMethod),
	"DPMPP2M_SDE_BT_SAMPLE_METHOD": int32(DPMPP2MSDEBTSampleMethod),
	"LMS_SAMPLE_METHOD":            int32(LMSSampleMethod),
	"SAMPLE_METHOD_COUNT":          int32(SampleMethodCount),

	"DISCRETE_SCHEDULER":     int32(DiscreteScheduler),
	"KARRAS_SCHEDULER":       int32(KarrasScheduler),
	"EXPONENTIAL_SCHEDULER":  int32(ExponentialScheduler),
	"AYS_SCHEDULER":          int32(AYSScheduler),
	"GITS_SCHEDULER":         int32(GITScheduler),
	"SGM_UNIFORM_SCHEDULER":  int32(SGMUniformScheduler),
	"SIMPLE_SCHEDULER":       int32(SimpleScheduler),
	"SMOOTHSTEP_SCHEDULER":   int32(SmoothstepScheduler),
	"KL_OPTIMAL_SCHEDULER":   int32(KLOptimalScheduler),
	"LCM_SCHEDULER":          int32(LCMScheduler),
	"BONG_TANGENT_SCHEDULER": int32(BongTangentScheduler),
	"LTX2_SCHEDULER":         int32(LTX2Scheduler),
	"LOGIT_NORMAL_SCHEDULER": int32(LogitNormalScheduler),
	"FLUX2_SCHEDULER":        int32(Flux2Scheduler),
	"FLUX_SCHEDULER":         int32(FluxScheduler),
	"BETA_SCHEDULER":         int32(BetaScheduler),
	"LLADA_IMAGE_SCHEDULER":  int32(LLaDAImageScheduler),
	"SCHEDULER_COUNT":        int32(SchedulerCount),

	"EPS_PRED":               int32(EPSPred),
	"V_PRED":                 int32(VPred),
	"EDM_V_PRED":             int32(EDMVPred),
	"FLOW_PRED":              int32(FlowPred),
	"FLUX_FLOW_PRED":         int32(FluxFlowPred),
	"SEFI_FLOW_PRED":         int32(SefiFlowPred),
	"MINIT2I_FLOW_PRED":      int32(Minit2iFlowPred),
	"SENSENOVA_U1_FLOW_PRED": int32(SenseNovaU1FlowPred),
	"FLUX3_FLOW_PRED":        int32(Flux3FlowPred),
	"H3_AV_FLOW_PRED":        int32(H3AVFlowPred),
	"MING_IMAGE_FLOW_PRED":   int32(MingImageFlowPred),
	"PREDICTION_COUNT":       int32(PredictionCount),

	"SD_TYPE_F32":     int32(SDTypeF32),
	"SD_TYPE_F16":     int32(SDTypeF16),
	"SD_TYPE_Q4_0":    int32(SDTypeQ4_0),
	"SD_TYPE_Q4_1":    int32(SDTypeQ4_1),
	"SD_TYPE_Q5_0":    SDTypeQ5_0,
	"SD_TYPE_Q5_1":    SDTypeQ5_1,
	"SD_TYPE_Q8_0":    SDTypeQ8_0,
	"SD_TYPE_Q8_1":    SDTypeQ8_1,
	"SD_TYPE_Q2_K":    SDTypeQ2_K,
	"SD_TYPE_Q3_K":    SDTypeQ3_K,
	"SD_TYPE_Q4_K":    SDTypeQ4_K,
	"SD_TYPE_Q5_K":    SDTypeQ5_K,
	"SD_TYPE_Q6_K":    SDTypeQ6_K,
	"SD_TYPE_Q8_K":    SDTypeQ8_K,
	"SD_TYPE_IQ2_XXS": SDTypeIQ2_XXS,
	"SD_TYPE_IQ2_XS":  SDTypeIQ2_XS,
	"SD_TYPE_IQ3_XXS": SDTypeIQ3_XXS,
	"SD_TYPE_IQ1_S":   SDTypeIQ1_S,
	"SD_TYPE_IQ4_NL":  SDTypeIQ4_NL,
	"SD_TYPE_IQ3_S":   SDTypeIQ3_S,
	"SD_TYPE_IQ2_S":   SDTypeIQ2_S,
	"SD_TYPE_IQ4_XS":  SDTypeIQ4_XS,
	"SD_TYPE_I8":      SDTypeI8,
	"SD_TYPE_I16":     SDTypeI16,
	"SD_TYPE_I32":     SDTypeI32,
	"SD_TYPE_I64":     SDTypeI64,
	"SD_TYPE_F64":     SDTypeF64,
	"SD_TYPE_IQ1_M":   SDTypeIQ1_M,
	"SD_TYPE_BF16":    SDTypeBF16,
	"SD_TYPE_TQ1_0":   SDTypeTQ1_0,
	"SD_TYPE_TQ2_0":   SDTypeTQ2_0,
	"SD_TYPE_MXFP4":   SDTypeMXFP4,
	"SD_TYPE_NVFP4":   SDTypeNVFP4,
	"SD_TYPE_Q1_0":    SDTypeQ1_0,
	"SD_TYPE_Q2_0":    SDTypeQ2_0,
	"SD_TYPE_F8_E4M3": SDTypeF8_E4M3,
	"SD_TYPE_F8_E5M2": SDTypeF8_E5M2,
	"SD_TYPE_COUNT":   SDTypeCount,

	"SD_LOG_DEBUG":   int32(SDLogDebug),
	"SD_LOG_VERBOSE": int32(SDLogVerbose),
	"SD_LOG_INFO":    int32(SDLogInfo),
	"SD_LOG_WARN":    int32(SDLogWarn),
	"SD_LOG_ERROR":   int32(SDLogError),

	"PREVIEW_NONE":  int32(PreviewNone),
	"PREVIEW_PROJ":  int32(PreviewProj),
	"PREVIEW_TAE":   int32(PreviewTAE),
	"PREVIEW_VAE":   int32(PreviewVAE),
	"PREVIEW_COUNT": int32(PreviewCount),

	"LORA_APPLY_AUTO":        int32(LoraApplyAuto),
	"LORA_APPLY_IMMEDIATELY": int32(LoraApplyImmediately),
	"LORA_APPLY_AT_RUNTIME":  int32(LoraApplyAtRuntime),
	"LORA_APPLY_MODE_COUNT":  int32(LoraApplyModeCount),

	"SD_VAE_FORMAT_AUTO":  int32(VAEFormatAuto),
	"SD_VAE_FORMAT_FLUX":  int32(FluxVAEFormat),
	"SD_VAE_FORMAT_SD3":   int32(SD3VAEFormat),
	"SD_VAE_FORMAT_FLUX2": int32(Flux2VAEFormat),
	"SD_VAE_FORMAT_WAN":   int32(WanVAEFormat),
	"SD_VAE_FORMAT_COUNT": int32(VAEFormatCount),

	"SD_CACHE_DISABLED":   int32(SDCacheDisabled),
	"SD_CACHE_EASYCACHE":  int32(SDCacheEasycache),
	"SD_CACHE_UCACHE":     int32(SDCacheUcache),
	"SD_CACHE_DBCACHE":    int32(SDCacheDbcache),
	"SD_CACHE_TAYLORSEER": int32(SDCacheTaylorseer),
	"SD_CACHE_CACHE_DIT":  int32(SDCacheCacheDit),
	"SD_CACHE_SPECTRUM":   int32(SDCacheSpectrum),

	"SD_HIRES_UPSCALER_NONE":                       int32(HiresUpscalerNone),
	"SD_HIRES_UPSCALER_LATENT":                     int32(HiresUpscalerLatent),
	"SD_HIRES_UPSCALER_LATENT_NEAREST":             int32(HiresUpscalerLatentNearest),
	"SD_HIRES_UPSCALER_LATENT_NEAREST_EXACT":       int32(HiresUpscalerLatentNearestExact),
	"SD_HIRES_UPSCALER_LATENT_ANTIALIASED":         int32(HiresUpscalerLatentAntialiased),
	"SD_HIRES_UPSCALER_LATENT_BICUBIC":             int32(HiresUpscalerLatentBicubic),
	"SD_HIRES_UPSCALER_LATENT_BICUBIC_ANTIALIASED": int32(HiresUpscalerLatentBicubicAntialiased),
	"SD_HIRES_UPSCALER_LANCZOS":                    int32(HiresUpscalerLanczos),
	"SD_HIRES_UPSCALER_NEAREST":                    int32(HiresUpscalerNearest),
	"SD_HIRES_UPSCALER_MODEL":                      int32(HiresUpscalerModel),
	"SD_HIRES_UPSCALER_COUNT":                      int32(HiresUpscalerCount),

	"SD_CANCEL_ALL":         int32(CancelAll),
	"SD_CANCEL_NEW_LATENTS": int32(CancelNewLatents),
	"SD_CANCEL_RESET":       int32(CancelReset),
}

// TestEnumsMatchHeader checks every enumerator the C compiler printed against
// its Go constant, in both directions, so a value added upstream cannot go
// unnoticed even when no _COUNT sentinel moves.
func TestEnumsMatchHeader(t *testing.T) {
	_, enums := readABILayout(t)
	for name, want := range enums {
		got, ok := goEnums[name]
		if !ok {
			t.Errorf("%s has no Go mirror in goEnums", name)
			continue
		}
		if got != want {
			t.Errorf("%s = %d in the header, %d in Go", name, want, got)
		}
	}
	for name := range goEnums {
		if _, ok := enums[name]; !ok {
			t.Errorf("%s is in goEnums but not in the header", name)
		}
	}
	if len(enums) == 0 {
		t.Fatal("no enumerators parsed from abi_layout.txt")
	}
}
