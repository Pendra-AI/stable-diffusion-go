//go:build darwin || linux

package sd

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// nativeAlloc returns libc's calloc. On these platforms the library allocates
// with the same malloc, so its free functions accept what this returns.
func nativeAlloc(t *testing.T) func(uintptr) unsafe.Pointer {
	t.Helper()
	handle, err := purego.Dlopen(libcCandidates()[0], purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("no libc handle: %v", err)
	}
	var calloc func(count, size uintptr) unsafe.Pointer
	purego.RegisterLibFunc(&calloc, handle, "calloc")
	return func(size uintptr) unsafe.Pointer { return calloc(1, size) }
}

func nativeString(alloc func(uintptr) unsafe.Pointer, s string) *uint8 {
	p := alloc(uintptr(len(s)) + 1)
	copy(unsafe.Slice((*byte)(p), len(s)), s)
	return (*uint8)(p)
}

// The library has to walk a result to free it, reading each item's data and
// generation_info. A result built with the Go mirrors and released by the real
// function crashes the process in free(3) if the two disagree on where those are.
func TestNativeFreeImageGenerationResult(t *testing.T) {
	loadNativeLibrary(t)
	alloc := nativeAlloc(t)

	const count = 3
	items := unsafe.Slice((*SDImageResult)(alloc(count*unsafe.Sizeof(SDImageResult{}))), count)
	for i := range items {
		items[i].Image = SDImage{Width: 1, Height: 1, Channel: 4, Data: (*uint8)(alloc(4))}
		items[i].GenerationInfo = nativeString(alloc, `{"schema":"sdcpp.generation/v1"}`)
	}
	raw := (*cImageGenerationResult)(alloc(unsafe.Sizeof(cImageGenerationResult{})))
	raw.Items, raw.Count = &items[0], count

	res := newImageGenerationResult(true, raw)
	if res.Count() != count || res.Image(2) == nil || res.GenerationInfo(2) != `{"schema":"sdcpp.generation/v1"}` {
		t.Fatalf("Count=%d Image(2)=%v GenerationInfo(2)=%q", res.Count(), res.Image(2), res.GenerationInfo(2))
	}
	res.Free()
	res.Free()
}

func TestNativeFreeVideoGenerationResult(t *testing.T) {
	loadNativeLibrary(t)
	alloc := nativeAlloc(t)

	const count = 2
	frames := unsafe.Slice((*SDImage)(alloc(count*unsafe.Sizeof(SDImage{}))), count)
	for i := range frames {
		frames[i] = SDImage{Width: 1, Height: 1, Channel: 3, Data: (*uint8)(alloc(3))}
	}
	audio := (*SDAudio)(alloc(unsafe.Sizeof(SDAudio{})))
	audio.SampleRate, audio.Channels, audio.SampleCount = 48000, 2, 4
	audio.Data = (*float32)(alloc(16))

	raw := (*cVideoGenerationResult)(alloc(unsafe.Sizeof(cVideoGenerationResult{})))
	raw.Frames, raw.FrameCount, raw.Audio, raw.Fps = &frames[0], count, audio, 16
	raw.GenerationInfo = nativeString(alloc, `{}`)

	res := newVideoGenerationResult(true, raw)
	if res.FrameCount() != count || res.Fps() != 16 || res.Audio().SampleRate != 48000 {
		t.Fatalf("FrameCount=%d Fps=%d Audio=%+v", res.FrameCount(), res.Fps(), res.Audio())
	}
	res.Free()
	res.Free()
}
