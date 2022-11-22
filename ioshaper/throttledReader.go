package ioshaper

import (
	"errors"
	"io"
	"log"
	"math"
	"time"
)

type throttledReader struct {
	source        io.ReadCloser
	bitrate       Bitrate
	maxBufferSize int
}

func NewThrottledReader(source io.ReadCloser, bitrate int64) io.ReadCloser {
	br := NewBitrate(bitrate)

	if br.IsZero() {
		log.Fatal("ThrottledReader cannot be configured with 0 bps bitrate")
	}

	// This is best effort ensure a trickle of bytes at least every
	// 200 ms in situation that caller is using a huge read buffer
	// and specifies very low bitrate
	maxBufferSize := br.DurationToBytes(200 * time.Millisecond)
	if maxBufferSize == 0 {
		maxBufferSize = 1
	}

	if maxBufferSize > math.MaxInt {
		panic(errors.New("maxBufferSize cannot be used for slice allocations"))
	}

	return &throttledReader{
		source:        source,
		bitrate:       br,
		maxBufferSize: int(maxBufferSize),
	}
}

func (r *throttledReader) Read(p []byte) (n int, err error) {
	if r.maxBufferSize < len(p) {
		p = p[:r.maxBufferSize]
	}

	startRead := time.Now()
	n, err = r.source.Read(p)

	if n > 0 {
		plannedDuration := r.bitrate.TransferBytes(n)

		elapsed := time.Since(startRead)
		stillToWait := plannedDuration - elapsed
		time.Sleep(stillToWait)
	}

	return n, err
}

func (r *throttledReader) Close() error {
	return r.source.Close()
}
