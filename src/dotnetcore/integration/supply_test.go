package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testSupply(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("with csproj file", func() {
			context("the app is pushed", func() {
				it("finds the supplied dependency in the runtime container", func() {
					deployment, logs, err := platform.Deploy.
						WithBuildpacks("go_buildpack", "dotnet_core_buildpack").
						Execute(name, filepath.Join(fixtures, "supply", "dotnet_app"))
					Expect(err).To(Succeed())

					Expect(logs).To(ContainSubstring("Installing go"), logs.String())
					Eventually(deployment).Should(Serve(MatchRegexp("go: go version go\\d+.\\d+.\\d+ linux/amd64")))
				})
			})
		})

		context("with no csproj file", func() {
			context("the app is pushed once", func() {
				it("finds the supplied dependency in the runtime container", func() {
					deployment, logs, err := platform.Deploy.
						WithBuildpacks("dotnet_core_buildpack", "staticfile_buildpack").
						Execute(name, filepath.Join(fixtures, "supply", "staticfile_app"))
					Expect(err).To(Succeed())

					Expect(logs).To(ContainSubstring("Supplying Dotnet Core"))
					Eventually(deployment).Should(Serve(ContainSubstring("This is an example app for Cloud Foundry that is only static HTML/JS/CSS assets.")))
				})
			})
		})
	}
}
