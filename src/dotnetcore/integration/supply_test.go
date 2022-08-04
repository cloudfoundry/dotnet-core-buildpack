package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSupply(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		app    *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "console_app"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("with csproj file", func() {
		context("the app is pushed", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "supply", "dotnet_app"))
				app.Buildpacks = []string{
					"https://buildpacks.cloudfoundry.org/fixtures/new_supply_bosh2.zip",
					"dotnet_core_buildpack",
				}
			})

			it("finds the supplied dependency in the runtime container", func() {
				PushAppAndConfirm(t, app)
				Expect(app.Stdout.String()).To(ContainSubstring("SUPPLYING BOSH2"))
				Expect(app.GetBody("/")).To(MatchRegexp("bosh2: version 2.0.1-74fad57"))
			})
		})
	})

	context("with no csproj file", func() {
		context("the app is pushed once", func() {
			it.Before(func() {
				// Staticfile does not support cflinuxfs4 yet
				SkipOnCflinuxfs4(t)
				app = cutlass.New(filepath.Join(settings.FixturesPath, "supply", "staticfile_app"))
				app.Buildpacks = []string{
					"dotnet_core_buildpack",
					"https://github.com/cloudfoundry/staticfile-buildpack/#master",
				}
				app.Disk = "1G"
			})

			it("finds the supplied dependency in the runtime container", func() {
				PushAppAndConfirm(t, app)
				Expect(app.Stdout.String()).To(ContainSubstring("Supplying Dotnet Core"))
				Expect(app.GetBody("/")).To(ContainSubstring("This is an example app for Cloud Foundry that is only static HTML/JS/CSS assets."))
			})
		})
	})
}
