package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOverride(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

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

		it("forces dotnet-sdk from override buildpack", func() {
			_, logs, err := platform.Deploy.
				WithBuildpacks("override_buildpack", "dotnet_core_buildpack").
				Execute(name, filepath.Join(fixtures, "console_app"))
			Expect(err).NotTo(Succeed())

			Expect(logs).To(ContainLines(ContainSubstring("-----> OverrideYML Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("-----> Installing dotnet-sdk")))
			Expect(logs).To(ContainLines(MatchRegexp("Copy .*/dotnet-sdk.tgz")))
			Expect(logs).To(ContainLines(ContainSubstring("Unable to install Dotnet SDK: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc")))
		})
	}
}
