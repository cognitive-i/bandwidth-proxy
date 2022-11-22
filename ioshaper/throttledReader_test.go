package ioshaper_test

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/cognitive-i/bandwidth-proxy/ioshaper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

func getBuffer(size int) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(make([]byte, size)))
}

var _ = Describe("ThrottledReader", func() {
	DescribeTable("ThrottledReader",
		func(bitrate int, bufferSize int, expectedDuration time.Duration) {
			tb := gmeasure.NewExperiment(fmt.Sprintf("%d bytes @ %d bps", bufferSize, bitrate))
			AddReportEntry(tb.Name, tb)

			tb.Sample(
				func(idx int) {
					source := getBuffer(bufferSize)
					sut := ioshaper.NewThrottledReader(source, int64(bitrate))

					tb.MeasureDuration("download", func() {
						_, err := io.ReadAll(sut)
						Expect(err).To(Succeed())
					})
				},
				gmeasure.SamplingConfig{
					N:           5,
					Duration:    10 * time.Second,
					NumParallel: 1,
				},
			)

			meanDuration := tb.GetStats("download").DurationFor(gmeasure.StatMean)
			Expect(meanDuration).To(BeNumerically("~", expectedDuration, expectedDuration/10))
		},
		Entry("slow", 1024, 256, 2*time.Second),
		Entry("medium", 1024*1024, 256*1024, 2*time.Second),
		Entry("fast", 10*1024*1024, 1024*1024, 800*time.Millisecond),
	)

	It("should ensure a single transfer is shorter than 250ms", func() {
		// this is to keep flow going for low bitrates
		tb := gmeasure.NewExperiment("low bitrate throttling")
		AddReportEntry(tb.Name, tb)

		tb.Sample(
			func(idx int) {
				source := getBuffer(16384)
				sut := ioshaper.NewThrottledReader(source, 1024)

				tb.MeasureDuration("download", func() {
					for total := 0; total < 2048; {
						tb.MeasureDuration("individual", func() {
							n, err := sut.Read(make([]byte, 512))
							Expect(err).To(Succeed())
							total += n
						})
					}
				})
			},
			gmeasure.SamplingConfig{
				N:        5,
				Duration: 20 * time.Second,
			},
		)

		meanDownload := tb.GetStats("download").DurationFor(gmeasure.StatMean)
		Expect(meanDownload).To(BeNumerically("~", 16*time.Second, 1*time.Second))

		individualDuration := tb.GetStats("individual").DurationFor(gmeasure.StatMax)
		Expect(individualDuration).To(BeNumerically("<", 250*time.Millisecond))
	})
})
