package domain

type ClusterConfig struct {
	Name    string `yaml:"name"`
	Workers int    `yaml:"workers"`
}

type Cluster struct {
	KubeconfigPath string
}
