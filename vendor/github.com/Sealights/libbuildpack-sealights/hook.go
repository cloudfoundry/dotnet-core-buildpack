package sealights

import (
	"io"
	"os"

	"github.com/cloudfoundry/libbuildpack"
)

// Command is an interface around libbuildpack.Command. Represents an executor for external command calls. We have it
// as an interface so that we can mock it and use in the unit tests.
type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
}

// SealightsHook implements libbuildpack.Hook. It downloads and install the Sealights OneAgent.
type SealightsHook struct {
	libbuildpack.DefaultHook
	Log     *libbuildpack.Logger
	Command Command
}

// NewHook returns a libbuildpack.Hook instance for integrating with Sealights
func NewHook() libbuildpack.Hook {
	return &SealightsHook{
		Log:     libbuildpack.NewLogger(os.Stdout),
		Command: &libbuildpack.Command{},
	}
}

// AfterCompile downloads and installs the Sealights agent, and modify application start command
func (h *SealightsHook) AfterCompile(stager *libbuildpack.Stager) error {

	h.Log.Debug("Sealights. Check service status...")

	conf := NewConfiguration(h.Log, stager)
	if !conf.UseSealights() {
		h.Log.Debug("Sealights service isn't configured")
		return nil
	}

	h.Log.Info("Sealights. Service is enabled")

	agentInstaller := NewAgentInstaller(h.Log, conf.Value)

	agentDir, agentVersion, err := agentInstaller.InstallAgent(stager)
	if err != nil {
		return err
	}
	h.Log.Info("Sealights. Agent is installed (version: %s)", agentVersion)

	launcher := NewLauncher(h.Log, conf.Value, agentDir, stager)
	launcher.ModifyStartParameters(stager)

	h.Log.Info("Sealights. Service is set up")

	return nil
}
