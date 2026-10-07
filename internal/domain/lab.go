package domain

import (
	"fmt"
	"path/filepath"
	"time"
)

const LabAPIVersion = "labs.k8s-system-labs/v1alpha1"

type Lab struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
	SourcePath string   `yaml:"-"`
}

type Spec struct {
	Providers ProviderSelection `yaml:"providers"`
	Cluster   ClusterConfig     `yaml:"cluster"`
	Manifests []string          `yaml:"manifests"`
	Ready     Readiness         `yaml:"ready"`
}

func (l Lab) ReadinessTimeout() time.Duration {
	timeout, _ := time.ParseDuration(l.Spec.Ready.Timeout)
	return timeout
}

func (l Lab) ManifestPath(manifest string) string {
	if filepath.IsAbs(manifest) {
		return manifest
	}
	return filepath.Join(filepath.Dir(l.SourcePath), manifest)
}

func (l Lab) Validate() error {
	if l.APIVersion != LabAPIVersion {
		return fmt.Errorf("unsupported apiVersion %q", l.APIVersion)
	}
	if l.Kind != "Lab" {
		return fmt.Errorf("unsupported kind %q", l.Kind)
	}
	if l.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if l.Spec.Providers.Lab == "" {
		return fmt.Errorf("spec.providers.lab is required")
	}
	if l.Spec.Providers.Cluster == "" {
		return fmt.Errorf("spec.providers.cluster is required")
	}
	if l.Spec.Cluster.Name == "" {
		return fmt.Errorf("spec.cluster.name is required")
	}
	if l.Spec.Cluster.Workers < 0 {
		return fmt.Errorf("spec.cluster.workers cannot be negative")
	}
	if len(l.Spec.Manifests) == 0 {
		return fmt.Errorf("spec.manifests must contain at least one manifest")
	}
	if timeout, err := time.ParseDuration(l.Spec.Ready.Timeout); err != nil || timeout <= 0 {
		return fmt.Errorf("spec.ready.timeout must be a positive duration")
	}
	if l.Spec.Ready.Pod.Namespace == "" || l.Spec.Ready.Pod.Name == "" || l.Spec.Ready.Pod.WaitingReason == "" {
		return fmt.Errorf("spec.ready.pod namespace, name, and waitingReason are required")
	}
	return nil
}
