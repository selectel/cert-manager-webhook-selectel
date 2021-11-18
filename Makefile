OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

ifeq (Darwin, $(shell uname))
	GREP_PREGEX_FLAG := E
else
	GREP_PREGEX_FLAG := P
endif

GO_VERSION ?= $(shell go mod edit -json | grep -${GREP_PREGEX_FLAG}o '"Go":\s+"([0-9.]+)"' | sed -E 's/.+"([0-9.]+)"/\1/')

IMAGE_NAME := "selectel/cert-manager-webhook-selectel"
IMAGE_TAG := "latest"

K8S_VERSION=1.21.2

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

test: _test/kubebuilder
	TEST_ASSET_ETCD=_test/kubebuilder/bin/etcd \
	TEST_ASSET_KUBE_APISERVER=_test/kubebuilder/bin/kube-apiserver \
	TEST_ASSET_KUBECTL=_test/kubebuilder/bin/kubectl \
	go test -v .

_test/kubebuilder:
	mkdir -p _test/kubebuilder
	curl -sSLo envtest-bins.tar.gz "https://go.kubebuilder.io/test-tools/${K8S_VERSION}/${OS}/${ARCH}"
	tar -C _test/kubebuilder --strip-components=1 -zvxf envtest-bins.tar.gz
	rm envtest-bins.tar.gz

clean: clean-kubebuilder

clean-kubebuilder:
	rm -Rf _test/kubebuilder

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

vendor:
	go mod vendor

golangci-lint: vendor
	@sh -c "'$(CURDIR)/scripts/golangci_lint_check.sh'"

unit-tests: vendor
	@sh -c "'$(CURDIR)/scripts/unit_tests.sh'"

.PHONY: rendered-manifest.yaml golangci-lint unit-tests
rendered-manifest.yaml:
	helm template \
	    --name cert-manager-webhook-selectel \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag=$(IMAGE_TAG) \
        deploy/cert-manager-webhook-selectel > "$(OUT)/rendered-manifest.yaml"
