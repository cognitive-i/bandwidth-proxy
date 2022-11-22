package ioshaper

import (
	"errors"
	"time"
)

// Use an embedded type as this stops accidental precision loss casting.
// Previously `type Bitrate int64` which meant `Bitrate(Math.MaxUint64)`
// would compile silently and give unexpected results

type Bitrate struct {
	int64
}

type bitrateCastFromType interface {
	~int | ~int16 | ~int32 | ~int64
}

func NewBitrate[T bitrateCastFromType](bitrate T) Bitrate {
	// hopefully `int` won't be `int128` anytime soon
	if b := int64(bitrate); b >= 0 {
		return Bitrate{b}
	}
	panic("bitrate must >= 0")
}

func (b Bitrate) ByteRate() int64 {
	return b.int64 / 8
}

func (b Bitrate) TransferBits(size int) time.Duration {
	if b.IsZero() {
		panic(errors.New("Bitrate.TransferBits invalid when bitrate = 0"))
	}

	return time.Duration(size) * time.Second / time.Duration(b.int64)
}

func (b Bitrate) TransferBytes(size int) time.Duration {
	return b.TransferBits(8 * size)
}

func (b Bitrate) DurationToBits(duration time.Duration) int64 {
	// work in nanoseconds to prevent integer division precision loss in seconds.
	return (b.int64 * duration.Nanoseconds()) / time.Second.Nanoseconds()
}

func (b Bitrate) DurationToBytes(duration time.Duration) int64 {
	bits := b.DurationToBits(duration)
	return bits / 8
}

func (b Bitrate) IsZero() bool {
	return b.int64 == 0
}
