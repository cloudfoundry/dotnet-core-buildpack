package integration_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testMultipleProjects(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		it("compiles both apps", func() {
			deployment, _, err := platform.Deploy.
				Execute(name, filepath.Join(fixtures, "source_apps", "multiple_projects_msbuild"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(deployment).Should(Serve(ContainSubstring("Hello, I'm a string!")))

			cmd := exec.Command("docker", "container", "logs", deployment.Name)

			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Hello from a secondary project!"))
		})

		context("Deploying a self-contained solution with multiple projects", func() {
			it("can run the app", func() {
				deployment, _, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "self_contained_apps", "self_contained_solution_2.2"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})
	}
}
