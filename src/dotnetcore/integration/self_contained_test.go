package integration_test

import (
	"github.com/cloudfoundry/switchblade"
	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"path/filepath"
	"testing"
)

func testSelfContained(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name   string
			source string
			err    error
		)

		it.Before(func() {
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("deploying a self contained msbuild app with RuntimeIdentfier", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "self_contained_apps", "msbuild"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("displays a simple text homepage", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context("deploying .NET Core 6 self-contained app", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "self_contained_apps", "executable_dotnet_6"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs successfully", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Hello, world!")))
			})
		})
	}
}
