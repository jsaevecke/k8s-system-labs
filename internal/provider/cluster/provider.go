package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	kindprovider "github.com/jsaevecke/k8s-system-labs/internal/provider/cluster/kind"
)

var ErrUnsupportedProvider = errors.New("unsupported cluster provider")

type Provider interface {
	Create(context.Context, domain.ClusterConfig) (domain.Cluster, error)
	Delete(context.Context, domain.ClusterConfig) error
	Resolve(domain.ClusterConfig) domain.Cluster
}

func New(name domain.ClusterProvider, stateRoot string, stdout, stderr io.Writer) (Provider, error) {
	switch name {
	case domain.ClusterProviderKind:
		return kindprovider.New(stateRoot, stdout, stderr)
	default:
		return nil, fmt.Errorf("%w %q", ErrUnsupportedProvider, name)
	}
}
