package netpols

import (
	"net/http"

	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
)

type Config struct {
	KubeConfig   string
	LogLevel     string
	ListenHost   string
	ListenPort   string
	WWWDirectory string
}

// Pod wraps core.Pod to bind applied network policies
type Pod struct {
	Pod                    core.Pod
	Namespace              core.Namespace
	AppliedNetworkPolicies []networking.NetworkPolicy
}

type Server struct {
	http.Server
	cfg               *Config
	alreadyRefreshing bool
	netpols           *[]networking.NetworkPolicy
	payload           Payload
	pods              *PodList
}

// Connection struct holds information from which pod to which other connections are allowed by what NetworkPolicy
type Connection struct {
	FromUID          string
	ToUID            string
	NetworkPolicyUID string
}

// Payload is return as JSON
type Payload struct {
	Pods               map[string][]core.Pod
	NetworkPolicies    []networking.NetworkPolicy
	AllowedConnections []Connection
}
