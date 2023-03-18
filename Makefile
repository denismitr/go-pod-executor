MASTER_URL := $(shell cat ~/.kube/config | grep 127 | awk -F 'server: ' '/server: /{print $$2}')
KUBE_CONFIG := $(shell cat ~/.kube/config | base64)

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: test
test: export K8S_MASTER_URL=$(MASTER_URL)
test: export K8S_CONFIG=$(KUBE_CONFIG)
test:
	@echo running integtation tests on $(K8S_MASTER_URL)
	go test -v ./... -count=1