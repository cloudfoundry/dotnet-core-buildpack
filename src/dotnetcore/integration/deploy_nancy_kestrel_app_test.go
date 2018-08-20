package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"fmt"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	BeforeEach(func() {
		stack := os.Getenv("CF_STACK")
		if stack == "cflinuxfs2" {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "nancy_kestrel_msbuild_dotnet1"))
		} else if stack == "cflinuxfs3" {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "nancy_kestrel_msbuild_dotnet2"))
		} else {
			Skip(fmt.Sprintf("Skip deployment of Nancy app on unknown stack: %s", stack))
		}
	})

	It("displays a simple text homepage", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})
})
