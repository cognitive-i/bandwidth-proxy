package ioshaper_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cognitive-i/bandwidth-proxy/ioshaper"
)

var _ = Describe("Bitrate", func() {
	It("should be able to tell if 0!", func() {
		Expect(ioshaper.NewBitrate(0).IsZero()).To(BeTrue())
		Expect(ioshaper.NewBitrate(10).IsZero()).To(BeFalse())
	})

	It("should calculate byterate", func() {
		sut := ioshaper.NewBitrate(32)
		Expect(sut.ByteRate()).To(BeEquivalentTo(4))
		Expect(sut.TransferBits(96000)).To(Equal(3000 * time.Second))
		Expect(sut.TransferBytes(1000)).To(Equal(250 * time.Second))
	})

	It("should calculate durations for amounts", func() {
		sut := ioshaper.NewBitrate(1024)
		Expect(sut.DurationToBytes(250 * time.Millisecond)).To(BeEquivalentTo(32))
	})

	It("should panic for Transfer{Bits,Bytes} when 0", func() {
		sut := ioshaper.NewBitrate(0)
		Expect(func() { _ = sut.TransferBits(2000) }).To(PanicWith(errors.New("Bitrate.TransferBits invalid when bitrate = 0")))
		Expect(func() { _ = sut.TransferBytes(2000) }).To(PanicWith(errors.New("Bitrate.TransferBits invalid when bitrate = 0")))
	})
})
