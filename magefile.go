//go:build mage

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	"github.com/jsaevecke/k8s-system-labs/internal/logging"
)

const kindVersion = "v0.33.0"

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

// Check runs the repository's tests and static analysis.
func Check() error {
	if err := runCommand("go", "test", "./..."); err != nil {
		return err
	}
	return runCommand("go", "vet", "./...")
}

// Build compiles labctl into .labctl/build.
func Build() error {
	destination := filepath.Join(".labctl", "build", executableName("labctl"))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create build directory: %w", err)
	}
	if err := runCommand("go", "build", "-o", destination, "./cmd/labctl"); err != nil {
		return err
	}
	logger.Info("built labctl", logging.FieldPath, destination)
	return nil
}

// Providers installs tools for the selected providers and verifies prerequisites.
func Providers(cluster *string, lab *string) error {
	clusterProvider, labProvider, err := selectProviders(cluster, lab)
	if err != nil {
		return err
	}
	return prepareProviders(clusterProvider, labProvider)
}

// Install prepares providers, checks the repository, installs labctl, and lists the labs.
func Install(cluster *string, lab *string) error {
	clusterProvider, labProvider, err := selectProviders(cluster, lab)
	if err != nil {
		return err
	}
	if err := prepareProviders(clusterProvider, labProvider); err != nil {
		return err
	}
	if err := Check(); err != nil {
		return err
	}

	installDirectory, err := goBinDirectory()
	if err != nil {
		return err
	}
	labctlPath, err := installLabctl(installDirectory)
	if err != nil {
		return err
	}
	logger.Info("installed labctl", logging.FieldPath, labctlPath)
	if !directoryOnPath(installDirectory) {
		logger.Warn("Go binary directory is not on PATH", logging.FieldPath, installDirectory)
	}
	return runCommand(labctlPath, "list")
}

// Labs lists labs through the installed labctl binary.
func Labs() error {
	labctlPath, err := exec.LookPath(executableName("labctl"))
	if err != nil {
		return fmt.Errorf("find labctl on PATH: %w", err)
	}
	return runCommand(labctlPath, "list")
}

func selectProviders(cluster *string, lab *string) (domain.ClusterProvider, domain.LabProvider, error) {
	clusterProvider := domain.ClusterProvider(selectedValue(cluster, domain.ClusterProviderKind.String()))
	labProvider := domain.LabProvider(selectedValue(lab, domain.LabProviderKubectl.String()))

	if clusterProvider != domain.ClusterProviderKind {
		return "", "", fmt.Errorf("unsupported cluster provider %q (supported: %s)", clusterProvider, domain.ClusterProviderKind)
	}
	if labProvider != domain.LabProviderKubectl {
		return "", "", fmt.Errorf("unsupported lab provider %q (supported: %s)", labProvider, domain.LabProviderKubectl)
	}
	return clusterProvider, labProvider, nil
}

func selectedValue(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func prepareProviders(clusterProvider domain.ClusterProvider, labProvider domain.LabProvider) error {
	if labProvider == domain.LabProviderKubectl {
		if _, err := exec.LookPath(executableName(domain.LabProviderKubectl.String())); err != nil {
			return fmt.Errorf("%s is required for the %s lab provider: %w", domain.LabProviderKubectl, domain.LabProviderKubectl, err)
		}
	}
	if clusterProvider == domain.ClusterProviderKind {
		installDirectory, err := goBinDirectory()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(installDirectory, 0o755); err != nil {
			return fmt.Errorf("create Go binary directory: %w", err)
		}
		logger.Info("installing cluster provider tool",
			logging.FieldClusterProvider, clusterProvider.String(),
			logging.FieldPath, installDirectory,
			logging.FieldVersion, kindVersion,
		)
		command := exec.Command("go", "install", "sigs.k8s.io/kind@"+kindVersion)
		command.Env = append(os.Environ(), "GOBIN="+installDirectory)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("install kind provider tool: %w", err)
		}
	}
	return nil
}

func installLabctl(directory string) (string, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("create Go binary directory: %w", err)
	}

	temporary, err := os.CreateTemp(directory, ".labctl-*")
	if err != nil {
		return "", fmt.Errorf("create temporary labctl executable: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary labctl executable: %w", err)
	}
	if err := os.Remove(temporaryPath); err != nil {
		return "", fmt.Errorf("prepare temporary labctl executable: %w", err)
	}
	defer os.Remove(temporaryPath)

	if err := runCommand("go", "build", "-o", temporaryPath, "./cmd/labctl"); err != nil {
		return "", err
	}
	if err := os.Chmod(temporaryPath, 0o755); err != nil {
		return "", fmt.Errorf("make labctl executable: %w", err)
	}

	destination := filepath.Join(directory, executableName("labctl"))
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("replace existing labctl executable: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return "", fmt.Errorf("install labctl executable: %w", err)
	}
	return destination, nil
}

func goBinDirectory() (string, error) {
	goBin, err := commandOutput("go", "env", "GOBIN")
	if err != nil {
		return "", err
	}
	if goBin != "" {
		return filepath.Clean(goBin), nil
	}

	goPath, err := commandOutput("go", "env", "GOPATH")
	if err != nil {
		return "", err
	}
	paths := filepath.SplitList(goPath)
	if len(paths) == 0 || paths[0] == "" {
		return "", fmt.Errorf("go env returned no GOBIN or GOPATH")
	}
	return filepath.Join(paths[0], "bin"), nil
}

func commandOutput(name string, arguments ...string) (string, error) {
	output, err := exec.Command(name, arguments...).Output()
	if err != nil {
		return "", fmt.Errorf("run %s %s: %w", name, strings.Join(arguments, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func runCommand(name string, arguments ...string) error {
	logger.Info("running command", logging.FieldCommand, name, logging.FieldArguments, arguments)
	command := exec.Command(name, arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	if err := command.Run(); err != nil {
		return fmt.Errorf("run %s %s: %w", name, strings.Join(arguments, " "), err)
	}
	return nil
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func directoryOnPath(directory string) bool {
	for _, pathEntry := range filepath.SplitList(os.Getenv("PATH")) {
		if runtime.GOOS == "windows" {
			if strings.EqualFold(filepath.Clean(pathEntry), filepath.Clean(directory)) {
				return true
			}
			continue
		}
		if filepath.Clean(pathEntry) == filepath.Clean(directory) {
			return true
		}
	}
	return false
}
