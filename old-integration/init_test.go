package integration_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/switchblade"
	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpack struct {
		Version string
		Path    string
	}

	Cached       bool
	Serial       bool
	FixturesPath string
	GitHubToken  string
	Platform     string
	Stack        string
}

func init() {
	flag.BoolVar(&settings.Cached, "cached", false, "run cached buildpack tests")
	flag.BoolVar(&settings.Serial, "serial", false, "run serial buildpack tests")
	flag.StringVar(&settings.Platform, "platform", "cf", `switchblade platform to test against ("cf" or "docker")`)
	flag.StringVar(&settings.GitHubToken, "github-token", "", "use the token to make GitHub API requests")
	flag.StringVar(&settings.Stack, "stack", "cflinuxfs4", "stack to use as default when pusing apps")
}

func TestIntegration(t *testing.T) {
	var Expect = NewWithT(t).Expect

	format.MaxLength = 0
	SetDefaultEventuallyTimeout(10 * time.Second)

	root, err := filepath.Abs("./../../..")
	Expect(err).NotTo(HaveOccurred())

	fixtures := filepath.Join(root, "fixtures")

	platform, err := switchblade.NewPlatform(settings.Platform, settings.GitHubToken, settings.Stack)
	Expect(err).NotTo(HaveOccurred())

	// goBuildpackFile, err := downloadBuildpack("go")
	// Expect(err).NotTo(HaveOccurred())

	err = platform.Initialize(
		switchblade.Buildpack{
			Name: "dotnet_core_buildpack",
			URI:  os.Getenv("BUILDPACK_FILE"),
		},
		// switchblade.Buildpack{
		// 	Name: "override_buildpack",
		// 	URI:  filepath.Join(fixtures, "util", "override_buildpack"),
		// },
		// // Go buildpack is needed for the proxy, multibuildpack, and the dynatrace apps
		// switchblade.Buildpack{
		// 	Name: "go_buildpack",
		// 	URI:  goBuildpackFile,
		// },
	)
	Expect(err).NotTo(HaveOccurred())

	// dynatraceName, err := switchblade.RandomName()
	// Expect(err).NotTo(HaveOccurred())

	// dynatraceDeploymentProcess := platform.Deploy.WithBuildpacks("go_buildpack")

	// dynatraceDeployment, _, err := dynatraceDeploymentProcess.
	// 	Execute(dynatraceName, filepath.Join(fixtures, "util", "dynatrace"))
	// Expect(err).NotTo(HaveOccurred())

	suite := spec.New("integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault(platform, fixtures))
	// suite("Dynatrace", testDynatrace(platform, fixtures, dynatraceDeployment.InternalURL))
	// suite("Errors", testErrors(platform, fixtures))
	// suite("Override", testOverride(platform, fixtures))

	// if !settings.Cached {
	// 	suite("Cache", testCache(platform, fixtures))
	// }

	suite.Run(t)

	// Expect(platform.Delete.Execute(dynatraceName)).To(Succeed())
	Expect(os.Remove(os.Getenv("BUILDPACK_FILE"))).To(Succeed())
	// Expect(os.Remove(goBuildpackFile)).To(Succeed())
	Expect(platform.Deinitialize()).To(Succeed())
}

// func downloadBuildpack(name string) (string, error) {
// 	uri := fmt.Sprintf("https://github.com/cloudfoundry/%s-buildpack/archive/master.zip", name)

// 	file, err := os.CreateTemp("", fmt.Sprintf("%s-buildpack-*.zip", name))
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()

// 	resp, err := http.Get(uri)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	_, err = io.Copy(file, resp.Body)
// 	return file.Name(), err
// }

// import (
// 	"bytes"
// 	"flag"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"testing"
// 	"time"

// 	"github.com/cloudfoundry/libbuildpack"
// 	"github.com/cloudfoundry/libbuildpack/cutlass"
// 	"github.com/onsi/gomega/format"
// 	"github.com/sclevine/spec"
// 	"github.com/sclevine/spec/report"

// 	. "github.com/onsi/gomega"
// )

// var settings struct {
// 	Buildpack struct {
// 		Version string
// 		Path    string
// 	}
// 	Dynatrace struct {
// 		App *cutlass.App
// 		URI string
// 	}
// 	FixturesPath string
// 	GitHubToken  string
// 	Platform     string
// 	Stack        string
// }

// func init() {
// 	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
// 	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
// 	flag.StringVar(&cutlass.DefaultDisk, "disk", "512M", "default disk for pushed apps")
// 	flag.StringVar(&settings.Buildpack.Version, "version", "", "version to use (builds if empty)")
// 	flag.StringVar(&settings.GitHubToken, "github-token", "", "use the token to make GitHub API requests")
// 	flag.StringVar(&settings.Platform, "platform", "cf", "platform to run against")
// 	flag.StringVar(&settings.Stack, "stack", "cflinuxfs3", "stack to use when pushing apps")
// }

// func TestIntegration(t *testing.T) {
// 	format.MaxLength = 0

// 	var (
// 		Expect     = NewWithT(t).Expect
// 		Eventually = NewWithT(t).Eventually

// 		packagedBuildpack cutlass.VersionedBuildpackPackage
// 	)

// 	root, err := cutlass.FindRoot()
// 	Expect(err).NotTo(HaveOccurred())

// 	settings.FixturesPath = filepath.Join(root, "fixtures")

// 	if settings.Buildpack.Version == "" {
// 		packagedBuildpack, err = cutlass.PackageUniquelyVersionedBuildpack(os.Getenv("CF_STACK"), true)
// 		Expect(err).NotTo(HaveOccurred())

// 		settings.Buildpack.Path = packagedBuildpack.File

// 		info, err := os.Stat(settings.Buildpack.Path)
// 		Expect(err).NotTo(HaveOccurred())
// 		Expect(info.Size() < 1024*1024*1024).To(BeTrue(), "Buildpack file size must be less than 1G")

// 		settings.Buildpack.Version = packagedBuildpack.Version
// 	}

// 	err = cutlass.CreateOrUpdateBuildpack("override", filepath.Join(settings.FixturesPath, "util", "override_buildpack"), "")
// 	Expect(err).NotTo(HaveOccurred())

// 	Expect(cutlass.CopyCfHome()).To(Succeed())
// 	cutlass.SeedRandom()

// 	settings.Dynatrace.App = cutlass.New(filepath.Join(settings.FixturesPath, "util", "dynatrace"))

// 	// This is done to have the dynatrace broker app running on default
// 	// cf-deployment envs. They do not come with cflinuxfs3 buildpacks.
// 	if os.Getenv("CF_STACK") == "cflinuxfs3" {
// 		settings.Dynatrace.App.Buildpacks = []string{"https://github.com/cloudfoundry/go-buildpack"}
// 	}
// 	settings.Dynatrace.App.SetEnv("BP_DEBUG", "true")

// 	Expect(settings.Dynatrace.App.Push()).To(Succeed())
// 	Eventually(func() ([]string, error) {
// 		return settings.Dynatrace.App.InstanceStates()
// 	}, 60*time.Second).Should(Equal([]string{"RUNNING"}))

// 	settings.Dynatrace.URI, err = settings.Dynatrace.App.GetUrl("")
// 	Expect(err).NotTo(HaveOccurred())

// 	suite := spec.New("integration", spec.Report(report.Terminal{}), spec.Parallel())
// 	suite("Default", testDefault)
// 	suite("Dynatrace", testDynatrace)
// 	suite("Fsharp", testFsharp)
// 	suite("MultipleProjects", testMultipleProjects)
// 	suite("Node", testNode)
// 	suite("Override", testOverride)
// 	suite("Supply", testSupply)
// 	suite("Sealights", testSealights)

// 	if cutlass.Cached {
// 		suite("Offline", testOffline)
// 	} else {
// 		suite("Cache", testCache)
// 	}

// 	suite.Run(t)

// 	DestroyApp(t, settings.Dynatrace.App)
// 	Expect(cutlass.RemovePackagedBuildpack(packagedBuildpack)).To(Succeed())
// 	Expect(cutlass.DeleteBuildpack("override")).To(Succeed())
// 	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
// }

// func PushAppAndConfirm(t *testing.T, app *cutlass.App) {
// 	t.Helper()

// 	var (
// 		Expect     = NewWithT(t).Expect
// 		Eventually = NewWithT(t).Eventually
// 	)

// 	Expect(app.Push()).To(Succeed())
// 	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
// 	Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
// }

// func DestroyApp(t *testing.T, app *cutlass.App) *cutlass.App {
// 	t.Helper()

// 	var Expect = NewWithT(t).Expect
// 	Expect(app.Destroy()).To(Succeed())
// 	return nil
// }

func GetLatestDepVersion(t *testing.T, dep, constraint string) string {
	t.Helper()

	var Expect = NewWithT(t).Expect

	root, err := filepath.Abs("./../../..")
	Expect(err).NotTo(HaveOccurred())

	manifest, err := libbuildpack.NewManifest(root, nil, time.Now())
	Expect(err).ToNot(HaveOccurred())
	deps := manifest.AllDependencyVersions(dep)
	runtimeVersion, err := libbuildpack.FindMatchingVersion(constraint, deps)
	Expect(err).ToNot(HaveOccurred())

	return runtimeVersion
}

// func ReplaceFileTemplate(t *testing.T, pathToFixture, file, templateVar, replaceVal string) *cutlass.App {
// 	t.Helper()

// 	var Expect = NewWithT(t).Expect

// 	dir, err := cutlass.CopyFixture(pathToFixture)
// 	Expect(err).ToNot(HaveOccurred())

// 	data, err := os.ReadFile(filepath.Join(dir, file))
// 	Expect(err).ToNot(HaveOccurred())
// 	data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", templateVar)), []byte(replaceVal), -1)
// 	Expect(os.WriteFile(filepath.Join(dir, file), data, 0644)).To(Succeed())

// 	return cutlass.New(dir)
// }

// func SkipOnCflinuxfs4(t *testing.T) {
// 	if os.Getenv("CF_STACK") == "cflinuxfs4" {
// 		t.Skip("Skipping test not relevant for stack cflinuxfs4")
// 	}
// }

// func SkipOnCflinuxfs3(t *testing.T) {
// 	if os.Getenv("CF_STACK") == "cflinuxfs3" {
// 		t.Skip("Skipping test not relevant for stack cflinuxfs3")
// 	}
// }
