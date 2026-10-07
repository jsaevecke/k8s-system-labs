package domain

type ClusterProvider string

const (
	ClusterProviderKind ClusterProvider = "kind"
)

func (p ClusterProvider) String() string {
	return string(p)
}

type LabProvider string

const (
	LabProviderKubectl LabProvider = "kubectl"
)

func (p LabProvider) String() string {
	return string(p)
}

type ProviderSelection struct {
	Cluster ClusterProvider `yaml:"cluster"`
	Lab     LabProvider     `yaml:"lab"`
}
