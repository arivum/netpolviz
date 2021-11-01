package netpols

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (s *Server) refresh(clientset *kubernetes.Clientset) error {
	var (
		err               error
		netpol            networking.NetworkPolicy
		netpolPtr         *networking.NetworkPolicy
		from, to          string
		netpolJSON        []byte
		connectionAllowed bool
	)

	if s.alreadyRefreshing {
		return nil
	}
	s.netpols = nil
	s.pods = nil
	s.alreadyRefreshing = true
	log.Infof("Refreshing k8s API resources")
	// Get all NetworkPolicies except those deployed to kube-system
	s.netpols, err = GetNetworkPoliciesExcludeNS(clientset, []string{"kube-system"})
	if err != nil {
		log.Error(err.Error())
		s.alreadyRefreshing = false
		return err
	}
	log.Debugf("There are %d netpols in the cluster\n", len(*s.netpols))
	for _, netpol = range *s.netpols {
		netpolJSON, _ = json.MarshalIndent(netpol.Spec, "", "  ")
		log.Debug(string(netpolJSON))
	}

	// Get all namespaces
	namespaces, err := GetNamespaces(clientset)
	if err != nil {
		log.Error(err.Error())
		s.alreadyRefreshing = false
		return err
	}
	log.Debugf("There are %d namespaces in the cluster\n", len(*namespaces))
	for _, namespace := range *namespaces {
		namespaceJSON, _ := json.MarshalIndent(namespace.Spec, "", "  ")
		log.Debug(string(namespaceJSON))
	}

	// Get all Pods except those deployed to kube-system
	s.pods, err = GetPodsExcludeNS(clientset, []string{"kube-system"})
	if err != nil {
		log.Error(err.Error())
		s.alreadyRefreshing = false
		return err
	}
	log.Debugf("There are %d pods in the cluster\n", len(*s.pods))
	s.pods.SetCorrespondingNamespaces(*namespaces)
	s.pods.SetAppliedNetworkPolicies(*s.netpols)

	for _, pod := range *s.pods {
		podJSON, _ := json.MarshalIndent(pod, "", "  ")
		log.Debug(string(podJSON))
	}

	s.payload = Payload{
		NetworkPolicies:    *s.netpols,
		Pods:               make(map[string][]core.Pod),
		AllowedConnections: []Connection{},
	}
	// Find allowed Connections by iterating all Pods
	for _, pod1 := range *s.pods {
		if _, ok := s.payload.Pods[pod1.Pod.Namespace]; !ok {
			s.payload.Pods[pod1.Pod.Namespace] = make([]core.Pod, 0)
		}
		s.payload.Pods[pod1.Pod.Namespace] = append(s.payload.Pods[pod1.Pod.Namespace], pod1.Pod)
		for _, pod2 := range *s.pods {
			connectionAllowed, netpolPtr = pod1.PodAllowedToConnect(pod2)
			if connectionAllowed && fmt.Sprintf("%s.%s", pod1.Pod.Name, pod1.Pod.Namespace) != fmt.Sprintf("%s.%s", pod2.Pod.Name, pod2.Pod.Namespace) {
				from = pod2.Pod.Name
				to = pod1.Pod.Name
				if len(pod2.Pod.ObjectMeta.OwnerReferences) > 0 {
					from = string(pod2.Pod.ObjectMeta.OwnerReferences[0].UID)
				}
				if len(pod1.Pod.ObjectMeta.OwnerReferences) > 0 {
					to = string(pod1.Pod.ObjectMeta.OwnerReferences[0].UID)
				}
				s.payload.AllowedConnections = append(s.payload.AllowedConnections, Connection{
					FromUID:          string(from),
					ToUID:            string(to),
					NetworkPolicyUID: string(netpolPtr.UID),
				})
			}
		}
	}
	log.Info("Successfully refreshed k8s API resources")
	s.alreadyRefreshing = false
	return nil
}

func (s *Server) watch(clientset *kubernetes.Clientset) {
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientset, 0)

	// UpdateFuncs are used to trigger refresh on k8s resource update
	updatedFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			go s.refresh(clientset)
		},
		DeleteFunc: func(obj interface{}) {
			go s.refresh(clientset)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			go s.refresh(clientset)
		},
	}

	// Watch for updated Pods
	podInformer := kubeInformerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(updatedFuncs)

	// Watch for updated NetworkPolicies
	netpolInformer := kubeInformerFactory.Networking().V1().NetworkPolicies().Informer()
	netpolInformer.AddEventHandler(updatedFuncs)

	// Watch for updated Namespaces
	namespaceInformer := kubeInformerFactory.Core().V1().Namespaces().Informer()
	namespaceInformer.AddEventHandler(updatedFuncs)

	stop := make(chan struct{})
	defer close(stop)
	kubeInformerFactory.Start(stop)
	for {
		time.Sleep(time.Second)
	}
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	payloadJSON, err := json.Marshal(&s.payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(payloadJSON)
}

func NewServer(cfg *Config) (*Server, error) {
	var (
		server = &Server{
			Server: http.Server{
				Addr: fmt.Sprintf("%s:%s", cfg.ListenHost, cfg.ListenPort),
			},
			cfg: cfg,
		}
		mux       = http.NewServeMux()
		err       error
		clientset *kubernetes.Clientset
	)

	SetLogLevel(cfg.LogLevel)

	if clientset, err = BuildK8SContext(cfg.KubeConfig); err != nil {
		return nil, err
	}

	if err = server.refresh(clientset); err != nil {
		return nil, err
	}

	go server.watch(clientset)

	// Serve files
	mux.Handle("/", http.FileServer(http.Dir(cfg.WWWDirectory)))

	// Serve simple api
	mux.HandleFunc("/api", server.handler)

	server.Handler = mux
	return server, nil
}
