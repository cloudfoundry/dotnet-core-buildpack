package dotnetframework_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDotnetframework(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dotnetframework Suite")
}
