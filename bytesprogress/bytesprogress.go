// Package bytesprogress provides a simple, generic io.Writer wrapper that
// periodically invokes a user-provided callback every N bytes written.
//
// It is particularly useful for progress reporting during long-running
// streaming operations (file uploads/downloads, encoding, compression, etc.)
// where you want updates at regular byte intervals rather than on every Write call.
//
// The package is deliberately minimal: it contains no logging, no UI code,
// and no assumptions about how progress should be displayed. All behavior
// is provided by the caller through a Callback.
package bytesprogress

import "io"

// Ensure Writer satisfies io.Writer at compile time.
var _ io.Writer = (*Writer)(nil)

// Callback is invoked whenever the configured byte step is reached or exceeded.
//
// total is the total number of bytes written to the Writer so far. It is
// monotonically increasing and always reflects the exact cumulative total
// at the moment the callback is fired.
type Callback func(total int64)

// Writer wraps an io.Writer-like use case by calling Callback every `Step` bytes.
//
// Key behaviors:
//   - Every byte written contributes to the total, even if a single Write
//     crosses the step boundary multiple times.
//   - The callback is invoked **at most once per Write call**, even for very
//     large writes. This prevents callback storms and keeps performance predictable.
//   - If Step is 0 or negative, callbacks are completely disabled (but Total()
//     continues to be updated correctly).
//   - Writer is **not** safe for concurrent use by multiple goroutines.
//     Wrap it with sync.Mutex if you need concurrent writes.
type Writer struct {
	step     int64    // byte interval between callbacks
	callback Callback // called when pending >= step

	total   int64 // total bytes written so far
	pending int64 // bytes written since last callback
}

// New returns a new Writer that will invoke cb every `step` bytes.
//
// To receive an initial callback with total = 0, call the callback manually
// before starting to write:
//
//	progress := bytesprogress.New(1024*1024, myCallback)
//	myCallback(0)           // optional: show 0% immediately
func New(step int64, cb Callback) *Writer {
	return &Writer{
		step:     step,
		callback: cb,
	}
}

// Write writes len(p) bytes and updates the internal progress counters.
// It always returns n = len(p), nil — it never fails.
//
// The callback is triggered at most once per Write call, after which
// `pending` is reduced modulo `step`.
func (w *Writer) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}

	w.total += int64(n)
	w.pending += int64(n)

	if w.step > 0 && w.callback != nil && w.pending >= w.step {
		w.pending %= w.step
		w.callback(w.total)
	}

	return n, nil
}

// Total returns the total number of bytes written to this Writer so far.
func (w *Writer) Total() int64 { return w.total }
