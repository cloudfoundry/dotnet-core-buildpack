package brats_test

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet buildpack", func() {
	bratshelper.UnbuiltBuildpack("dotnet", CopyBrats)

	bratshelper.DeployAppWithExecutableProfileScript("dotnet-sdk", CopyBrats)

	compatible := func(sdkVersion, runtimeVersion string) bool {
		sdkSemVer := semver.MustParse(sdkVersion)
		runtimeSemVer := semver.MustParse(runtimeVersion)
		return sdkSemVer.Major == runtimeSemVer.Major && sdkSemVer.Minor == runtimeSemVer.Minor
	}

	ensureAppWorks := func(sdkVersion, runtimeVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of .NET SDK + .NET Runtime", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet-sdk " + sdkVersion))
			Expect(app.Stdout.String()).To(MatchRegexp(
				"(Using dotnet runtime installed in .*\\Q/dotnet-sdk/shared/Microsoft.NETCore.App/%s\\E|\\QInstalling dotnet-runtime %s\\E)",
				runtimeVersion,
				runtimeVersion,
			))
		})

		By("runs a simple web server", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	}

	Context("for C# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet-sdk",
			"dotnet-runtime",
			compatible,
			"with .NET SDK version: %s and .NET Runtime version: %s",
			CopyCSharpBratsWithRuntime,
			ensureAppWorks,
		)
	})

	Context("for F# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet-sdk",
			"dotnet-runtime",
			compatible,
			"with .NET SDK version: %s and .NET Runtime version: %s",
			CopyFSharpBratsWithRuntime,
			ensureAppWorks,
		)
	})
})
