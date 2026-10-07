package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/jsaevecke/k8s-system-labs/cmd/labctl/commands"
	"github.com/jsaevecke/k8s-system-labs/internal/catalog"
	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	"github.com/jsaevecke/k8s-system-labs/internal/kubernetes/kubectl"
	yamlloader "github.com/jsaevecke/k8s-system-labs/internal/loader/yaml"
	"github.com/jsaevecke/k8s-system-labs/internal/logging"
	clusterprovider "github.com/jsaevecke/k8s-system-labs/internal/provider/cluster"
	labprovider "github.com/jsaevecke/k8s-system-labs/internal/provider/lab"
)

const usage = "usage: labctl <list|start|delete> [lab-name]"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(ctx, os.Args, os.Stdout, os.Stderr, exec.LookPath, os.Getwd, logger); err != nil {
		logger.ErrorContext(ctx, "labctl failed", logging.FieldError, err)
		stop()
		os.Exit(1)
	}
	stop()
}

func run(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	lookPath func(string) (string, error),
	getwd func() (string, error),
	logger *slog.Logger,
) error {
	command, labName, err := parseArguments(args)
	if err != nil {
		return err
	}

	if command == commands.List {
		labCatalog, err := catalog.Find(getwd)
		if err != nil {
			return err
		}
		labs, err := labCatalog.List()
		if err != nil {
			return err
		}
		return labs.Print(stdout)
	}

	kubectlBinary, err := lookPath(domain.LabProviderKubectl.String())
	if err != nil {
		return fmt.Errorf("find %s executable: %w", domain.LabProviderKubectl, err)
	}

	labCatalog, err := catalog.Find(getwd)
	if err != nil {
		return err
	}
	definitionPath, err := labCatalog.Resolve(labName)
	if err != nil {
		return err
	}
	definition, err := yamlloader.Load(definitionPath)
	if err != nil {
		return err
	}

	stateRoot := filepath.Join(filepath.Dir(labCatalog.Root), ".labctl")

	labProvider, err := labprovider.New(definition.Spec.Providers.Lab, kubectl.New(kubectlBinary, stdout, stderr), logger)
	if err != nil {
		return err
	}

	clusterProvider, err := clusterprovider.New(definition.Spec.Providers.Cluster, stateRoot, stdout, stderr)
	if err != nil {
		return err
	}

	switch command {
	case commands.Start:
		logger.InfoContext(ctx, "provisioning cluster",
			logging.FieldCluster, definition.Spec.Cluster.Name,
			logging.FieldClusterProvider, definition.Spec.Providers.Cluster.String(),
		)

		cluster, err := clusterProvider.Create(ctx, definition.Spec.Cluster)
		if err != nil {
			return err
		}

		logger.InfoContext(ctx, "starting lab",
			logging.FieldLab, definition.Metadata.Name,
			logging.FieldLabProvider, definition.Spec.Providers.Lab.String(),
		)

		return labProvider.Start(ctx, definition, cluster)
	case commands.Delete:
		cluster := clusterProvider.Resolve(definition.Spec.Cluster)

		logger.InfoContext(ctx, "deleting lab",
			logging.FieldLab, definition.Metadata.Name,
			logging.FieldLabProvider, definition.Spec.Providers.Lab.String(),
		)

		labErr := labProvider.Delete(ctx, definition, cluster)

		logger.InfoContext(ctx, "deleting cluster",
			logging.FieldCluster, definition.Spec.Cluster.Name,
			logging.FieldClusterProvider, definition.Spec.Providers.Cluster.String(),
		)

		clusterErr := clusterProvider.Delete(ctx, definition.Spec.Cluster)

		return errors.Join(labErr, clusterErr)
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}

func parseArguments(args []string) (commands.Command, string, error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("%s", usage)
	}

	command, supported := commands.Parse(args[1])
	if !supported {
		return "", "", fmt.Errorf("%s", usage)
	}

	if command == commands.List {
		if len(args) != 2 {
			return "", "", fmt.Errorf("%s", usage)
		}
		return command, "", nil
	}

	if len(args) != 3 {
		return "", "", fmt.Errorf("%s", usage)
	}
	return command, args[2], nil
}
