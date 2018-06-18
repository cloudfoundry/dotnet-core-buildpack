package integration_test

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("deploying simple web app with dotnet 2.0", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "dotnet2"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})

		AssertUsesProxyDuringStagingIfPresent("dotnet2")
	})

	Context("deploying simple web app with dotnet 2.0 using dotnet 2.0 sdk", func() {
		var sdkVersion string
		BeforeEach(func() {
			sdkVersion = GetLatestPatchVersion("dotnet", "2.0.x", bpDir)
			app = ReplaceFileTemplate(bpDir, "dotnet2_with_global_json", "global.json", "sdk_version", sdkVersion)
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet %s", sdkVersion)))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})
	})

	Context("deploying with a buildpack.yml and global.json files", func() {
		Context("when SDK versions match/overlap", func() {
			var sdkVersion string
			BeforeEach(func() {
				sdkVersion = GetLatestPatchVersion("dotnet", "2.0.x", bpDir)
				app = ReplaceFileTemplate(bpDir, "with_buildpack_yml", "buildpack.yml", "sdk_version", "2.0.x")
			})

			It("buildpacks.yml sdk version overrides global.json and floats on patch", func() {
				PushAppAndConfirm(app)

				Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet %s", sdkVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
			})
		})

		Context("when SDK versions don't match", func() {
			var sdkVersion string
			BeforeEach(func() {
				sdkVersion = GetLatestPatchVersion("dotnet", "1.0.x", bpDir)
				app = ReplaceFileTemplate(bpDir, "with_buildpack_yml", "buildpack.yml", "sdk_version", "1.0.x")
			})

			It("the buildpack installs the version from buildpack.yml and dotnet complains", func() {
				Expect(app.Push()).ToNot(Succeed())

				Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet %s", sdkVersion)))
				Eventually(app.Stdout.String).Should(ContainSubstring("The specified SDK version [2.0.1] from global.json [/tmp/app/global.json] not found"))
			})
		})

		Context("when SDK version from buildpack.yml is not available", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(bpDir, "with_buildpack_yml", "buildpack.yml", "sdk_version", "2.0.0-preview7")
			})

			It("fails due to missing SDK", func() {
				Expect(app.Push()).ToNot(Succeed())

				Eventually(app.Stdout.String).Should(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
				Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet: no match found for 2.0.0-preview7"))
			})
		})
	})

	Context("deploying an mvc app with node prerendering", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "asp_prerender_node"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("1 + 2 = 3"))
		})
	})
	Context("deploying simple web app with missing sdk", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "missing_sdk"))
		})

		It("Logs a warning about using default SDK", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("SDK 2.0.0-preview-007 in global.json is not available"))
			Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})
	})
	Context("deploying an msbuild app with RuntimeIdentfier", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained_msbuild"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(MatchRegexp("Removing dotnet"))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
	Context("simple netcoreapp2 (dotnet new mvc --framework netcoreapp2.0)", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "netcoreapp2"))
		})

		It("publishes and runs", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
		})
	})
	Context("with runtimeconfig.json", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "float_runtimeconfig"))
		})

		It("installs the latest patch of dotnet runtime framework from the runtimeconfig.json", func() {
			latestPatch := GetLatestPatchVersion("dotnet-framework", "2.0.x", bpDir)
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Required dotnetframework versions: [%s]", latestPatch)))
		})
	})
	Context("with runtimeconfig.json and applyPatches false", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "apply_patches_false"))
		})

		It("installs the exact version of dotnet runtime framework from the runtimeconfig.json", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet-framework 2.0.0"))
		})
	})
})
