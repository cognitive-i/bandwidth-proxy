package ioshaper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIoshaper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ioshaper Suite")
}
