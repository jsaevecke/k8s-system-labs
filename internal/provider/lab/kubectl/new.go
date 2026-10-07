package kubectl

import (
	"log/slog"

	kubectlclient "github.com/jsaevecke/k8s-system-labs/internal/kubernetes/kubectl"
)

type Provider struct {
	logger *slog.Logger
	client *kubectlclient.Client
}

func New(client *kubectlclient.Client, logger *slog.Logger) *Provider {
	return &Provider{logger: logger, client: client}
}
