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
	oldDotnetVersion := FirstOfVersionLine("dotnet", "2.1.301")
	bratshelper.StagingWithADepThatIsNotTheLatestConstrained("dotnet", oldDotnetVersion, func(v string) *cutlass.App { return CopyBratsWithFramework(v, "2.1.x") })
	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(`dotnet\.[\d\.]+\.linux\-amd64\-[\da-f]+\.tar.xz`, CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript("dotnet", CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	compatible := func(sdkVersion, frameworkVersion string) bool {

		var sdk, framework semver.Version
		var err error
		if sdk, err = semver.Parse(sdkVersion); err != nil {
			panic(err)
		}
		if framework, err = semver.Parse(frameworkVersion); err != nil {
			panic(err)
		}

		ret := sdk.Major == framework.Major

		sdk2_1_300, _ := semver.Parse("2.1.300")
		framework2_1_0, _ := semver.Parse("2.1.0")

		if framework.GTE(framework2_1_0) {
			ret = ret && sdk.GTE(sdk2_1_300)
		}

		return ret

	}
	bratshelper.ForAllSupportedVersions2("dotnet", "dotnet-framework", compatible, "with .NET SDK version: %s and .NET Framework version: %s", CopyBratsWithFramework, func(sdkVersion, frameworkVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of .NET SDK + .NET Framework", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet " + sdkVersion))
			Expect(app.Stdout.String()).To(MatchRegexp("(Using dotnet framework installed in .*\\Q/dotnet/shared/Microsoft.NETCore.App/%s\\E|\\QInstalling dotnet-framework %s\\E)", frameworkVersion, frameworkVersion))
		})

		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
})
