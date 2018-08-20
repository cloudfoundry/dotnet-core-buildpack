package integration_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/agouti"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var packagedBuildpack cutlass.VersionedBuildpackPackage
var agoutiDriver *agouti.WebDriver

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use (builds if empty)")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "512M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	if buildpackVersion == "" {
		packagedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpack(os.Getenv("CF_STACK"), ApiHasStackAssociation())
		Expect(err).NotTo(HaveOccurred())

		buildpackFile, err := os.Open(packagedBuildpack.File) // For read access.
		Expect(err).NotTo(HaveOccurred())
		buildpackFileStat, err := buildpackFile.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(buildpackFileStat.Size() < 1024*1024*1024).To(BeTrue(), "Buildpack file size must be less than 1G")

		data, err := json.Marshal(packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	return []byte{}
}, func(data []byte) {
	// Run on all nodes
	var err error
	if len(data) > 0 {
		err = json.Unmarshal(data, &packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		buildpackVersion = packagedBuildpack.Version
	}

	bpDir, err = cutlass.FindRoot()
	Expect(err).NotTo(HaveOccurred())

	Expect(cutlass.CopyCfHome()).To(Succeed())

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter

	// agoutiDriver = agouti.PhantomJS()
	// agoutiDriver = agouti.Selenium()
	agoutiDriver = agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}))
	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
	Expect(agoutiDriver.Stop()).To(Succeed())
}, func() {
	// Run once
	cutlass.RemovePackagedBuildpack(packagedBuildpack)
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
}

func Restart(app *cutlass.App) {
	Expect(app.Restart()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func ApiGreaterThan(version string) bool {
	apiVersionString, err := cutlass.ApiVersion()
	Expect(err).To(BeNil())
	apiVersion, err := semver.Make(apiVersionString)
	Expect(err).To(BeNil())
	reqVersion, err := semver.ParseRange(">= " + version)
	Expect(err).To(BeNil())
	return reqVersion(apiVersion)
}

func ApiHasTask() bool {
	supported, err := cutlass.ApiGreaterThan("2.75.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasMultiBuildpack() bool {
	supported, err := cutlass.ApiGreaterThan("2.90.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func SkipUnlessUncached() {
	if cutlass.Cached {
		Skip("Running cached tests")
	}
}

func SkipUnlessCached() {
	if !cutlass.Cached {
		Skip("Running uncached tests")
	}
}

func SkipUnlessStack(requiredStack string) {
	currentStack := os.Getenv("CF_STACK")
	if currentStack != requiredStack {
		Skip(fmt.Sprintf("Skipping because the stack \"%s\" is not supported", currentStack))
	}
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func DefaultVersion(name string) string {
	m := &libbuildpack.Manifest{}
	err := (&libbuildpack.YAML{}).Load(filepath.Join(bpDir, "manifest.yml"), m)
	Expect(err).ToNot(HaveOccurred())
	dep, err := m.DefaultVersion(name)
	Expect(err).ToNot(HaveOccurred())
	Expect(dep.Version).ToNot(Equal(""))
	return dep.Version
}

func AssertUsesProxyDuringStagingIfPresent(fixtureName string) {
	Context("with an uncached buildpack", func() {
		BeforeEach(SkipUnlessUncached)

		It("uses a proxy during staging if present", func() {
			proxy, err := cutlass.NewProxy()
			Expect(err).To(BeNil())
			defer proxy.Close()

			bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
			cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
			err = cmd.Run()
			Expect(err).To(BeNil())
			defer os.Remove(bpFile)

			traffic, built, _, err := cutlass.InternetTraffic(
				bpDir,
				filepath.Join("fixtures", fixtureName),
				bpFile,
				[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
			)
			Expect(err).To(BeNil())
			Expect(built).To(BeTrue())

			destUrl, err := url.Parse(proxy.URL)
			Expect(err).To(BeNil())

			Expect(cutlass.UniqueDestination(
				traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
			)).To(BeNil())
		})
	})
}

func AssertNoInternetTraffic(fixtureName string) {
	It("has no traffic", func() {
		SkipUnlessCached()

		bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
		cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
		err := cmd.Run()
		Expect(err).To(BeNil())
		defer os.Remove(bpFile)

		traffic, _, _, err := cutlass.InternetTraffic(
			bpDir,
			filepath.Join("fixtures", fixtureName),
			bpFile,
			[]string{},
		)
		Expect(err).To(BeNil())
		// Expect(built).To(BeTrue())
		Expect(traffic).To(BeEmpty())
	})
}

func GetLatestPatchVersion(dep, constraint, bpDir string) string {
	manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
	Expect(err).ToNot(HaveOccurred())
	deps := manifest.AllDependencyVersions(dep)
	runtimeVersion, err := libbuildpack.FindMatchingVersion(constraint, deps)
	Expect(err).ToNot(HaveOccurred())

	return runtimeVersion
}

func ReplaceFileTemplate(bpDir, fixture, file, templateVar, replaceVal string) *cutlass.App {
	dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", fixture))
	Expect(err).ToNot(HaveOccurred())

	data, err := ioutil.ReadFile(filepath.Join(dir, file))
	Expect(err).ToNot(HaveOccurred())
	data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", templateVar)), []byte(replaceVal), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, file), data, 0644)).To(Succeed())

	return cutlass.New(dir)
}

func PrintFailureLogs(appName string) error {
	if !CurrentGinkgoTestDescription().Failed {
		return nil
	}
	command := exec.Command("cf", "logs", appName, "--recent")
	command.Stdout = GinkgoWriter
	command.Stderr = GinkgoWriter
	return command.Run()
}
