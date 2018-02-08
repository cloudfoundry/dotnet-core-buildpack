package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("override yml", func() {
	var app *cutlass.App
	var buildpackName string
	AfterEach(func() {
		if buildpackName != "" {
			cutlass.DeleteBuildpack(buildpackName)
		}
		app = DestroyApp(app)
	})

	BeforeEach(func() {
		if !ApiHasMultiBuildpack() {
			Skip("Multi buildpack support is required")
		}

		buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
		Expect(cutlass.CreateOrUpdateBuildpack(buildpackName, filepath.Join(bpDir, "fixtures", "overrideyml_bp"))).To(Succeed())

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "console_app"))
		app.Buildpacks = []string{buildpackName + "_buildpack", "dotnet-core_buildpack"}
	})

	It("Forces dotnet from override buildpack", func() {
		Expect(app.Push()).ToNot(Succeed())
		Eventually(app.Stdout.String).Should(ContainSubstring("-----> OverrideYML Buildpack"))
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())

		Eventually(app.Stdout.String).Should(ContainSubstring("-----> Installing dotnet"))
		Eventually(app.Stdout.String).Should(MatchRegexp("Copy .*/dotnet.tgz"))
		Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc"))
	})
})
