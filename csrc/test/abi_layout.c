// abi_layout.c prints the memory layout of every struct, and the value of every
// enumerator, that this binding mirrors from include/stable-diffusion.h. It is
// the source of truth for pkg/sd/testdata/abi_layout.txt, which pkg/sd/abi_test.go
// compares against the Go structs, because the compiler (not a hand count)
// decides padding.
//
// scripts/check-abi-layout.sh compiles it against an upstream tree and diffs the
// output against that file. Regenerate the file with:
//
//   scripts/check-abi-layout.sh <upstream-dir> --update
//
// sd_adetailer_params_t is left out because the binding does not mirror it.
#include <stddef.h>
#include <stdio.h>

#include "stable-diffusion.h"

#if defined(__LP64__)
#define LONG_BITS "64"
#else
#define LONG_BITS "32"
#endif

// KIND classifies a field by how it crosses the FFI boundary. Enums are `int` or
// `unsigned`, depending on the compiler, so every 4-byte integer is one kind.
#define KIND(x) _Generic((x),         \
    bool: "bool",                     \
    int: "int32",                     \
    unsigned int: "int32",            \
    float: "float32",                 \
    long: "int" LONG_BITS,            \
    unsigned long: "int" LONG_BITS,   \
    long long: "int64",               \
    unsigned long long: "int64",      \
    default: "other")

#define STRUCT(t) printf("struct %s %zu\n", #t, sizeof(t))
#define FIELD(t, f)                                              \
    printf("field %s %s %zu %zu %s\n", #t, #f, offsetof(t, f),   \
           sizeof(((t*)0)->f), KIND(((t*)0)->f))
#define ENUM(t, v) printf("enum %s %s %d\n", #t, #v, (int)(v))

int main(void) {
    STRUCT(sd_tiling_params_t);
    FIELD(sd_tiling_params_t, enabled);
    FIELD(sd_tiling_params_t, temporal_tiling);
    FIELD(sd_tiling_params_t, tile_size_w);
    FIELD(sd_tiling_params_t, tile_size_h);
    FIELD(sd_tiling_params_t, target_overlap);
    FIELD(sd_tiling_params_t, rel_size_w);
    FIELD(sd_tiling_params_t, rel_size_h);
    FIELD(sd_tiling_params_t, extra_tiling_args);

    STRUCT(sd_embedding_t);
    FIELD(sd_embedding_t, name);
    FIELD(sd_embedding_t, path);

    STRUCT(sd_ctx_params_t);
    FIELD(sd_ctx_params_t, model_path);
    FIELD(sd_ctx_params_t, clip_l_path);
    FIELD(sd_ctx_params_t, clip_g_path);
    FIELD(sd_ctx_params_t, clip_vision_path);
    FIELD(sd_ctx_params_t, t5xxl_path);
    FIELD(sd_ctx_params_t, llm_path);
    FIELD(sd_ctx_params_t, llm_vision_path);
    FIELD(sd_ctx_params_t, diffusion_model_path);
    FIELD(sd_ctx_params_t, high_noise_diffusion_model_path);
    FIELD(sd_ctx_params_t, uncond_diffusion_model_path);
    FIELD(sd_ctx_params_t, embeddings_connectors_path);
    FIELD(sd_ctx_params_t, vae_path);
    FIELD(sd_ctx_params_t, audio_vae_path);
    FIELD(sd_ctx_params_t, audio_encoder_path);
    FIELD(sd_ctx_params_t, taesd_path);
    FIELD(sd_ctx_params_t, control_net_path);
    FIELD(sd_ctx_params_t, ip_adapter_path);
    FIELD(sd_ctx_params_t, motion_module_path);
    FIELD(sd_ctx_params_t, embeddings);
    FIELD(sd_ctx_params_t, embedding_count);
    FIELD(sd_ctx_params_t, photo_maker_path);
    FIELD(sd_ctx_params_t, pulid_weights_path);
    FIELD(sd_ctx_params_t, tensor_type_rules);
    FIELD(sd_ctx_params_t, n_threads);
    FIELD(sd_ctx_params_t, wtype);
    FIELD(sd_ctx_params_t, rng_type);
    FIELD(sd_ctx_params_t, sampler_rng_type);
    FIELD(sd_ctx_params_t, prediction);
    FIELD(sd_ctx_params_t, lora_apply_mode);
    FIELD(sd_ctx_params_t, enable_mmap);
    FIELD(sd_ctx_params_t, flash_attn);
    FIELD(sd_ctx_params_t, diffusion_flash_attn);
    FIELD(sd_ctx_params_t, tae_preview_only);
    FIELD(sd_ctx_params_t, diffusion_conv_direct);
    FIELD(sd_ctx_params_t, vae_conv_direct);
    FIELD(sd_ctx_params_t, force_sdxl_vae_conv_scale);
    FIELD(sd_ctx_params_t, vae_format);
    FIELD(sd_ctx_params_t, max_vram);
    FIELD(sd_ctx_params_t, disable_prefetch);
    FIELD(sd_ctx_params_t, eager_load);
    FIELD(sd_ctx_params_t, backend);
    FIELD(sd_ctx_params_t, params_backend);
    FIELD(sd_ctx_params_t, split_mode);
    FIELD(sd_ctx_params_t, auto_fit);
    FIELD(sd_ctx_params_t, rpc_servers);
    FIELD(sd_ctx_params_t, model_args);
    FIELD(sd_ctx_params_t, disable_segmented_compute);
    FIELD(sd_ctx_params_t, batched_cfg);
    FIELD(sd_ctx_params_t, linear_scale);
    FIELD(sd_ctx_params_t, attn_scale);
    FIELD(sd_ctx_params_t, tokenizer);
    FIELD(sd_ctx_params_t, sage_attn);
    FIELD(sd_ctx_params_t, conditioning_cache_size);
    FIELD(sd_ctx_params_t, flash_attn_mode);
    FIELD(sd_ctx_params_t, conv_direct_mode);

    STRUCT(sd_audio_t);
    FIELD(sd_audio_t, sample_rate);
    FIELD(sd_audio_t, channels);
    FIELD(sd_audio_t, sample_count);
    FIELD(sd_audio_t, data);

    STRUCT(sd_image_t);
    FIELD(sd_image_t, width);
    FIELD(sd_image_t, height);
    FIELD(sd_image_t, channel);
    FIELD(sd_image_t, data);

    STRUCT(sd_image_result_t);
    FIELD(sd_image_result_t, image);
    FIELD(sd_image_result_t, generation_info);

    STRUCT(sd_image_generation_result_t);
    FIELD(sd_image_generation_result_t, items);
    FIELD(sd_image_generation_result_t, count);

    STRUCT(sd_video_generation_result_t);
    FIELD(sd_video_generation_result_t, frames);
    FIELD(sd_video_generation_result_t, frame_count);
    FIELD(sd_video_generation_result_t, audio);
    FIELD(sd_video_generation_result_t, fps);
    FIELD(sd_video_generation_result_t, generation_info);

    STRUCT(sd_image_preprocess_params_t);
    FIELD(sd_image_preprocess_params_t, rules);

    STRUCT(sd_ref_video_t);
    FIELD(sd_ref_video_t, frames);
    FIELD(sd_ref_video_t, frame_count);
    FIELD(sd_ref_video_t, fps);
    FIELD(sd_ref_video_t, audio);

    STRUCT(sd_slg_params_t);
    FIELD(sd_slg_params_t, layers);
    FIELD(sd_slg_params_t, layer_count);
    FIELD(sd_slg_params_t, layer_start);
    FIELD(sd_slg_params_t, layer_end);
    FIELD(sd_slg_params_t, scale);

    STRUCT(sd_guidance_params_t);
    FIELD(sd_guidance_params_t, txt_cfg);
    FIELD(sd_guidance_params_t, img_cfg);
    FIELD(sd_guidance_params_t, distilled_guidance);
    FIELD(sd_guidance_params_t, slg);

    STRUCT(sd_sample_params_t);
    FIELD(sd_sample_params_t, guidance);
    FIELD(sd_sample_params_t, scheduler);
    FIELD(sd_sample_params_t, sample_method);
    FIELD(sd_sample_params_t, sample_steps);
    FIELD(sd_sample_params_t, eta);
    FIELD(sd_sample_params_t, shifted_timestep);
    FIELD(sd_sample_params_t, custom_sigmas);
    FIELD(sd_sample_params_t, custom_sigmas_count);
    FIELD(sd_sample_params_t, flow_shift);
    FIELD(sd_sample_params_t, extra_sample_args);

    STRUCT(sd_pm_params_t);
    FIELD(sd_pm_params_t, id_images);
    FIELD(sd_pm_params_t, id_images_count);
    FIELD(sd_pm_params_t, id_embed_path);
    FIELD(sd_pm_params_t, style_strength);

    STRUCT(sd_pulid_params_t);
    FIELD(sd_pulid_params_t, id_embedding_path);
    FIELD(sd_pulid_params_t, id_weight);

    STRUCT(sd_cache_params_t);
    FIELD(sd_cache_params_t, mode);
    FIELD(sd_cache_params_t, reuse_threshold);
    FIELD(sd_cache_params_t, start_percent);
    FIELD(sd_cache_params_t, end_percent);
    FIELD(sd_cache_params_t, error_decay_rate);
    FIELD(sd_cache_params_t, use_relative_threshold);
    FIELD(sd_cache_params_t, reset_error_on_compute);
    FIELD(sd_cache_params_t, Fn_compute_blocks);
    FIELD(sd_cache_params_t, Bn_compute_blocks);
    FIELD(sd_cache_params_t, residual_diff_threshold);
    FIELD(sd_cache_params_t, max_warmup_steps);
    FIELD(sd_cache_params_t, max_cached_steps);
    FIELD(sd_cache_params_t, max_continuous_cached_steps);
    FIELD(sd_cache_params_t, taylorseer_n_derivatives);
    FIELD(sd_cache_params_t, taylorseer_skip_interval);
    FIELD(sd_cache_params_t, scm_mask);
    FIELD(sd_cache_params_t, scm_policy_dynamic);
    FIELD(sd_cache_params_t, spectrum_w);
    FIELD(sd_cache_params_t, spectrum_m);
    FIELD(sd_cache_params_t, spectrum_lam);
    FIELD(sd_cache_params_t, spectrum_window_size);
    FIELD(sd_cache_params_t, spectrum_flex_window);
    FIELD(sd_cache_params_t, spectrum_warmup_steps);
    FIELD(sd_cache_params_t, spectrum_stop_percent);

    STRUCT(sd_lora_t);
    FIELD(sd_lora_t, is_high_noise);
    FIELD(sd_lora_t, multiplier);
    FIELD(sd_lora_t, path);

    STRUCT(sd_hires_params_t);
    FIELD(sd_hires_params_t, enabled);
    FIELD(sd_hires_params_t, upscaler);
    FIELD(sd_hires_params_t, model_path);
    FIELD(sd_hires_params_t, scale);
    FIELD(sd_hires_params_t, target_width);
    FIELD(sd_hires_params_t, target_height);
    FIELD(sd_hires_params_t, steps);
    FIELD(sd_hires_params_t, denoising_strength);
    FIELD(sd_hires_params_t, upscale_tile_size);
    FIELD(sd_hires_params_t, custom_sigmas);
    FIELD(sd_hires_params_t, custom_sigmas_count);

    STRUCT(sd_img_gen_params_t);
    FIELD(sd_img_gen_params_t, loras);
    FIELD(sd_img_gen_params_t, lora_count);
    FIELD(sd_img_gen_params_t, prompt);
    FIELD(sd_img_gen_params_t, negative_prompt);
    FIELD(sd_img_gen_params_t, clip_skip);
    FIELD(sd_img_gen_params_t, init_image);
    FIELD(sd_img_gen_params_t, ref_images);
    FIELD(sd_img_gen_params_t, ref_images_count);
    FIELD(sd_img_gen_params_t, ref_image_args);
    FIELD(sd_img_gen_params_t, mask_image);
    FIELD(sd_img_gen_params_t, width);
    FIELD(sd_img_gen_params_t, height);
    FIELD(sd_img_gen_params_t, sample_params);
    FIELD(sd_img_gen_params_t, strength);
    FIELD(sd_img_gen_params_t, seed);
    FIELD(sd_img_gen_params_t, batch_count);
    FIELD(sd_img_gen_params_t, control_image);
    FIELD(sd_img_gen_params_t, control_strength);
    FIELD(sd_img_gen_params_t, ip_adapter_image);
    FIELD(sd_img_gen_params_t, ip_adapter_strength);
    FIELD(sd_img_gen_params_t, pm_params);
    FIELD(sd_img_gen_params_t, pulid_params);
    FIELD(sd_img_gen_params_t, vae_tiling_params);
    FIELD(sd_img_gen_params_t, cache);
    FIELD(sd_img_gen_params_t, hires);
    FIELD(sd_img_gen_params_t, qwen_image_layers);
    FIELD(sd_img_gen_params_t, circular_x);
    FIELD(sd_img_gen_params_t, circular_y);
    FIELD(sd_img_gen_params_t, image_preprocess);

    STRUCT(sd_vid_gen_params_t);
    FIELD(sd_vid_gen_params_t, loras);
    FIELD(sd_vid_gen_params_t, lora_count);
    FIELD(sd_vid_gen_params_t, prompt);
    FIELD(sd_vid_gen_params_t, negative_prompt);
    FIELD(sd_vid_gen_params_t, clip_skip);
    FIELD(sd_vid_gen_params_t, init_image);
    FIELD(sd_vid_gen_params_t, end_image);
    FIELD(sd_vid_gen_params_t, ref_images);
    FIELD(sd_vid_gen_params_t, ref_images_count);
    FIELD(sd_vid_gen_params_t, ref_videos);
    FIELD(sd_vid_gen_params_t, ref_videos_count);
    FIELD(sd_vid_gen_params_t, ref_audios);
    FIELD(sd_vid_gen_params_t, ref_audios_count);
    FIELD(sd_vid_gen_params_t, control_frames);
    FIELD(sd_vid_gen_params_t, control_frames_size);
    FIELD(sd_vid_gen_params_t, width);
    FIELD(sd_vid_gen_params_t, height);
    FIELD(sd_vid_gen_params_t, sample_params);
    FIELD(sd_vid_gen_params_t, high_noise_sample_params);
    FIELD(sd_vid_gen_params_t, moe_boundary);
    FIELD(sd_vid_gen_params_t, strength);
    FIELD(sd_vid_gen_params_t, seed);
    FIELD(sd_vid_gen_params_t, video_frames);
    FIELD(sd_vid_gen_params_t, fps);
    FIELD(sd_vid_gen_params_t, vace_strength);
    FIELD(sd_vid_gen_params_t, vae_tiling_params);
    FIELD(sd_vid_gen_params_t, cache);
    FIELD(sd_vid_gen_params_t, hires);
    FIELD(sd_vid_gen_params_t, circular_x);
    FIELD(sd_vid_gen_params_t, circular_y);
    FIELD(sd_vid_gen_params_t, image_preprocess);
    FIELD(sd_vid_gen_params_t, ref_image_args);

    STRUCT(sd_preview_info_t);
    FIELD(sd_preview_info_t, sample_pass);
    FIELD(sd_preview_info_t, total_steps);

    ENUM(rng_type_t, STD_DEFAULT_RNG);
    ENUM(rng_type_t, CUDA_RNG);
    ENUM(rng_type_t, CPU_RNG);
    ENUM(rng_type_t, RNG_TYPE_COUNT);

    ENUM(sample_method_t, EULER_SAMPLE_METHOD);
    ENUM(sample_method_t, EULER_A_SAMPLE_METHOD);
    ENUM(sample_method_t, HEUN_SAMPLE_METHOD);
    ENUM(sample_method_t, DPM2_SAMPLE_METHOD);
    ENUM(sample_method_t, DPMPP2S_A_SAMPLE_METHOD);
    ENUM(sample_method_t, DPMPP2M_SAMPLE_METHOD);
    ENUM(sample_method_t, DPMPP2Mv2_SAMPLE_METHOD);
    ENUM(sample_method_t, IPNDM_SAMPLE_METHOD);
    ENUM(sample_method_t, IPNDM_V_SAMPLE_METHOD);
    ENUM(sample_method_t, LCM_SAMPLE_METHOD);
    ENUM(sample_method_t, DDIM_TRAILING_SAMPLE_METHOD);
    ENUM(sample_method_t, TCD_SAMPLE_METHOD);
    ENUM(sample_method_t, RES_MULTISTEP_SAMPLE_METHOD);
    ENUM(sample_method_t, RES_2S_SAMPLE_METHOD);
    ENUM(sample_method_t, ER_SDE_SAMPLE_METHOD);
    ENUM(sample_method_t, EULER_CFG_PP_SAMPLE_METHOD);
    ENUM(sample_method_t, EULER_A_CFG_PP_SAMPLE_METHOD);
    ENUM(sample_method_t, EULER_GE_SAMPLE_METHOD);
    ENUM(sample_method_t, DPMPP2M_SDE_SAMPLE_METHOD);
    ENUM(sample_method_t, DPMPP2M_SDE_BT_SAMPLE_METHOD);
    ENUM(sample_method_t, LMS_SAMPLE_METHOD);
    ENUM(sample_method_t, SAMPLE_METHOD_COUNT);

    ENUM(scheduler_t, DISCRETE_SCHEDULER);
    ENUM(scheduler_t, KARRAS_SCHEDULER);
    ENUM(scheduler_t, EXPONENTIAL_SCHEDULER);
    ENUM(scheduler_t, AYS_SCHEDULER);
    ENUM(scheduler_t, GITS_SCHEDULER);
    ENUM(scheduler_t, SGM_UNIFORM_SCHEDULER);
    ENUM(scheduler_t, SIMPLE_SCHEDULER);
    ENUM(scheduler_t, SMOOTHSTEP_SCHEDULER);
    ENUM(scheduler_t, KL_OPTIMAL_SCHEDULER);
    ENUM(scheduler_t, LCM_SCHEDULER);
    ENUM(scheduler_t, BONG_TANGENT_SCHEDULER);
    ENUM(scheduler_t, LTX2_SCHEDULER);
    ENUM(scheduler_t, LOGIT_NORMAL_SCHEDULER);
    ENUM(scheduler_t, FLUX2_SCHEDULER);
    ENUM(scheduler_t, FLUX_SCHEDULER);
    ENUM(scheduler_t, BETA_SCHEDULER);
    ENUM(scheduler_t, LLADA_IMAGE_SCHEDULER);
    ENUM(scheduler_t, SCHEDULER_COUNT);

    ENUM(prediction_t, EPS_PRED);
    ENUM(prediction_t, V_PRED);
    ENUM(prediction_t, EDM_V_PRED);
    ENUM(prediction_t, FLOW_PRED);
    ENUM(prediction_t, FLUX_FLOW_PRED);
    ENUM(prediction_t, SEFI_FLOW_PRED);
    ENUM(prediction_t, MINIT2I_FLOW_PRED);
    ENUM(prediction_t, SENSENOVA_U1_FLOW_PRED);
    ENUM(prediction_t, FLUX3_FLOW_PRED);
    ENUM(prediction_t, H3_AV_FLOW_PRED);
    ENUM(prediction_t, MING_IMAGE_FLOW_PRED);
    ENUM(prediction_t, PREDICTION_COUNT);

    ENUM(sd_type_t, SD_TYPE_F32);
    ENUM(sd_type_t, SD_TYPE_F16);
    ENUM(sd_type_t, SD_TYPE_Q4_0);
    ENUM(sd_type_t, SD_TYPE_Q4_1);
    ENUM(sd_type_t, SD_TYPE_Q5_0);
    ENUM(sd_type_t, SD_TYPE_Q5_1);
    ENUM(sd_type_t, SD_TYPE_Q8_0);
    ENUM(sd_type_t, SD_TYPE_Q8_1);
    ENUM(sd_type_t, SD_TYPE_Q2_K);
    ENUM(sd_type_t, SD_TYPE_Q3_K);
    ENUM(sd_type_t, SD_TYPE_Q4_K);
    ENUM(sd_type_t, SD_TYPE_Q5_K);
    ENUM(sd_type_t, SD_TYPE_Q6_K);
    ENUM(sd_type_t, SD_TYPE_Q8_K);
    ENUM(sd_type_t, SD_TYPE_IQ2_XXS);
    ENUM(sd_type_t, SD_TYPE_IQ2_XS);
    ENUM(sd_type_t, SD_TYPE_IQ3_XXS);
    ENUM(sd_type_t, SD_TYPE_IQ1_S);
    ENUM(sd_type_t, SD_TYPE_IQ4_NL);
    ENUM(sd_type_t, SD_TYPE_IQ3_S);
    ENUM(sd_type_t, SD_TYPE_IQ2_S);
    ENUM(sd_type_t, SD_TYPE_IQ4_XS);
    ENUM(sd_type_t, SD_TYPE_I8);
    ENUM(sd_type_t, SD_TYPE_I16);
    ENUM(sd_type_t, SD_TYPE_I32);
    ENUM(sd_type_t, SD_TYPE_I64);
    ENUM(sd_type_t, SD_TYPE_F64);
    ENUM(sd_type_t, SD_TYPE_IQ1_M);
    ENUM(sd_type_t, SD_TYPE_BF16);
    ENUM(sd_type_t, SD_TYPE_TQ1_0);
    ENUM(sd_type_t, SD_TYPE_TQ2_0);
    ENUM(sd_type_t, SD_TYPE_MXFP4);
    ENUM(sd_type_t, SD_TYPE_NVFP4);
    ENUM(sd_type_t, SD_TYPE_Q1_0);
    ENUM(sd_type_t, SD_TYPE_Q2_0);
    ENUM(sd_type_t, SD_TYPE_F8_E4M3);
    ENUM(sd_type_t, SD_TYPE_F8_E5M2);
    ENUM(sd_type_t, SD_TYPE_COUNT);

    ENUM(sd_log_level_t, SD_LOG_DEBUG);
    ENUM(sd_log_level_t, SD_LOG_VERBOSE);
    ENUM(sd_log_level_t, SD_LOG_INFO);
    ENUM(sd_log_level_t, SD_LOG_WARN);
    ENUM(sd_log_level_t, SD_LOG_ERROR);

    ENUM(preview_t, PREVIEW_NONE);
    ENUM(preview_t, PREVIEW_PROJ);
    ENUM(preview_t, PREVIEW_TAE);
    ENUM(preview_t, PREVIEW_VAE);
    ENUM(preview_t, PREVIEW_COUNT);

    ENUM(lora_apply_mode_t, LORA_APPLY_AUTO);
    ENUM(lora_apply_mode_t, LORA_APPLY_IMMEDIATELY);
    ENUM(lora_apply_mode_t, LORA_APPLY_AT_RUNTIME);
    ENUM(lora_apply_mode_t, LORA_APPLY_MODE_COUNT);

    ENUM(sd_vae_format_t, SD_VAE_FORMAT_AUTO);
    ENUM(sd_vae_format_t, SD_VAE_FORMAT_FLUX);
    ENUM(sd_vae_format_t, SD_VAE_FORMAT_SD3);
    ENUM(sd_vae_format_t, SD_VAE_FORMAT_FLUX2);
    ENUM(sd_vae_format_t, SD_VAE_FORMAT_WAN);
    ENUM(sd_vae_format_t, SD_VAE_FORMAT_COUNT);

    ENUM(sd_cache_mode_t, SD_CACHE_DISABLED);
    ENUM(sd_cache_mode_t, SD_CACHE_EASYCACHE);
    ENUM(sd_cache_mode_t, SD_CACHE_UCACHE);
    ENUM(sd_cache_mode_t, SD_CACHE_DBCACHE);
    ENUM(sd_cache_mode_t, SD_CACHE_TAYLORSEER);
    ENUM(sd_cache_mode_t, SD_CACHE_CACHE_DIT);
    ENUM(sd_cache_mode_t, SD_CACHE_SPECTRUM);

    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_NONE);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT_NEAREST);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT_NEAREST_EXACT);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT_ANTIALIASED);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT_BICUBIC);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LATENT_BICUBIC_ANTIALIASED);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_LANCZOS);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_NEAREST);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_MODEL);
    ENUM(sd_hires_upscaler_t, SD_HIRES_UPSCALER_COUNT);

    ENUM(sd_cancel_mode_t, SD_CANCEL_ALL);
    ENUM(sd_cancel_mode_t, SD_CANCEL_NEW_LATENTS);
    ENUM(sd_cancel_mode_t, SD_CANCEL_RESET);
    return 0;
}
