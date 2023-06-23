package integration_test

import (
	"fmt"
	"github.com/cloudfoundry/switchblade"
	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"path/filepath"
	"testing"
)

func testFrameworkDependant(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("with libgdiplus", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "fde_apps", "libgdiplus"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("displays a simple text homepage", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
					ContainSubstring("Installing libgdiplus"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context("with .NET Core 6", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "fde_apps", "dotnet_6"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs successfully", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				fmt.Println(logs.String())

				Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
			})
		})
	}
}
