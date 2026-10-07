package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

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
	for _, manifest := range slices.Backward(definition.Spec.Manifests) {
		manifestPath := definition.ManifestPath(manifest)
		l.logger.InfoContext(ctx, "deleting manifest", logging.FieldPath, manifestPath)
		if err := l.client.Delete(ctx, cluster.KubeconfigPath, manifestPath); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}
