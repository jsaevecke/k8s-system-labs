package kind

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
)

type Provider struct {
	binary    string
	stateRoot string
	stdout    io.Writer
	stderr    io.Writer
}

func New(stateRoot string, stdout, stderr io.Writer) (*Provider, error) {
	binary, err := exec.LookPath(domain.ClusterProviderKind.String())
	if err != nil {
		return nil, fmt.Errorf("find %s executable: %w", domain.ClusterProviderKind, err)
	}
	return &Provider{
		binary:    binary,
		stateRoot: stateRoot,
		stdout:    stdout,
		stderr:    stderr,
	}, nil
}
