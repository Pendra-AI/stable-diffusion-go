// sd_devmem.cpp — Pendra addition to github.com/pendra-ai/stable-diffusion-go
//
// Whitelisted, sd_*-prefixed wrappers around ggml's device-memory registry, so a
// host that only dlopens libstable-diffusion (which statically links ggml with
// hidden visibility) can still read per-device free/total VRAM in-process — the
// same measurement llama.cpp/whisper.cpp already expose through their own libggml.
//
// Without these, a worker measuring a stable-diffusion model's footprint on a
// discrete GPU has no per-device VRAM reading, falls back to a process-RSS delta
// that misses dedicated VRAM, and reports ~0.1 GB — so the model's real weights
// land un-attributed in the console memory map's "Other" bucket (Pendra #1489).
//
// This file is injected into a freshly cloned upstream tree by
// scripts/inject-devmem-wrapper.sh and compiled into the `stable-diffusion`
// target via its `src/*.cpp` CONFIGURE_DEPENDS glob (no CMakeLists edit needed).
// SD_API gives these functions default visibility under the build's
// hidden-visibility preset (dllexport on Windows); the fail-closed symbol gate
// (scripts/check-symbols.sh) REQUIRES them and still forbids any exported ggml_*
// symbol, so ggml stays statically linked and hidden — no base-name collision
// with another in-process ggml-bearing backend.

#include "stable-diffusion.h"
#include "ggml-backend.h"

#ifdef __cplusplus
extern "C" {
#endif

// Number of ggml backend devices. The first call initialises ggml's static
// backend registry (GGML_BACKEND_DL=OFF), exactly as the first llama.cpp device
// query does — no separate init is required before enumerating.
SD_API size_t sd_backend_dev_count(void) {
    return ggml_backend_dev_count();
}

// Opaque handle for the i-th ggml backend device, or NULL when i is out of range.
SD_API ggml_backend_dev_t sd_backend_dev_get(size_t i) {
    return ggml_backend_dev_get(i);
}

// Live free/total device memory in bytes. On CUDA these are real VRAM figures;
// on Metal ggml reports the working-set. Either out pointer may be NULL — ggml's
// get_memory implementations dereference both unconditionally, so we read into
// locals here and copy out only what the caller asked for.
SD_API void sd_backend_dev_memory(ggml_backend_dev_t dev, size_t* free, size_t* total) {
    size_t f = 0, t = 0;
    ggml_backend_dev_memory(dev, &f, &t);
    if (free) {
        *free = f;
    }
    if (total) {
        *total = t;
    }
}

// Backend-specific device name, e.g. "CUDA0", "Metal", "Vulkan0", "CPU".
SD_API const char* sd_backend_dev_name(ggml_backend_dev_t dev) {
    return ggml_backend_dev_name(dev);
}

// ggml_backend_dev_type as an int: 0=CPU, 1=GPU, 2=IGPU, 3=ACCEL, 4=META.
SD_API int sd_backend_dev_type(ggml_backend_dev_t dev) {
    return (int) ggml_backend_dev_type(dev);
}

#ifdef __cplusplus
}
#endif
