package integration_test

import (
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
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "dotnet2_with_global_json"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet 2.0.3"))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
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
			Expect(app.Stdout.String()).To(ContainSubstring("SDK 2.0.0-preview-007 not available"))
			// Expect(app.Stdout.String()).To(ContainSubstring("using the default SDK"))
			Expect(app.Stdout.String()).To(ContainSubstring("using latest version in version line"))
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
})
