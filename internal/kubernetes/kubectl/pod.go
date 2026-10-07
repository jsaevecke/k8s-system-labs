package kubectl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type podStatus struct {
	Status struct {
		Phase             string            `json:"phase"`
		ContainerStatuses []containerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type containerStatus struct {
	State struct {
		Waiting *waitingState `json:"waiting"`
	} `json:"state"`
}

type waitingState struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (c *Client) WaitForPodReason(ctx context.Context, kubeconfig, namespace, name, wantedReason string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastObservation := "Pod status was not available"
	for {
		reason, message, err := c.getPodWaitingReason(ctx, kubeconfig, namespace, name)
		if err != nil {
			lastObservation = err.Error()
		} else if reason != "" {
			lastObservation = reason + ": " + message
			if reason == wantedReason {
				return message, nil
			}
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return "", fmt.Errorf("wait for Pod %s/%s to report %s: timed out after %s; last observation: %s", namespace, name, wantedReason, timeout, lastObservation)
			}
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *Client) getPodWaitingReason(ctx context.Context, kubeconfig, namespace, name string) (string, string, error) {
	command := exec.CommandContext(ctx, c.binary, "--kubeconfig", kubeconfig, "get", "pod", name, "--namespace", namespace, "--output", "json")
	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", "", fmt.Errorf("kubectl get pod: %s", exitError.Stderr)
		}
		return "", "", fmt.Errorf("kubectl get pod: %w", err)
	}

	var pod podStatus
	if err := json.Unmarshal(output, &pod); err != nil {
		return "", "", fmt.Errorf("decode Pod status: %w", err)
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil {
			return container.State.Waiting.Reason, container.State.Waiting.Message, nil
		}
	}
	return "", "Pod phase is " + pod.Status.Phase, nil
}
