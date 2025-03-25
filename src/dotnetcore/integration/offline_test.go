package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("when deploying a fdd without internet", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "fdd_apps", "fdd_8.0"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Welcome")))
			})
		})

		context("when deploying a self contained app without internet", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "self_contained_apps", "msbuild"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})
	}
}
