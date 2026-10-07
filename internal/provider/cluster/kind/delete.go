package kind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
)

func (p *Provider) Delete(ctx context.Context, config domain.ClusterConfig) error {
	command := exec.CommandContext(ctx, p.binary, "delete", "cluster", "--name", config.Name)
	command.Stdout = p.stdout
	command.Stderr = p.stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("delete Kind cluster %q: %w", config.Name, err)
	}
	if err := os.RemoveAll(filepath.Join(p.stateRoot, config.Name)); err != nil {
		return fmt.Errorf("remove lab state: %w", err)
	}
	return nil
}
