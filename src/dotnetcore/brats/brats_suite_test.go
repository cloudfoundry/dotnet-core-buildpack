package brats_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "512M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return bratshelper.InitBpData(os.Getenv("CF_STACK"), ApiHasStackAssociation()).Marshal()
}, func(data []byte) {
	// Run on all nodes
	bratshelper.Data.Unmarshal(data)
	Expect(cutlass.CopyCfHome()).To(Succeed())
	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Uncached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(bratshelper.Data.CachedFile)).To(Succeed())
	Expect(os.Remove(bratshelper.Data.UncachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func FirstOfVersionLine(dependency, line string) string {
	bpDir, err := cutlass.FindRoot()
	if err != nil {
		panic(err)
	}
	manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
	if err != nil {
		panic(err)
	}
	deps := manifest.AllDependencyVersions(dependency)
	versions, err := libbuildpack.FindMatchingVersions(line, deps)
	if err != nil {
		panic(err)
	}
	return versions[0]
}

func copyBratsWithRuntime(sdkVersion, runtimeVersion, fixture string) *cutlass.App {
	manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
	Expect(err).ToNot(HaveOccurred())

	if sdkVersion == "" {
		sdkVersion = "x"
	}

	if strings.Contains(sdkVersion, "x") {
		deps := manifest.AllDependencyVersions("dotnet-sdk")
		sdkVersion, err = libbuildpack.FindMatchingVersion(sdkVersion, deps)
		Expect(err).ToNot(HaveOccurred())
	}

	if runtimeVersion == "" {
		majorVersion := strings.Split(sdkVersion, ".")[0]
		runtimeVersion = majorVersion + ".x"
	}

	if strings.Contains(runtimeVersion, "x") {
		deps := manifest.AllDependencyVersions("dotnet-runtime")
		runtimeVersion, err = libbuildpack.FindMatchingVersion(runtimeVersion, deps)
		Expect(err).ToNot(HaveOccurred())
	}

	versionParts := strings.Split(runtimeVersion, ".")
	netCoreApp := fmt.Sprintf("netcoreapp%s.%s", versionParts[0], versionParts[1])

	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", fixture))
	Expect(err).ToNot(HaveOccurred())

	projectFile, err := filepath.Glob(filepath.Join(dir, fixture+".*sproj"))
	Expect(err).NotTo(HaveOccurred())
	Expect(len(projectFile)).To(Equal(1))

	for _, file := range []string{filepath.Base(projectFile[0]), "global.json"} {
		data, err := ioutil.ReadFile(filepath.Join(dir, file))
		Expect(err).ToNot(HaveOccurred())

		data = bytes.Replace(data, []byte("<%= net_core_app %>"), []byte(netCoreApp), -1)
		data = bytes.Replace(data, []byte("<%= runtime_version %>"), []byte(runtimeVersion), -1)
		data = bytes.Replace(data, []byte("<%= sdk_version %>"), []byte(sdkVersion), -1)
		Expect(ioutil.WriteFile(filepath.Join(dir, file), data, 0644)).To(Succeed())
	}

	return cutlass.New(dir)
}

func CopyCSharpBratsWithRuntime(sdkVersion, runtimeVersion string) *cutlass.App {
	return copyBratsWithRuntime(sdkVersion, runtimeVersion, "simple_brats")
}

func CopyFSharpBratsWithRuntime(sdkVersion, runtimeVersion string) *cutlass.App {
	return copyBratsWithRuntime(sdkVersion, runtimeVersion, "simple_fsharp_brats")
}

func CopyBrats(sdkVersion string) *cutlass.App {
	return CopyCSharpBratsWithRuntime(sdkVersion, "")
}

func PushApp(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}
