package lab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	kubectlclient "github.com/jsaevecke/k8s-system-labs/internal/kubernetes/kubectl"
	kubectllab "github.com/jsaevecke/k8s-system-labs/internal/provider/lab/kubectl"
)

var ErrUnsupportedProvider = errors.New("unsupported lab provider")

type Provider interface {
	Start(context.Context, domain.Lab, domain.Cluster) error
	Delete(context.Context, domain.Lab, domain.Cluster) error
}

func New(name domain.LabProvider, logger *slog.Logger, client *kubectlclient.Client) (Provider, error) {
	switch name {
	case domain.LabProviderKubectl:
		return kubectllab.New(logger, client), nil
	default:
		return nil, fmt.Errorf("%w %q", ErrUnsupportedProvider, name)
	}
}
