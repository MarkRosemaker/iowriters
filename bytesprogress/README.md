# bytesprogress

`bytesprogress` provides an [`io.Writer`](https://pkg.go.dev/io#Writer) that invokes a user‑defined callback every time a configurable number of bytes have been written.

It is ideal for progress reporting during long‑running streaming operations: file uploads/downloads, compression, encoding, or any I/O where you want periodic updates without being flooded by one callback per `Write` call.

## Features

- **Simple** – just a `Writer` struct, a `Callback` type, and a `New` constructor.
- **Efficient** – callback fires **at most once per `Write` call**, even if a single write crosses the step boundary many times.
- **Safe defaults** – zero or negative `step` disables callbacks; `nil` callback is ignored (no panic).
- **Monotonic totals** – the `total` passed to the callback is always increasing.
- **Zero dependencies** – uses only the standard library.

## Installation

```bash
go get github.com/MarkRosemaker/iowriters/bytesprogress
```

## Usage

### Basic example

```go
package main

import (
    "io"
    "log"
    "os"

    "github.com/MarkRosemaker/iowriters/bytesprogress"
)

func main() {
    src, _ := os.Open("largefile.dat")
    dst, _ := os.Create("copy.dat")
    defer src.Close()
    defer dst.Close()

    // Print progress every 1 MiB
    progress := bytesprogress.New(1024*1024, func(total int64) {
        log.Printf("%d MiB written", total/(1024*1024))
    })

    // Write through the progress writer
    _, err := io.Copy(io.MultiWriter(dst, progress), src)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Combining with other writers

Because `bytesprogress.Writer` implements `io.Writer`, you can combine it with others using `io.MultiWriter`:

```go
uploader := &myUploadWriter{}
hasher := sha256.New()
progress := bytesprogress.New(8192, func(total int64) {
    fmt.Printf("Uploaded %d bytes\n", total)
})

multi := io.MultiWriter(uploader, hasher, progress)
io.Copy(multi, someReader)
```

### Manual initial callback

If you want to show “0%” before any data is written, call your callback manually:

```go
cb := func(total int64) { fmt.Printf("progress: %d bytes\n", total) }
progress := bytesprogress.New(1024, cb)
cb(0) // prints "progress: 0 bytes"
io.Copy(progress, src)
```
