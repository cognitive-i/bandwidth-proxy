package ioshaper

import (
	"errors"
	"time"
)

type Bitrate uint64

func (b Bitrate) ByteRate() int {
	return int(b) / 8
}

func (b Bitrate) TransferBits(size int) time.Duration {
	if b.IsZero() {
		panic(errors.New("Bitrate.TransferBits invalid when bitrate = 0"))
	}

	return time.Duration(size) * time.Second / time.Duration(b)
}

func (b Bitrate) TransferBytes(size int) time.Duration {
	return b.TransferBits(8 * size)
}

func (b Bitrate) DurationToBits(duration time.Duration) int {
	return int((int64(b) * int64(duration.Nanoseconds())) / int64(time.Second))
}

func (b Bitrate) DurationToBytes(duration time.Duration) int {
	bits := b.DurationToBits(duration)
	return bits / 8
}

func (b Bitrate) IsZero() bool {
	return b == 0
}
