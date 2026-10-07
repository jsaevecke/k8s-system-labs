package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
	clusterprovider "github.com/jsaevecke/k8s-system-labs/internal/provider/cluster"
	labprovider "github.com/jsaevecke/k8s-system-labs/internal/provider/lab"
)

func TestRunRejectsUnsupportedClusterProvider(t *testing.T) {
	root := writeLab(t, "unsupported-cluster", domain.ClusterProvider("unsupported"), domain.LabProviderKubectl)

	err := run(
		context.Background(),
		[]string{"labctl", "start", "unsupported-cluster"},
		io.Discard,
		executableFound,
		workingDirectory(root),
		discardLogger(),
	)
	if !errors.Is(err, clusterprovider.ErrUnsupportedProvider) {
		t.Fatalf("expected unsupported cluster provider error, got %v", err)
	}
}

func TestRunRejectsUnsupportedLabProvider(t *testing.T) {
	root := writeLab(t, "unsupported-lab", domain.ClusterProviderKind, domain.LabProvider("unsupported"))

	err := run(
		context.Background(),
		[]string{"labctl", "start", "unsupported-lab"},
		io.Discard,
		executableFound,
		workingDirectory(root),
		discardLogger(),
	)
	if !errors.Is(err, labprovider.ErrUnsupportedProvider) {
		t.Fatalf("expected unsupported lab provider error, got %v", err)
	}
}

func TestRunListsLabsWithoutKubectl(t *testing.T) {
	root := writeLab(t, "listed-lab", domain.ClusterProviderKind, domain.LabProviderKubectl)
	lookupCalled := false
	lookPath := func(string) (string, error) {
		lookupCalled = true
		return "", nil
	}
	var output bytes.Buffer

	err := run(
		context.Background(),
		[]string{"labctl", "list"},
		&output,
		lookPath,
		workingDirectory(root),
		discardLogger(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if lookupCalled {
		t.Fatal("lab listing unexpectedly looked up kubectl")
	}
	for _, value := range []string{"listed-lab", domain.ClusterProviderKind.String(), domain.LabProviderKubectl.String()} {
		if !strings.Contains(output.String(), value) {
			t.Fatalf("lab list did not contain %q: %s", value, output.String())
		}
	}
}

func TestRunRequiresKubectlBeforeDiscoveringLabs(t *testing.T) {
	missingKubectl := errors.New(domain.LabProviderKubectl.String() + " missing")
	lookPath := func(string) (string, error) {
		return "", missingKubectl
	}
	getwdCalled := false
	getwd := func() (string, error) {
		getwdCalled = true
		return "", nil
	}

	err := run(
		context.Background(),
		[]string{"labctl", "start", "missing-lab"},
		io.Discard,
		lookPath,
		getwd,
		discardLogger(),
	)
	if !errors.Is(err, missingKubectl) {
		t.Fatalf("expected missing kubectl error, got %v", err)
	}
	if getwdCalled {
		t.Fatal("lab discovery happened before kubectl lookup")
	}
}

func TestRunChecksArgumentsBeforeDependencies(t *testing.T) {
	lookupCalled := false
	lookPath := func(string) (string, error) {
		lookupCalled = true
		return "", nil
	}
	getwdCalled := false
	getwd := func() (string, error) {
		getwdCalled = true
		return "", nil
	}

	err := run(context.Background(), []string{"labctl"}, io.Discard, lookPath, getwd, discardLogger())
	if err == nil {
		t.Fatal("expected argument error")
	}
	if lookupCalled || getwdCalled {
		t.Fatal("dependency lookup happened before argument validation")
	}
}

func writeLab(t *testing.T, name string, clusterProvider domain.ClusterProvider, labProvider domain.LabProvider) string {
	t.Helper()
	root := t.TempDir()
	labDirectory := filepath.Join(root, "labs", name)
	if err := os.MkdirAll(labDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := []byte(fmt.Sprintf(`apiVersion: labs.k8s-system-labs/v1alpha1
kind: Lab
metadata:
  name: %s
spec:
  providers:
    cluster: %s
    lab: %s
  cluster:
    name: %s
    workers: 0
  manifests:
    - pod.yaml
  ready:
    timeout: 1m
    pod:
      namespace: default
      name: broken-image
      waitingReason: ImagePullBackOff
`, name, clusterProvider, labProvider, name))
	if err := os.WriteFile(filepath.Join(labDirectory, "lab.yaml"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func executableFound(name string) (string, error) {
	return name, nil
}

func workingDirectory(path string) func() (string, error) {
	return func() (string, error) {
		return path, nil
	}
}
