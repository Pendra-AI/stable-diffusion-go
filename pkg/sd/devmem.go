package sd

import "unsafe"

// devMemAvailable is set true once the optional sd_backend_dev_* device-memory
// wrappers register on Load. They ship in libstable-diffusion v0.3.0+; an older
// lib (e.g. a PENDRA_SD_LIB override pointing at an earlier build) simply lacks
// them, so GpuDevices reports the reading unavailable rather than failing Load
// or image generation.
var devMemAvailable bool

// Device-memory wrapper function pointers, bound by registerDevMemFunctions
// (in load.go, so symbols_test.go's load.go↔expected-symbols reconciliation
// sees them). The device handle is passed as an opaque unsafe.Pointer, matching
// how the binding already carries the sd_ctx handle.
var (
	sdBackendDevCount  func() uint64
	sdBackendDevGet    func(index uint64) unsafe.Pointer
	sdBackendDevMemory func(dev unsafe.Pointer, free *uint64, total *uint64)
	sdBackendDevName   func(dev unsafe.Pointer) *uint8
	sdBackendDevType   func(dev unsafe.Pointer) int32
)

// GpuDeviceType classifies a ggml backend device (ggml_backend_dev_type).
type GpuDeviceType int

const (
	GpuDeviceCPU  GpuDeviceType = 0 // GGML_BACKEND_DEVICE_TYPE_CPU
	GpuDeviceGPU  GpuDeviceType = 1 // GGML_BACKEND_DEVICE_TYPE_GPU  — dedicated VRAM
	GpuDeviceIGPU GpuDeviceType = 2 // GGML_BACKEND_DEVICE_TYPE_IGPU — shares host RAM
)

// GpuDevice is one ggml backend device reported by the statically-linked ggml
// inside libstable-diffusion, with its live free/total memory.
type GpuDevice struct {
	Name       string        // "CUDA0", "Metal", "Vulkan0", "CPU", ...
	Type       GpuDeviceType // CPU / GPU / IGPU (see constants)
	FreeBytes  uint64        // live free memory (real VRAM on CUDA; working-set on Metal)
	TotalBytes uint64        // total memory
}

// GpuDevices enumerates the ggml backend devices and their live free/total
// memory. ok is false when the library isn't loaded, or the running
// libstable-diffusion predates the sd_backend_dev_* wrappers (v0.3.0+). The
// enumeration is panic-safe: a misbehaving FFI call yields (nil, false) rather
// than crashing the caller, mirroring Load's contract.
func GpuDevices() (devs []GpuDevice, ok bool) {
	if libSD == 0 || !devMemAvailable {
		return nil, false
	}
	defer func() {
		if r := recover(); r != nil {
			devs, ok = nil, false
		}
	}()
	n := sdBackendDevCount()
	out := make([]GpuDevice, 0, n)
	for i := uint64(0); i < n; i++ {
		dev := sdBackendDevGet(i)
		if dev == nil {
			continue
		}
		var free, total uint64
		sdBackendDevMemory(dev, &free, &total)
		out = append(out, GpuDevice{
			Name:       CGoString(sdBackendDevName(dev)),
			Type:       GpuDeviceType(sdBackendDevType(dev)),
			FreeBytes:  free,
			TotalBytes: total,
		})
	}
	return out, true
}
