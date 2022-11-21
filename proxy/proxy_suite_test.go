package proxy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Proxy Suite")
}

func Must[T any](r T, err error) T {
	ExpectWithOffset(1, err).To(Succeed())
	return r
}
