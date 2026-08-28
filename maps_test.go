package stable_diffusion

import (
	"testing"

	"github.com/pendra-ai/stable-diffusion-go/pkg/sd"
)

// TestNewEnumMapEntries guards the string→enum convenience maps for the values
// added during the master-685 and master-802 upstream resyncs. The map keys
// are this binding's own convention; the values must match the corresponding
// pkg/sd constants (which are pinned to the C ABI by
// pkg/sd.TestEnumABIValues). A drift here would silently mis-route a caller's
// "euler_cfg_pp" to the wrong sampler.
func TestNewEnumMapEntries(t *testing.T) {
	t.Run("SampleMethodMap", func(t *testing.T) {
		want := map[string]sd.SampleMethod{
			"res_multistep":  sd.ResMultistepSampleMethod,
			"res_2s":         sd.Res2SSampleMethod,
			"er_sde":         sd.ERSDESampleMethod,
			"euler_cfg_pp":   sd.EulerCFGPPSampleMethod,
			"euler_a_cfg_pp": sd.EulerACFGPPSampleMethod,
			"euler_ge":       sd.EulerGESampleMethod,
			"dpm++2m_sde":    sd.DPMPP2MSDESampleMethod,
			"dpm++2m_sde_bt": sd.DPMPP2MSDEBTSampleMethod,
		}
		for k, v := range want {
			if got, ok := SampleMethodMap[k]; !ok || got != v {
				t.Errorf("SampleMethodMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})

	t.Run("SchedulerMap", func(t *testing.T) {
		want := map[string]sd.Scheduler{
			"bong_tangent": sd.BongTangentScheduler,
			"ltx2":         sd.LTX2Scheduler,
			"logit_normal": sd.LogitNormalScheduler,
			"flux2":        sd.Flux2Scheduler,
			"flux":         sd.FluxScheduler,
			"beta":         sd.BetaScheduler,
		}
		for k, v := range want {
			if got, ok := SchedulerMap[k]; !ok || got != v {
				t.Errorf("SchedulerMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})

	t.Run("SDTypeMap", func(t *testing.T) {
		want := map[string]sd.SDType{
			"nvfp4": sd.SDTypeNVFP4,
			"q1_0":  sd.SDTypeQ1_0,
		}
		for k, v := range want {
			if got, ok := SDTypeMap[k]; !ok || got != v {
				t.Errorf("SDTypeMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})

	t.Run("PredictionMap", func(t *testing.T) {
		want := map[string]sd.Prediction{
			"eps":          sd.EPSPred,
			"v":            sd.VPred,
			"edm_v":        sd.EDMVPred,
			"flow":         sd.FlowPred,
			"sd3_flow":     sd.FlowPred,
			"flux_flow":    sd.FluxFlowPred,
			"sefi_flow":    sd.SefiFlowPred,
			"minit2i_flow": sd.Minit2iFlowPred,
			"default":      sd.PredictionCount,
		}
		if len(PredictionMap) != len(want) {
			t.Errorf("PredictionMap has %d entries, want %d", len(PredictionMap), len(want))
		}
		if _, ok := PredictionMap["flux2_flow"]; ok {
			t.Error("PredictionMap still contains \"flux2_flow\", which upstream removed (its enum slot is now sefi_flow)")
		}
		for k, v := range want {
			if got, ok := PredictionMap[k]; !ok || got != v {
				t.Errorf("PredictionMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})

	t.Run("VAEFormatMap", func(t *testing.T) {
		want := map[string]sd.SDVAEFormat{
			"auto":  sd.VAEFormatAuto,
			"flux":  sd.FluxVAEFormat,
			"sd3":   sd.SD3VAEFormat,
			"flux2": sd.Flux2VAEFormat,
			"wan":   sd.WanVAEFormat,
		}
		if len(VAEFormatMap) != len(want) {
			t.Errorf("VAEFormatMap has %d entries, want %d", len(VAEFormatMap), len(want))
		}
		for k, v := range want {
			if got, ok := VAEFormatMap[k]; !ok || got != v {
				t.Errorf("VAEFormatMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})

	t.Run("HiresUpscalerMap", func(t *testing.T) {
		want := map[string]sd.SDHiresUpscaler{
			"none":                       sd.HiresUpscalerNone,
			"latent":                     sd.HiresUpscalerLatent,
			"latent_nearest":             sd.HiresUpscalerLatentNearest,
			"latent_nearest_exact":       sd.HiresUpscalerLatentNearestExact,
			"latent_antialiased":         sd.HiresUpscalerLatentAntialiased,
			"latent_bicubic":             sd.HiresUpscalerLatentBicubic,
			"latent_bicubic_antialiased": sd.HiresUpscalerLatentBicubicAntialiased,
			"lanczos":                    sd.HiresUpscalerLanczos,
			"nearest":                    sd.HiresUpscalerNearest,
			"model":                      sd.HiresUpscalerModel,
		}
		if len(HiresUpscalerMap) != len(want) {
			t.Errorf("HiresUpscalerMap has %d entries, want %d", len(HiresUpscalerMap), len(want))
		}
		for k, v := range want {
			if got, ok := HiresUpscalerMap[k]; !ok || got != v {
				t.Errorf("HiresUpscalerMap[%q] = %v (ok=%v), want %v", k, got, ok, v)
			}
		}
	})
}
