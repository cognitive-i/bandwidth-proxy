package proxy

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync/atomic"

	"github.com/cognitive-i/bandwidth-proxy/ioshaper"
)

type config struct {
	bitrate int64
}
type configOption func(c *config) error

func (c *config) setOptions(configOptions []configOption) error {
	for _, option := range configOptions {
		if err := option(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) GetMaxBitrate() int64 {
	return atomic.LoadInt64(&c.bitrate)
}

func (c *config) SetMaxBitrate(bitrate int64) error {
	if bitrate < 0 {
		return errors.New("bitrate < 0")
	}

	atomic.StoreInt64(&c.bitrate, bitrate)
	return nil
}

func SetMaxBitrate(bps int) configOption {
	return func(c *config) error { return c.SetMaxBitrate(int64(bps)) }
}

func RunProxy(ctx context.Context, proxyListener, controlListener net.Listener, configOptions ...configOption) {
	config := &config{}
	if err := config.setOptions(configOptions); err != nil {
		log.Fatal("configuration error ", err)
	}

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
			r.Header.Add("x-max-bitrate", strconv.FormatInt(bitrate, 10))

			if bitrate > 0 {
				r.Body = ioshaper.NewThrottledReader(r.Body, bitrate)
			}

			return nil
		},
	}

	return &http.Server{Handler: handler}
}
