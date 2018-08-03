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
		"dotnet",
		FirstOfVersionLine("dotnet", "2.1.301"),
		func(v string) *cutlass.App { return CopyCSharpBratsWithFramework(v, "2.1.x") },
	)

	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(
		`dotnet\.[\d\.]+\.linux\-amd64\-.*\-[\da-f]+\.tar.xz`,
		CopyBrats,
	)

	bratshelper.DeployAppWithExecutableProfileScript("dotnet", CopyBrats)

	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	compatible := func(sdkVersion, frameworkVersion string) bool {
		sdk := semver.MustParse(sdkVersion)

		framework := semver.MustParse(frameworkVersion)

		isCompatible := sdk.Major == framework.Major

		framework210 := semver.MustParse("2.1.0")
		if framework.GTE(framework210) {
			sdk21300 := semver.MustParse("2.1.300")
			isCompatible = isCompatible && sdk.GTE(sdk21300)
		}

		return isCompatible
	}

	// Skip 1.0.X versions of the SDK when testing F# apps
	compatibleWithFSharp := func(sdkVersion, frameworkVersion string) bool {
		sdk := semver.MustParse(sdkVersion)
		if sdk.Major <= 1 && sdk.Minor < 1 {
			return false
		}
		return compatible(sdkVersion, frameworkVersion)
	}

	ensureAppWorks := func(sdkVersion, frameworkVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of .NET SDK + .NET Framework", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet " + sdkVersion))
			Expect(app.Stdout.String()).To(MatchRegexp(
				"(Using dotnet framework installed in .*\\Q/dotnet/shared/Microsoft.NETCore.App/%s\\E|\\QInstalling dotnet-framework %s\\E)",
				frameworkVersion,
				frameworkVersion,
			))
		})

		By("runs a simple web server", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	}

	Context("for C# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet",
			"dotnet-framework",
			compatible,
			"with .NET SDK version: %s and .NET Framework version: %s",
			CopyCSharpBratsWithFramework,
			ensureAppWorks,
		)
	})

	Context("for F# apps", func() {
		bratshelper.ForAllSupportedVersions2(
			"dotnet",
			"dotnet-framework",
			compatibleWithFSharp,
			"with .NET SDK version: %s and .NET Framework version: %s",
			CopyFSharpBratsWithFramework,
			ensureAppWorks,
		)
	})
})
