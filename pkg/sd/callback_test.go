package sd

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// TestCallbackSignaturesAcceptedByPurego is the regression guard for the
// Windows image-generation crash:
//
//	stablediffusion generate: handler panic: compileCallback:
//	expected function with one uintptr-sized result
//
// A Go func becomes a machine callback through purego.NewCallback, which on
// Windows delegates to syscall.NewCallback -> runtime.compileCallback. That path REQUIRES the callback return exactly one
// uintptr-sized value; a void Go callback panics. The log/progress/preview
// callback types (and their wrapper closures in SetLogCallback /
// SetProgressCallback / SetPreviewCallback) therefore return uintptr.
//
// The log callback is the one the in-process worker actually registers (on the
// first generate), so it is the production-critical path and must wrap on every
// platform — including Windows, where the bug bit. The preview callback has no
// float in its signature and must also wrap everywhere.
//
// No native library is needed: the panic (if any) happens at callback
// compilation, not at the C call.
func TestCallbackSignaturesAcceptedByPurego(t *testing.T) {
	mustWrap := func(t *testing.T, name string, fn any) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("purego.NewCallback(%s) panicked (void-callback regression): %v", name, r)
			}
		}()
		if got := purego.NewCallback(fn); got == 0 {
			t.Fatalf("purego.NewCallback(%s) returned a 0 handle", name)
		}
	}

	var logCB SDLogCallback = func(level SDLogLevel, text *uint8, data unsafe.Pointer) uintptr {
		return 0
	}
	var previewCB SDPreviewCallback = func(step, frameCount int32, frames *SDImage, isNoisy bool, data unsafe.Pointer) uintptr {
		return 0
	}

	mustWrap(t, "SDLogCallback", logCB)
	mustWrap(t, "SDPreviewCallback", previewCB)
}

// The log callback's level type must take the SDLog* constants, whose integers
// moved at master-1031, so callers never need a literal.
var _ SDLogLevelType = SDLogVerbose

// Passing nil clears a callback. purego panics when handed a nil func, so the
// setters must reach the library with a 0 callback instead.
func TestClearingCallbacksPassesNull(t *testing.T) {
	saveLog, saveProgress, savePreview := sdSetLogCallback, sdSetProgressCallback, sdSetPreviewCallback
	t.Cleanup(func() {
		sdSetLogCallback, sdSetProgressCallback, sdSetPreviewCallback = saveLog, saveProgress, savePreview
	})

	var calls int
	expectNull := func(name string, cb uintptr) {
		calls++
		if cb != 0 {
			t.Errorf("%s: cb = %#x, want 0", name, cb)
		}
	}
	sdSetLogCallback = func(cb uintptr, _ unsafe.Pointer) { expectNull("log", cb) }
	sdSetProgressCallback = func(cb uintptr, _ unsafe.Pointer) { expectNull("progress", cb) }
	sdSetPreviewCallback = func(cb uintptr, _ Preview, _ int32, _, _ bool, _ unsafe.Pointer) { expectNull("preview", cb) }

	SetLogCallback(nil, nil)
	SetProgressCallback(nil, nil)
	SetPreviewCallback(nil, PreviewNone, 0, false, false, nil)
	if calls != 3 {
		t.Errorf("setters called %d times, want 3", calls)
	}
}

// TestSetProgressCallbackNoPanicOnWindows verifies the graceful handling of a
// SEPARATE, pre-existing limitation surfaced by the work above: SDProgressCallback
// takes a `float32` argument, and Windows' syscall.NewCallback (purego's backend
// there) supports neither float arguments nor float returns. Rather than leave a
// latent panic ("float arguments not supported"), SetProgressCallback skips
// registration on Windows. This test asserts that registering a progress
// callback on Windows does NOT panic. It runs only on Windows; the no-op path
// returns before touching the (lib-loaded) purego symbol, so no native library
// is required.
func TestSetProgressCallbackNoPanicOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific: progress-callback registration is skipped on Windows")
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetProgressCallback panicked on Windows; expected a graceful no-op: %v", r)
		}
	}()
	SetProgressCallback(func(step, steps int, tm float32, data interface{}) {}, nil)
}

// TestProgressCallbackWrapsOffWindows confirms that off Windows (where purego
// supports float callback args) the progress callback type — with its uintptr
// return — still wraps cleanly via purego.NewCallback.
func TestProgressCallbackWrapsOffWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("float-arg callbacks aren't wrappable on Windows; see TestSetProgressCallbackNoPanicOnWindows")
	}
	var progressCB SDProgressCallback = func(step, steps int32, tm float32, data unsafe.Pointer) uintptr {
		return 0
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("purego.NewCallback(SDProgressCallback) panicked off-Windows: %v", r)
		}
	}()
	if got := purego.NewCallback(progressCB); got == 0 {
		t.Fatal("purego.NewCallback(SDProgressCallback) returned a 0 handle")
	}
}
