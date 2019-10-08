package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("running supply buildpacks before the dotnet-core buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("the app is pushed", func() {
		BeforeEach(func() {
			if version, err := cutlass.ApiVersion(); err != nil || version == "2.65.0" {
				Skip("API version does not have multi-buildpack support")
			}

			app = cutlass.New(Fixtures("fake_supply_dotnet_app"))
			app.Buildpacks = []string{
				"https://buildpacks.cloudfoundry.org/fixtures/new_supply_bosh2.zip",
				"dotnet_core_buildpack",
			}
		})

		It("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring("SUPPLYING BOSH2"))

			Expect(app.GetBody("/")).To(MatchRegexp("bosh2: version 2.0.1-74fad57"))
		})
	})
})
