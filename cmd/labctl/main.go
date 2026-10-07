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
	"text/tabwriter"

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
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(ctx, logger, os.Args, os.Stdout, os.Stderr, exec.LookPath, os.Getwd); err != nil {
		logger.ErrorContext(ctx, "labctl failed", logging.FieldError, err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	logger *slog.Logger,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	lookPath func(string) (string, error),
	getwd func() (string, error),
) error {
	command, labName, err := parseArguments(args)
	if err != nil {
		return err
	}

	if command == commands.List {
		labCatalog, err := findCatalog(getwd)
		if err != nil {
			return err
		}
		return printLabs(stdout, labCatalog)
	}

	kubectlBinary, err := lookPath(domain.LabProviderKubectl.String())
	if err != nil {
		return fmt.Errorf("find %s executable: %w", domain.LabProviderKubectl, err)
	}

	labCatalog, err := findCatalog(getwd)
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

	stateRoot := filepath.Join(labCatalog.RepositoryRoot(), ".labctl")
	kubernetesClient := kubectl.New(kubectlBinary, stdout, stderr)
	labProvider, err := labprovider.New(definition.Spec.Providers.Lab, logger, kubernetesClient)
	if err != nil {
		return err
	}
	clusterProvider, err := clusterprovider.New(definition.Spec.Providers.Cluster, stateRoot, stdout, stderr)
	if err != nil {
		return err
	}

	if command == commands.Delete {
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
	}
	return start(ctx, logger, definition, clusterProvider, labProvider)
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

func findCatalog(getwd func() (string, error)) (*catalog.Catalog, error) {
	workingDirectory, err := getwd()
	if err != nil {
		return nil, fmt.Errorf("read working directory: %w", err)
	}
	return catalog.Find(workingDirectory)
}

func printLabs(output io.Writer, labCatalog *catalog.Catalog) error {
	labs, err := labCatalog.List()
	if err != nil {
		return err
	}

	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "NAME\tCLUSTER PROVIDER\tLAB PROVIDER\tWORKERS"); err != nil {
		return fmt.Errorf("write lab list: %w", err)
	}
	for _, definition := range labs {
		if _, err := fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%d\n",
			definition.Metadata.Name,
			definition.Spec.Providers.Cluster.String(),
			definition.Spec.Providers.Lab.String(),
			definition.Spec.Cluster.Workers,
		); err != nil {
			return fmt.Errorf("write lab list: %w", err)
		}
	}
	if err := table.Flush(); err != nil {
		return fmt.Errorf("write lab list: %w", err)
	}
	return nil
}

func start(
	ctx context.Context,
	logger *slog.Logger,
	definition domain.Lab,
	clusterProvider clusterprovider.Provider,
	labProvider labprovider.Provider,
) error {
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
}
