package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testMemoryLimit(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			fixture string
			name    string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())

			fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		it("exports DOTNET_GCHeapHardLimit computed from MEMORY_LIMIT", func() {
			deployment, _, err := platform.Deploy.Execute(name, fixture)
			Expect(err).NotTo(HaveOccurred())

			// 75% (the default) of the switchblade docker harness's fixed MEMORY_LIMIT=1024m
			// (1073741824 bytes) = 805306368 bytes = 0x30000000.
			Eventually(deployment).Should(Serve(Equal("0x30000000")).WithEndpoint("/env/dotnet-gc-heap-hard-limit"))
		})

		it("respects a caller-provided DOTNET_GCHeapHardLimitPercent", func() {
			deployment, _, err := platform.Deploy.
				WithEnv(map[string]string{"DOTNET_GCHeapHardLimitPercent": "50"}).
				Execute(name, fixture)
			Expect(err).NotTo(HaveOccurred())

			Eventually(deployment).Should(Serve(Equal("0x20000000")).WithEndpoint("/env/dotnet-gc-heap-hard-limit"))
		})
	}
}
