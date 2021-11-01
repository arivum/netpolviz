package netpols

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func contains(a []networking.PolicyType, x networking.PolicyType) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// PodAllowedToConnect returns if the given pod can connect to the current pod instance
func (p Pod) PodAllowedToConnect(pod Pod) (bool, *networking.NetworkPolicy) {
	if p.Pod.Name == pod.Pod.Name && p.Pod.Namespace == pod.Pod.Namespace {
		log.Debugf("Pod [%s.%s] allowed to connect itself\n", pod.Pod.Name, pod.Pod.Namespace)
		return true, nil
	}
	for _, netpol := range p.AppliedNetworkPolicies {
		if !contains(netpol.Spec.PolicyTypes, "Ingress") {
			continue
		}
		// if len(netpol.Spec.PodSelector.MatchLabels) == 0 && len(netpol.Spec.Ingress) == 1 && netpol.Spec.Ingress[0].Size() == 1 {
		// 	log.Infof("Pod [%s.%s] allows all incoming connections by netpol %s\n", pod.Pod.Name, pod.Pod.Namespace, netpol.Name)
		// 	return true, &netpol
		// }
		for _, rule := range netpol.Spec.Ingress {
			if rule.Size() == 0 {
				log.Infof("Netpol %s allows all incoming connections to pod [%s.%s]\n", netpol.Name, p.Pod.Name, p.Pod.Namespace)
				return true, &netpol
			}
			for _, peer := range rule.From {
				peerAllowed := true
				if peer.IPBlock != nil {
					_, ipnet, err := net.ParseCIDR(peer.IPBlock.CIDR)
					if err != nil {
						peerAllowed = peerAllowed && false
					}
					podIP := net.ParseIP(pod.Pod.Status.PodIP)
					peerAllowed = peerAllowed && ipnet.Contains(podIP)
					for _, exceptedIP := range peer.IPBlock.Except {
						if pod.Pod.Status.PodIP == exceptedIP {
							peerAllowed = peerAllowed && false
						}
					}
					log.Infof("access allowed=%v after testing %s", peerAllowed, peer)
				}
				if peer.PodSelector != nil {
					peerAllowed = peerAllowed && LabelsMatch(peer.PodSelector.MatchLabels, pod.Pod.ObjectMeta.Labels)
					log.Infof("access allowed=%v after testing %s", peerAllowed, peer)
				}
				if peer.NamespaceSelector != nil {
					peerAllowed = peerAllowed && LabelsMatch(peer.NamespaceSelector.MatchLabels, pod.Namespace.ObjectMeta.Labels)
					log.Infof("access allowed=%v after testing %s", peerAllowed, peer)
				} else {
					peerAllowed = peerAllowed && (p.Pod.ObjectMeta.Namespace == pod.Pod.ObjectMeta.Namespace)
				}
				if peerAllowed {
					log.Infof("Pod [%s.%s] allowed to connect to pod [%s.%s] by networkPolicy [%s]\n", pod.Pod.Name, pod.Pod.Namespace, p.Pod.Name, p.Pod.Namespace, netpol.Name)
					return true, &netpol
				}
			}
		}
	}
	log.Debugf("Pod [%s.%s] forbidden to connect to pod [%s.%s]\n", pod.Pod.Name, pod.Pod.Namespace, p.Pod.Name, p.Pod.Namespace)
	return false, nil
}

// PodList wraps slice of Pods to define functions upon
type PodList []Pod

// SetAppliedNetworkPolicies sets alls matching networking from given list to each Pod
func (p PodList) SetAppliedNetworkPolicies(netpols []networking.NetworkPolicy) {
	for index, pod := range p {
		for _, netpol := range netpols {
			if netpol.Namespace == pod.Pod.Namespace {
				if LabelsMatch(netpol.Spec.PodSelector.MatchLabels, pod.Pod.ObjectMeta.Labels) {
					p[index].AppliedNetworkPolicies = append(p[index].AppliedNetworkPolicies, netpol)
				}
			}
		}
	}
}

// SetCorrespondingNamespaces sets the corresponding namespace for each pod
func (p PodList) SetCorrespondingNamespaces(namespaces map[string]core.Namespace) {
	for index, pod := range p {
		p[index].Namespace = namespaces[pod.Pod.Namespace]
	}
}

// BuildK8SContext initializes k8s API context
func BuildK8SContext(kubeconfig string) (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	if kubeconfig == "in-cluster" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}

// GetNamespaces gets all Namespaces
func GetNamespaces(clientset *kubernetes.Clientset) (*map[string]core.Namespace, error) {
	namespaceMap := make(map[string]core.Namespace)
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces.Items {
		namespaceMap[namespace.Name] = namespace
	}
	return &namespaceMap, nil
}

// GetNetworkPoliciesExcludeNS gets all NetworkPolicies excepting those matching one namespace from given list
func GetNetworkPoliciesExcludeNS(clientset *kubernetes.Clientset, namespaces []string) (*[]networking.NetworkPolicy, error) {
	var netpolSlice []networking.NetworkPolicy
	netpols, err := clientset.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, netpol := range netpols.Items {
		exclude := false
		for _, namespace := range namespaces {
			if netpol.Namespace == namespace {
				exclude = true
				break
			}
		}
		if !exclude {
			netpolSlice = append(netpolSlice, netpol)
		}
	}
	return &netpolSlice, nil
}

// GetPodsExcludeNS gets all Pods excepting those matching one namespace from given list
func GetPodsExcludeNS(clientset *kubernetes.Clientset, namespaces []string) (*PodList, error) {
	var podSlice PodList
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})

	//pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		exclude := false
		for _, namespace := range namespaces {
			if pod.Namespace == namespace {
				exclude = true
				break
			}
		}
		if !exclude {
			podSlice = append(podSlice, Pod{
				Pod:                    pod,
				AppliedNetworkPolicies: []networking.NetworkPolicy{},
			})
		}
	}
	return &podSlice, nil
}
