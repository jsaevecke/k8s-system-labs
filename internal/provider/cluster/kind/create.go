package kind

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
)

func (p *Provider) Create(ctx context.Context, config domain.ClusterConfig) (domain.Cluster, error) {
	runDirectory := filepath.Join(p.stateRoot, config.Name)
	if err := os.MkdirAll(runDirectory, 0o700); err != nil {
		return domain.Cluster{}, fmt.Errorf("create lab state directory: %w", err)
	}

	kindConfigPath := filepath.Join(runDirectory, "kind.yaml")
	kubeconfigPath := filepath.Join(runDirectory, "kubeconfig")
	if err := os.WriteFile(kindConfigPath, renderConfig(config.Workers), 0o600); err != nil {
		return domain.Cluster{}, fmt.Errorf("write Kind configuration: %w", err)
	}

	command := exec.CommandContext(
		ctx,
		p.binary,
		"create", "cluster",
		"--name", config.Name,
		"--config", kindConfigPath,
		"--kubeconfig", kubeconfigPath,
	)
	command.Stdout = p.stdout
	command.Stderr = p.stderr
	if err := command.Run(); err != nil {
		return domain.Cluster{}, fmt.Errorf("create Kind cluster %q: %w", config.Name, err)
	}
	return p.Resolve(config), nil
}

func renderConfig(workers int) []byte {
	var config bytes.Buffer
	config.WriteString("kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nnodes:\n- role: control-plane\n")
	for range workers {
		config.WriteString("- role: worker\n")
	}
	return config.Bytes()
}
