package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	"github.com/jsaevecke/k8s-system-labs/internal/logging"
)

func (l *Provider) Delete(ctx context.Context, definition domain.Lab, cluster domain.Cluster) error {
	if _, err := os.Stat(cluster.KubeconfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read cluster kubeconfig: %w", err)
	}

	var result error
	for index := len(definition.Spec.Manifests) - 1; index >= 0; index-- {
		manifestPath := definition.ManifestPath(definition.Spec.Manifests[index])
		l.logger.InfoContext(ctx, "deleting manifest", logging.FieldPath, manifestPath)
		if err := l.client.Delete(ctx, cluster.KubeconfigPath, manifestPath); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}
