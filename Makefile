.PHONY: build

build:
	CGO_ENALED=0 go build github.com/arivum/netpolviz

IMG ?= "netpolviz"

pack:
	docker build -t $(IMG) -f build/docker/Dockerfile .

k8s-deploy-test:
	kubectl create namespace netpol-test || true ; \
	kubectl create namespace test || true ; \
	kubectl create namespace test2 || true ; \
	kubectl label --overwrite namespace netpol-test name=netpol-test ; \
	for namespace in "default" "netpol-test" "test" "test2"; do kubectl --namespace $$namespace apply -f tests/k8s/ns-$$namespace-deployment.yaml; done

k8s-undeploy-test:
	for namespace in "default" "netpol-test" "test" "test2"; do kubectl --namespace $$namespace delete -f tests/k8s/ns-$$namespace-deployment.yaml; done ; \
	kubectl delete namespace netpol-test || true ; \
	kubectl delete namespace test || true ; \
	kubectl delete namespace test2 || true

k8s-test:
	kubectl --namespace netpol-test exec -it svc/netpoltest -- curl --connect-timeout 1 netpoltest.default > /dev/null 2>&1 && echo Successfully connected to netpoltest in namespace [default] || echo Cannot connect to netpoltest in namespace [default] ; \
	kubectl --namespace default exec -it svc/netpoltest -- curl --connect-timeout 1 netpoltest.netpol-test > /dev/null 2>&1 && echo Successfully connected to netpoltest in namespace [netpol-test] || echo Cannot connect to netpoltest in namespace [netpol-test]
