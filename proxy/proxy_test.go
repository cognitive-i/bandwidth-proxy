package proxy_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cognitive-i/bandwidth-proxy/ioshaper"
	"github.com/cognitive-i/bandwidth-proxy/proxy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

const testChunkSize = 10 * 1024 * 1024

var _ = Describe("Proxy", func() {
	var sutListener net.Listener
	var sutClient *http.Client

	var wwwAddress string
	var wwwCloser io.Closer

	var ctx context.Context
	var cancel context.CancelFunc

	BeforeEach(func() {
		// Using IPv4 addresses as IPv6 is disabled by default in Docker
		sutListener = Must(net.Listen("tcp", "127.0.0.1:0"))
		sutClient = HttpClientViaProxy(HttpAddressFromListener(sutListener))

		wwwListener := Must(net.Listen("tcp", "127.0.0.1:0"))
		wwwAddress, wwwCloser = RunServer(wwwListener)
		Expect(wwwAddress).ToNot(BeEmpty())

		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	})
	AfterEach(func() {
		cancel()
		wwwCloser.Close()
	})

	It("should pass through content from a webserver", func() {
		go proxy.RunProxy(ctx, sutListener, nil)

		resp, err := sutClient.Get(wwwAddress)
		Expect(resp, err).ToNot(BeNil())
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		Expect(resp.Header.Get("x-max-bitrate")).To(Equal("0"))
		Expect(b, err).To(HaveLen(testChunkSize))
	})

	DescribeTable("should throttle when asked", func(bps uint, consume int) {
		go proxy.RunProxy(ctx, sutListener, nil, proxy.SetMaxBitrate(bps))

		// ioshaper.Bitrate is unit tested elsewhere so acceptable for expectedDuration
		expectedDuration := ioshaper.Bitrate(bps).TransferBytes(consume)
		expectedBandwidthHeader := strconv.FormatUint(uint64(bps), 10)

		tb := gmeasure.NewExperiment(fmt.Sprintf("download rate @ %d bps", bps))
		AddReportEntry(tb.Name, tb)

		tb.Sample(
			func(idx int) {
				resp, err := sutClient.Get(wwwAddress)
				Expect(resp, err).ToNot(BeNil())
				defer resp.Body.Close()

				tb.MeasureDuration("download", func() {
					buf, err := io.ReadAll(io.LimitReader(resp.Body, int64(consume)))
					Expect(resp.Header.Get("x-max-bitrate")).To(Equal(expectedBandwidthHeader))
					Expect(buf, err).To(HaveLen(consume))
				})
			},
			gmeasure.SamplingConfig{
				N:           5,
				Duration:    10 * time.Second,
				NumParallel: 1,
			},
		)

		medianDuration := tb.GetStats("download").DurationFor(gmeasure.StatMean)
		Expect(medianDuration).To(BeNumerically("~", expectedDuration, expectedDuration/10))
	},
		Entry("1 Mbps", uint(1024*1024), 512*1024),
		Entry("3 Mbps", uint(3*1024*1024), 1*1024*1024),
		Entry("10 Mbps", uint(10*1024*1024), 3*1024*1024),
	)
})

func RunServer(listener net.Listener) (string, io.Closer) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, testChunkSize))
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	return HttpAddressFromListener(listener).String(), server
}

func HttpClientViaProxy(proxyUrl *url.URL) *http.Client {
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}
}

func HttpAddressFromListener(listener net.Listener) *url.URL {
	return &url.URL{Scheme: "http", Host: listener.Addr().String()}
}
