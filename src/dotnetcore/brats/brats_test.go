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

	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)

	bratshelper.StagingWithADepThatIsNotTheLatestConstrained(
		"dotnet-sdk",
		FirstOfVersionLine("dotnet-sdk", "2.1.400"),
		func(v string) *cutlass.App { return CopyCSharpBratsWithRuntime(v, "2.1.x") },
	)

	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(
		`dotnet-sdk\.[\d\.]+\.linux\-amd64\-.*\-[\da-f]+\.tar.xz`,
		CopyBrats,
	)

	bratshelper.DeployAppWithExecutableProfileScript("dotnet-sdk", CopyBrats)

	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	compatible := func(sdkVersion, runtimeVersion string) bool {
		sdk := semver.MustParse(sdkVersion)

		runtime := semver.MustParse(runtimeVersion)

		isCompatible := sdk.Major == runtime.Major

		runtime210 := semver.MustParse("2.1.0")
		if runtime.GTE(runtime210) {
			sdk21300 := semver.MustParse("2.1.300")
			isCompatible = isCompatible && sdk.GTE(sdk21300)
		}

		return isCompatible
	}

	// Skip 1.0.X versions of the SDK when testing F# apps
	compatibleWithFSharp := func(sdkVersion, runtimeVersion string) bool {
		sdk := semver.MustParse(sdkVersion)
		if sdk.Major <= 1 && sdk.Minor < 1 {
			return false
		}
		return compatible(sdkVersion, runtimeVersion)
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
			compatibleWithFSharp,
			"with .NET SDK version: %s and .NET Runtime version: %s",
			CopyFSharpBratsWithRuntime,
			ensureAppWorks,
		)
	})
})
