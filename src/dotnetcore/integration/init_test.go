package integration_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
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

	goBuildpackFile, err := downloadBuildpack("go")
	Expect(err).NotTo(HaveOccurred())

	staticfileBuildpackFile, err := downloadBuildpack("staticfile")
	Expect(err).NotTo(HaveOccurred())

	err = platform.Initialize(
		switchblade.Buildpack{
			Name: "dotnet_core_buildpack",
			URI:  os.Getenv("BUILDPACK_FILE"),
		},
		switchblade.Buildpack{
			Name: "override_buildpack",
			URI:  filepath.Join(fixtures, "util", "override_buildpack"),
		},
		// Go buildpack is needed for the supply and the dynatrace apps
		switchblade.Buildpack{
			Name: "go_buildpack",
			URI:  goBuildpackFile,
		},
		// Staticfile buildpack is needed for the supply apps
		switchblade.Buildpack{
			Name: "staticfile_buildpack",
			URI:  staticfileBuildpackFile,
		},
	)
	Expect(err).NotTo(HaveOccurred())

	dynatraceName, err := switchblade.RandomName()
	Expect(err).NotTo(HaveOccurred())

	dynatraceDeploymentProcess := platform.Deploy.WithBuildpacks("go_buildpack")

	dynatraceDeployment, _, err := dynatraceDeploymentProcess.
		Execute(dynatraceName, filepath.Join(fixtures, "util", "dynatrace"))
	Expect(err).NotTo(HaveOccurred())

	suite := spec.New("integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault(platform, fixtures))
	suite("Dynatrace", testDynatrace(platform, fixtures, dynatraceDeployment.InternalURL))
	suite("Fsharp", testFsharp(platform, fixtures))
	suite("MultipleProjects", testMultipleProjects(platform, fixtures))
	suite("Node", testNode(platform, fixtures))
	suite("Override", testOverride(platform, fixtures))
	suite("Supply", testSupply(platform, fixtures))
	suite("Sealights", testSealights(platform, fixtures))

	if settings.Cached {
		suite("Offline", testOffline(platform, fixtures))
	} else {
		suite("Cache", testCache(platform, fixtures))
	}

	suite.Run(t)

	Expect(platform.Delete.Execute(dynatraceName)).To(Succeed())
	Expect(os.Remove(os.Getenv("BUILDPACK_FILE"))).To(Succeed())
	Expect(os.Remove(goBuildpackFile)).To(Succeed())
	Expect(os.Remove(staticfileBuildpackFile)).To(Succeed())
	Expect(platform.Deinitialize()).To(Succeed())
}

func downloadBuildpack(name string) (string, error) {
	uri := fmt.Sprintf("https://github.com/cloudfoundry/%s-buildpack/archive/master.zip", name)

	file, err := os.CreateTemp("", fmt.Sprintf("%s-buildpack-*.zip", name))
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return file.Name(), err
}

func GetLatestDepVersion(t *testing.T, dep, constraint string) (string, error) {
	t.Helper()

	root, err := filepath.Abs("./../../..")
	if err != nil {
		return "", err
	}

	manifest, err := libbuildpack.NewManifest(root, nil, time.Now())
	if err != nil {
		return "", err
	}
	deps := manifest.AllDependencyVersions(dep)
	runtimeVersion, err := libbuildpack.FindMatchingVersion(constraint, deps)
	if err != nil {
		return "", err
	}

	return runtimeVersion, nil
}

func ReplaceFileTemplate(t *testing.T, pathToFixture, file, templateVar, replaceVal string) error {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(pathToFixture, file))
	if err != nil {
		return err
	}
	data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", templateVar)), []byte(replaceVal), -1)
	return os.WriteFile(filepath.Join(pathToFixture, file), data, 0644)
}

func SkipOnCflinuxfs3(t *testing.T) {
	if settings.Stack == "cflinuxfs3" {
		t.Skip("Skipping test not relevant for stack cflinuxfs3")
	}
}
