package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testNode(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("deploying an angular app", func() {
			it("displays a simple text homepage", func() {
				prevTimeout := os.Getenv("CF_STAGING_TIMEOUT")
				os.Setenv("CF_STAGING_TIMEOUT", "30")
				defer os.Setenv("CF_STAGING_TIMEOUT", prevTimeout)

				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "node_apps", "angular_dotnet"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("<title>source_app</title>")))
			})
		})
	}
}
