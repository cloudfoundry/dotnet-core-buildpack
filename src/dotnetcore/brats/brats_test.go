package brats_test

import (
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet buildpack", func() {
	bratshelper.UnbuiltBuildpack("dotnet", CopyBrats)
	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)
	// bratshelper.StagingWithBuildpackThatSetsEOL("dotnet", func(_ string) *cutlass.App {
	// 	Skip("No EOL dates in dotnet manifest")
	// 	return nil
	// })
	oldVersion := FirstOfVersionLine("1.1.x")
	bratshelper.StagingWithADepThatIsNotTheLatestConstrained("dotnet", oldVersion, func(v string) *cutlass.App { return CopyBratsWithFramework(v, v) })
	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(`dotnet\.[\d\.]+\.linux\-amd64\-[\da-f]+\.tar.xz`, CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript("dotnet", CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	compatible := func(sdkVersion, frameworkVersion string) bool {
		return sdkVersion[0] == frameworkVersion[0]
	}
	bratshelper.ForAllSupportedVersions2("dotnet", "dotnet-framework", compatible, "with .NET SDK version: %s and .NET Framework version: %s", CopyBratsWithFramework, func(sdkVersion, frameworkVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct version of .NET SDK + .NET Framework", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet " + sdkVersion))
			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet-framework " + frameworkVersion))
		})
		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
})
