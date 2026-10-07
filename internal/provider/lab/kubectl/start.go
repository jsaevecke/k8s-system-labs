package kubectl

import (
	"context"
	"fmt"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	"github.com/jsaevecke/k8s-system-labs/internal/logging"
)

func (l *Provider) Start(ctx context.Context, definition domain.Lab, cluster domain.Cluster) error {
	for _, manifest := range definition.Spec.Manifests {
		manifestPath := definition.ManifestPath(manifest)
		l.logger.InfoContext(ctx, "applying manifest", logging.FieldPath, manifestPath)
		if err := l.client.Apply(ctx, cluster.KubeconfigPath, manifestPath); err != nil {
			return fmt.Errorf("start lab %q: %w", definition.Metadata.Name, err)
		}
	}

	ready := definition.Spec.Ready.Pod
	l.logger.InfoContext(ctx, "waiting for Pod readiness",
		logging.FieldNamespace, ready.Namespace,
		logging.FieldPod, ready.Name,
		logging.FieldWaitingReason, ready.WaitingReason,
	)
	message, err := l.client.WaitForPodReason(
		ctx,
		cluster.KubeconfigPath,
		ready.Namespace,
		ready.Name,
		ready.WaitingReason,
		definition.ReadinessTimeout(),
	)
	if err != nil {
		return fmt.Errorf("start lab %q: %w", definition.Metadata.Name, err)
	}

	l.logger.InfoContext(ctx, "lab ready",
		logging.FieldLab, definition.Metadata.Name,
		logging.FieldNamespace, ready.Namespace,
		logging.FieldPod, ready.Name,
		logging.FieldWaitingReason, ready.WaitingReason,
		logging.FieldMessage, message,
	)
	return nil
}
