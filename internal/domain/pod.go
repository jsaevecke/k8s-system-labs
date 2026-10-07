package domain

type Readiness struct {
	Timeout string       `yaml:"timeout"`
	Pod     PodReadiness `yaml:"pod"`
}

type PodReadiness struct {
	Namespace     string `yaml:"namespace"`
	Name          string `yaml:"name"`
	WaitingReason string `yaml:"waitingReason"`
}
