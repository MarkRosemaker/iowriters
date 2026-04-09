package bytesprogress

import (
	"bytes"
	"io"
	"slices"
	"testing"
)

func TestWriter_FiresCallbackEveryStep(t *testing.T) {
	t.Parallel()

	var calls []int64
	w := New(10, func(total int64) { calls = append(calls, total) })

	// Five writes of 4 bytes each -> totals 4,8,12,16,20.
	// Callback should fire once the pending crosses 10, and again after 20.
	for range 5 {
		if n, err := w.Write([]byte("abcd")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if n != 4 {
			t.Errorf("Write returned %d, want 4", n)
		}
	}

	if total := w.Total(); total != 20 {
		t.Errorf("Total() = %d, want 20", total)
	}

	if got, want := calls, []int64{12, 20}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

func TestWriter_LargeWriteFiresOnce(t *testing.T) {
	t.Parallel()

	var calls int
	w := New(10, func(int64) { calls++ })

	// A single 100-byte write crosses the step boundary many times,
	// but the callback should only fire once.

	if n, err := w.Write(make([]byte, 100)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 100 {
		t.Errorf("Write returned %d, want 100", n)
	}

	if calls != 1 {
		t.Errorf("callback called %d times, want 1", calls)
	}

	if total := w.Total(); total != 100 {
		t.Errorf("Total() = %d, want 100", total)
	}
}

func TestWriter_ZeroStepDisablesCallback(t *testing.T) {
	t.Parallel()

	called := false
	w := New(0, func(int64) { called = true })

	if n, err := w.Write(make([]byte, 1000)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 1000 {
		t.Errorf("Write returned %d, want 1000", n)
	}

	if called {
		t.Error("callback called unexpectedly")
	}

	if total := w.Total(); total != 1000 {
		t.Errorf("Total() = %d, want 1000", total)
	}
}

func TestWriter_NegativeStepDisablesCallback(t *testing.T) {
	t.Parallel()

	called := false
	w := New(-1, func(int64) { called = true })

	if n, err := w.Write([]byte("hello")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}

	if called {
		t.Error("callback called unexpectedly")
	}
}

func TestWriter_NilCallbackSafe(t *testing.T) {
	t.Parallel()

	w := New(1, nil)

	if n, err := w.Write([]byte("hello")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}

	if total := w.Total(); total != 5 {
		t.Errorf("Total() = %d, want 5", total)
	}
}

func TestWriter_EmptyWrite(t *testing.T) {
	t.Parallel()

	var calls int
	w := New(10, func(int64) { calls++ })

	if n, err := w.Write(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 0 {
		t.Errorf("Write returned %d, want 0", n)
	}

	if calls != 0 {
		t.Errorf("callback called %d times, want 0", calls)
	}

	if total := w.Total(); total != 0 {
		t.Errorf("Total() = %d, want 0", total)
	}
}

func TestWriter_ExactStep(t *testing.T) {
	t.Parallel()

	var calls []int64
	w := New(10, func(total int64) { calls = append(calls, total) })

	if n, err := w.Write(make([]byte, 10)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 10 {
		t.Errorf("Write returned %d, want 10", n)
	}

	if got, want := calls, []int64{10}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}

	if w.pending != 0 {
		t.Errorf("pending should be 0 after exact step, got %d", w.pending)
	}

	// Write one more byte – should not fire callback again

	if n, err := w.Write([]byte("x")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 1 {
		t.Errorf("Write returned %d, want 1", n)
	}
	if got, want := calls, []int64{10}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}
	if total := w.Total(); total != 11 {
		t.Errorf("Total() = %d, want 11", total)
	}
}

func TestWriter_StepOne(t *testing.T) {
	t.Parallel()

	var calls []int64
	w := New(1, func(total int64) {
		calls = append(calls, total)
	})

	if n, err := w.Write([]byte("abcde")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}

	if got, want := calls, []int64{5}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}

	if total := w.Total(); total != 5 {
		t.Errorf("Total() = %d, want 5", total)
	}
}

func TestWriter_MultipleLargeWrites(t *testing.T) {
	t.Parallel()

	var calls int
	w := New(10, func(int64) { calls++ })

	// Two large writes, each crossing the step multiple times

	if n, err := w.Write(make([]byte, 100)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 100 {
		t.Errorf("Write returned %d, want 100", n)
	}

	if calls != 1 {
		t.Errorf("callback called %d times after first write, want 1", calls)
	}

	if n, err := w.Write(make([]byte, 100)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 100 {
		t.Errorf("Write returned %d, want 100", n)
	}
	if calls != 2 {
		t.Errorf("callback called %d times after second write, want 2", calls)
	}
	if total := w.Total(); total != 200 {
		t.Errorf("Total() = %d, want 200", total)
	}
}

func TestWriter_ModuloBehaviour(t *testing.T) {
	t.Parallel()

	var calls []int64
	w := New(10, func(total int64) {
		calls = append(calls, total)
	})

	// Write 25 bytes: should fire one callback at total=25, pending becomes 5
	if n, err := w.Write(make([]byte, 25)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 25 {
		t.Errorf("Write returned %d, want 25", n)
	}

	if got, want := calls, []int64{25}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}

	// Write 5 more bytes: pending 5+5=10 -> second callback at total=30
	if n, err := w.Write(make([]byte, 5)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}

	if got, want := calls, []int64{25, 30}; !slices.Equal(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}

	if total := w.Total(); total != 30 {
		t.Errorf("Total() = %d, want 30", total)
	}
}

func TestWriter_ComposesWithIoCopy(t *testing.T) {
	t.Parallel()

	// Use LimitedReader to prevent WriteTo optimisation.
	srcData := make([]byte, 1024)
	src := &io.LimitedReader{
		R: bytes.NewReader(srcData),
		N: int64(len(srcData)),
	}
	dst := &bytes.Buffer{}

	var totals []int64
	progress := New(256, func(total int64) { totals = append(totals, total) })

	// Use a buffer of 256 bytes so each Write is exactly the step size.
	buf := make([]byte, 256)
	if n, err := io.CopyBuffer(io.MultiWriter(dst, progress), src, buf); err != nil {
		t.Fatalf("io.CopyBuffer failed: %v", err)
	} else if n != 1024 {
		t.Errorf("io.Copy returned %d, want 1024", n)
	}

	if total := progress.Total(); total != 1024 {
		t.Errorf("progress.Total() = %d, want 1024", total)
	}

	if dst.Len() != 1024 {
		t.Errorf("dst buffer length = %d, want 1024", dst.Len())
	}

	if want := []int64{256, 512, 768, 1024}; !slices.Equal(totals, want) {
		t.Errorf("callback totals = %v, want %v", totals, want)
	}
}
