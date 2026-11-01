package sd

import "unsafe"

// SDImageGenerationResult owns the native result of GenerateImage or Upscale: the
// pixels and generation info of every image. They stay in native memory until
// Free, independently of the context that produced them.
//
// All methods are safe on a nil result and after Free. A result must not be used
// from several goroutines at once.
type SDImageGenerationResult struct {
	ptr *cImageGenerationResult
}

// newImageGenerationResult takes ownership of raw. It returns nil, releasing raw,
// when the call failed or the result holds no images, so a non-nil result always
// has at least one item.
func newImageGenerationResult(ok bool, raw *cImageGenerationResult) *SDImageGenerationResult {
	if raw == nil {
		return nil
	}
	if !ok || raw.Items == nil || raw.Count <= 0 {
		freeImageGenerationResult(raw)
		return nil
	}
	return &SDImageGenerationResult{ptr: raw}
}

// freeImageGenerationResult releases raw through the library, the only allocator
// that may free it. Its items are not an SDImage array, so free_sd_images must
// never be used on them.
func freeImageGenerationResult(raw *cImageGenerationResult) {
	if freeSDImageGenerationResult != nil {
		freeSDImageGenerationResult(raw)
	}
}

// Count returns the number of images in the result.
func (r *SDImageGenerationResult) Count() int {
	if r == nil || r.ptr == nil || r.ptr.Items == nil || r.ptr.Count <= 0 {
		return 0
	}
	return int(r.ptr.Count)
}

func (r *SDImageGenerationResult) item(i int) *SDImageResult {
	if i < 0 || i >= r.Count() {
		return nil
	}
	return (*SDImageResult)(unsafe.Add(unsafe.Pointer(r.ptr.Items), uintptr(i)*unsafe.Sizeof(SDImageResult{})))
}

// Image returns the i-th image, or nil when i is out of range or the slot has no
// pixels. The image points into native memory and is valid until Free.
func (r *SDImageGenerationResult) Image(i int) *SDImage {
	item := r.item(i)
	if item == nil || item.Image.Data == nil {
		return nil
	}
	return &item.Image
}

// GenerationInfo returns a copy of the i-th image's JSON metadata (schema
// sdcpp.generation/v1), or "" when i is out of range.
func (r *SDImageGenerationResult) GenerationInfo(i int) string {
	item := r.item(i)
	if item == nil {
		return ""
	}
	return CGoString(item.GenerationInfo)
}

// Free releases the result and everything it owns, including images obtained from
// Image. It is safe to call more than once.
func (r *SDImageGenerationResult) Free() {
	if r == nil || r.ptr == nil {
		return
	}
	raw := r.ptr
	r.ptr = nil
	freeImageGenerationResult(raw)
}

// SDVideoGenerationResult owns the native result of GenerateVideo: frames, audio
// and generation info. The ownership rules of SDImageGenerationResult apply.
type SDVideoGenerationResult struct {
	ptr *cVideoGenerationResult
}

// newVideoGenerationResult takes ownership of raw. It returns nil, releasing raw,
// when the call failed or the result holds no frames.
func newVideoGenerationResult(ok bool, raw *cVideoGenerationResult) *SDVideoGenerationResult {
	if raw == nil {
		return nil
	}
	if !ok || raw.Frames == nil || raw.FrameCount <= 0 {
		freeVideoGenerationResult(raw)
		return nil
	}
	return &SDVideoGenerationResult{ptr: raw}
}

func freeVideoGenerationResult(raw *cVideoGenerationResult) {
	if freeSDVideoGenerationResult != nil {
		freeSDVideoGenerationResult(raw)
	}
}

// FrameCount returns the number of frames in the result.
func (r *SDVideoGenerationResult) FrameCount() int {
	if r == nil || r.ptr == nil || r.ptr.Frames == nil || r.ptr.FrameCount <= 0 {
		return 0
	}
	return int(r.ptr.FrameCount)
}

// Frame returns the i-th frame, or nil when i is out of range or the slot has no
// pixels. The frame points into native memory and is valid until Free.
func (r *SDVideoGenerationResult) Frame(i int) *SDImage {
	if i < 0 || i >= r.FrameCount() {
		return nil
	}
	frame := (*SDImage)(unsafe.Add(unsafe.Pointer(r.ptr.Frames), uintptr(i)*unsafe.Sizeof(SDImage{})))
	if frame.Data == nil {
		return nil
	}
	return frame
}

// Frames returns the frames as a slice. The slice itself is a Go copy, but each
// element's pixels stay in native memory and are valid until Free.
func (r *SDVideoGenerationResult) Frames() []SDImage {
	n := r.FrameCount()
	if n == 0 {
		return nil
	}
	frames := make([]SDImage, n)
	copy(frames, unsafe.Slice(r.ptr.Frames, n))
	return frames
}

// Audio returns the generated audio track, or nil when the model produced none.
// It is valid until Free and must not be passed to FreeAudio.
func (r *SDVideoGenerationResult) Audio() *SDAudio {
	if r == nil || r.ptr == nil {
		return nil
	}
	return r.ptr.Audio
}

// Fps returns the frame rate of the generated video, or 0 after Free.
func (r *SDVideoGenerationResult) Fps() int {
	if r == nil || r.ptr == nil {
		return 0
	}
	return int(r.ptr.Fps)
}

// GenerationInfo returns a copy of the video's JSON metadata (schema
// sdcpp.generation/v1).
func (r *SDVideoGenerationResult) GenerationInfo() string {
	if r == nil || r.ptr == nil {
		return ""
	}
	return CGoString(r.ptr.GenerationInfo)
}

// Free releases the result and everything it owns, including frames and audio
// obtained from it. It is safe to call more than once.
func (r *SDVideoGenerationResult) Free() {
	if r == nil || r.ptr == nil {
		return
	}
	raw := r.ptr
	r.ptr = nil
	freeVideoGenerationResult(raw)
}
