package integration_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDynatrace(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		app      *cutlass.App
		services []string
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "simple"))
		app.SetEnv("BP_DEBUG", "true")
		PushAppAndConfirm(t, app)
	})

	it.After(func() {
		app = DestroyApp(t, app)

		for _, service := range services {
			command := exec.Command("cf", "delete-service", "-f", service)
			_, err := command.Output()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	context("deploying a .NET Core app with Dynatrace agent with single credentials service", func() {
		it("checks if Dynatrace injection was successful", func() {
			service := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", service, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))
			services = append(services, service)

			command = exec.Command("cf", "bind-service", app.Name, service)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			command = exec.Command("cf", "restage", app.Name)
			output, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent"))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying a .Net Core app with Dynatrace agent with configured network zone", func() {
		it("checks if Dynatrace injection was successful", func() {
			serviceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", serviceName, "-p", fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\", \"networkzone\":\"testzone\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, serviceName)

			command = exec.Command("cf", "bind-service", app.Name, serviceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent"))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Setting DT_NETWORK_ZONE..."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with two credentials services", func() {
		it("checks if detection of second service with credentials works", func() {
			credentialsServiceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", credentialsServiceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, credentialsServiceName)

			duplicateCredentialsServiceName := "dynatrace-dupe-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", duplicateCredentialsServiceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, duplicateCredentialsServiceName)

			command = exec.Command("cf", "bind-service", app.Name, credentialsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			command = exec.Command("cf", "bind-service", app.Name, duplicateCredentialsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.Stdout.String()).To(ContainSubstring("More than one matching service found!"))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with failing agent download and ignoring errors", func() {
		it("checks if skipping download errors works", func() {
			credentialsServiceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", credentialsServiceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s/no-such-endpoint\",\"environmentid\":\"envid\",\"skiperrors\":\"true\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, credentialsServiceName)

			command = exec.Command("cf", "bind-service", app.Name, credentialsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.Stdout.String()).To(ContainSubstring("Download returned with status 404"))
			Expect(app.Stdout.String()).To(ContainSubstring("Error during installer download, skipping installation"))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with two dynatrace services", func() {
		it("check if service detection isn't disturbed by a service with tags", func() {
			credentialsServiceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", credentialsServiceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, credentialsServiceName)

			tagsServiceName := "dynatrace-tags-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", tagsServiceName, "-p", "'{\"tag:dttest\":\"dynatrace_test\"}'")
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, tagsServiceName)

			command = exec.Command("cf", "bind-service", app.Name, credentialsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			command = exec.Command("cf", "bind-service", app.Name, tagsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with single credentials service and without manifest.json", func() {
		it("checks if Dynatrace injection was successful", func() {
			serviceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", serviceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, serviceName)

			command = exec.Command("cf", "bind-service", app.Name, serviceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with failing agent download and checking retry", func() {
		it.Before(func() {
			SetDefaultEventuallyTimeout(5 * time.Second)
		})

		it("checks if retrying downloads works", func() {
			credentialsServiceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", credentialsServiceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s/no-such-endpoint\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, credentialsServiceName)

			command = exec.Command("cf", "bind-service", app.Name, credentialsServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			command = exec.Command("cf", "restage", app.Name)
			_, err = command.CombinedOutput()

			Eventually(app.Stdout.String).Should(ContainSubstring("Error during installer download, retrying in 4s"))
			Eventually(app.Stdout.String).Should(ContainSubstring("Error during installer download, retrying in 5s"))
			Eventually(app.Stdout.String).Should(ContainSubstring("Error during installer download, retrying in 7s"))
			Eventually(app.Stdout.String).Should(ContainSubstring("Download returned with status 404"))

			Eventually(app.Stdout.String).Should(ContainSubstring("Failed to compile droplet"))
		})
	})

	context("deploying a .NET Core app with Dynatrace agent with single credentials service and a redis service", func() {
		it("checks if Dynatrace injection was successful", func() {
			serviceName := "dynatrace-" + cutlass.RandStringRunes(20) + "-service"
			command := exec.Command("cf", "cups", serviceName, "-p",
				fmt.Sprintf("'{\"apitoken\":\"secretpaastoken\",\"apiurl\":\"%s\",\"environmentid\":\"envid\"}'", settings.Dynatrace.URI))
			_, err := command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, serviceName)
			command = exec.Command("cf", "bind-service", app.Name, serviceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			redisServiceName := "redis-" + cutlass.RandStringRunes(20) + "-service"
			command = exec.Command("cf", "cups", redisServiceName, "-p", "'{\"name\":\"redis\", \"credentials\":{\"db_type\":\"redis\", \"instance_administration_api\":{\"deployment_id\":\"12345asdf\", \"instance_id\":\"12345asdf\", \"root\":\"https://doesnotexi.st\"}}}'")
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())
			services = append(services, redisServiceName)
			command = exec.Command("cf", "bind-service", app.Name, redisServiceName)
			_, err = command.CombinedOutput()
			Expect(err).To(BeNil())

			command = exec.Command("cf", "restage", app.Name)
			_, err = command.Output()
			Expect(err).To(BeNil())

			Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent."))
			Expect(app.Stdout.String()).To(ContainSubstring("Starting Dynatrace OneAgent installer"))
			Expect(app.Stdout.String()).To(ContainSubstring("Copy dynatrace-env.sh"))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent installed."))
			Expect(app.Stdout.String()).To(ContainSubstring("Dynatrace OneAgent injection is set up."))
		})
	})
}
