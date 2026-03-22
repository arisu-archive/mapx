package mapx_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMapx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mapx Suite")
}
