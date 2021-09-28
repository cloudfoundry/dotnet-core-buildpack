package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOverride(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		app        *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "console_app"))
		app.Buildpacks = []string{"override_buildpack", "dotnet_core_buildpack"}
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("forces dotnet-sdk from override buildpack", func() {
		Expect(app.V3Push()).ToNot(Succeed())
		Expect(app.Stdout.String()).To(ContainSubstring("-----> OverrideYML Buildpack"))
		Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())

		Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing dotnet-sdk"))
		Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/dotnet-sdk.tgz"))
		Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc"))
	})
}
