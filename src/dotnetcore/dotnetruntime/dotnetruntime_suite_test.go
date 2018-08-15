package dotnetruntime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDotnetruntime(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dotnetruntime Suite")
}
