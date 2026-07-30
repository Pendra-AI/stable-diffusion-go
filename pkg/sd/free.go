package sd

import "unsafe"

// cFree holds the C runtime free(3) function once bound by Load (best-effort,
// via bindCFree). When it is nil — because the native library has not been
// loaded yet or the C runtime could not be resolved — FreeImage and FreeImages
// are no-ops, so they are always safe to call.
var cFree func(unsafe.Pointer)

// FreeImage frees a single SDImage returned by the native library: its pixel
// buffer and the struct allocation itself. It is safe to call with a nil image
// or when the free binding is unavailable (in which case it does nothing).
func FreeImage(img *SDImage) {
	FreeImages(img, 1)
}

// FreeImages frees an array of count SDImages returned by the native library
// (for example the result of generate_image, which mallocs an array of
// images, each owning a malloc'd pixel buffer). It frees every image's pixel
// buffer and then the backing array allocation. It is safe to call with a nil
// pointer, a non-positive count, or when the free binding is unavailable (in
// which case it does nothing).
//
// When the library exports free_sd_images (added upstream in the master-802
// resync) that is used, so the memory is released by the same C runtime that
// allocated it — the safe path on Windows, where the library and the process
// may link different CRTs. The manual free(3) walk is kept only as a fallback
// for the pathological case where the symbol failed to bind.
func FreeImages(imgs *SDImage, count int) {
	if imgs == nil || count <= 0 {
		return
	}

	if freeSDImages != nil {
		freeSDImages(imgs, int32(count))
		return
	}

	if cFree == nil {
		return
	}

	base := unsafe.Pointer(imgs)
	stride := unsafe.Sizeof(SDImage{})
	for i := 0; i < count; i++ {
		img := (*SDImage)(unsafe.Add(base, uintptr(i)*stride))
		if img.Data != nil {
			cFree(unsafe.Pointer(img.Data))
			img.Data = nil
		}
	}
	cFree(base)
}
