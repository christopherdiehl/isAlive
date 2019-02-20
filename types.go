package main

const (
	unknown = 0
	healthy = 1
	timeout = 2
	failed  = 3
)

type Host struct {
	Endpoint string
	Status   int
}

func CreateHost(endpoint string) *Host {
	return &Host{
		Endpoint: endpoint,
		Status:   unknown,
	}
}
