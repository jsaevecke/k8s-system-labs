package kubectl

import (
	"context"
	"fmt"
)

func (c *Client) Apply(ctx context.Context, kubeconfig, manifestPath string) error {
	if err := c.run(ctx, "--kubeconfig", kubeconfig, "apply", "--filename", manifestPath); err != nil {
		return fmt.Errorf("apply manifest %q: %w", manifestPath, err)
	}
	return nil
}
