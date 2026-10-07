package kind

import (
	"path/filepath"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
)

func (p *Provider) Resolve(config domain.ClusterConfig) domain.Cluster {
	return domain.Cluster{
		KubeconfigPath: filepath.Join(p.stateRoot, config.Name, "kubeconfig"),
	}
}
