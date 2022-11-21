package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync/atomic"

	"github.com/cognitive-i/bandwidth-proxy/ioshaper"
)

type config struct {
	bitrate uint64
}
type configOption func(c *config)

func (c *config) setOptions(configOptions []configOption) {
	for _, option := range configOptions {
		option(c)
	}
}

func (c *config) GetMaxBitrate() uint64 {
	return atomic.LoadUint64(&c.bitrate)
}

func (c *config) SetMaxBitrate(bitrate uint64) {
	atomic.StoreUint64(&c.bitrate, bitrate)
}

func SetMaxBitrate(bps uint) configOption {
	return func(c *config) { c.SetMaxBitrate(uint64(bps)) }
}

func RunProxy(ctx context.Context, proxyListener, controlListener net.Listener, configOptions ...configOption) {
	config := &config{}
	config.setOptions(configOptions)

	if proxyListener != nil {
		proxy := newProxyServer(config)
		defer proxy.Close()

		go proxy.Serve(proxyListener)
	}

	if controlListener != nil {
		control := newControlServer(config)
		defer control.Close()

		go control.Serve(controlListener)
	}

	<-ctx.Done()
}

func newProxyServer(config *config) *http.Server {
	handler := &httputil.ReverseProxy{
		Director: func(r *http.Request) {},
		ModifyResponse: func(r *http.Response) error {
			bitrate := config.GetMaxBitrate()
			r.Header.Add("x-max-bitrate", strconv.FormatUint(bitrate, 10))

			if bitrate > 0 {
				r.Body = ioshaper.NewThrottledReader(r.Body, bitrate)
			}

			return nil
		},
	}

	return &http.Server{Handler: handler}
}
