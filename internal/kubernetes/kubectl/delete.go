package kubectl

import (
	"context"
	"fmt"
)

func (c *Client) Delete(ctx context.Context, kubeconfig, manifestPath string) error {
	if err := c.run(ctx, "--kubeconfig", kubeconfig, "delete", "--filename", manifestPath, "--ignore-not-found=true"); err != nil {
		return fmt.Errorf("delete manifest %q: %w", manifestPath, err)
	}
	return nil
}
