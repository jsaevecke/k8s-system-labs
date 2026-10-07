package kubectl

import (
	"context"
	"io"
	"os/exec"
)

type Client struct {
	binary string
	stdout io.Writer
	stderr io.Writer
}

func New(binary string, stdout, stderr io.Writer) *Client {
	return &Client{binary: binary, stdout: stdout, stderr: stderr}
}

func (c *Client) run(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, c.binary, args...)
	command.Stdout = c.stdout
	command.Stderr = c.stderr
	return command.Run()
}
