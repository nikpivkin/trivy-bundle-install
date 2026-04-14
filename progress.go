package main

import (
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

type progressTracker struct{}

func (p *progressTracker) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	bar := progressbar.NewOptions64(totalSize,
		progressbar.OptionSetDescription(src),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(func() { fmt.Fprintln(os.Stderr) }),
	)
	bar.Add64(currentSize)
	r := progressbar.NewReader(stream, bar)
	return &r
}
