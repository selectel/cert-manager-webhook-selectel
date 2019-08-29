IMAGE_NAME := "selectel/cert-manager-webhook-selectel"
IMAGE_TAG := "latest"

OUT := $(shell pwd)/.out

$(shell mkdir -p "$(OUT)")

verify:
	sh ./scripts/fetch-test-binaries.sh
	go test -v .

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

golangci-lint:
	@sh -c "'$(CURDIR)/scripts/golangci_lint_check.sh'"

unit-tests:
	@sh -c "'$(CURDIR)/scripts/unit_tests.sh'"

.PHONY: rendered-manifest.yaml golangci-lint unit-tests
rendered-manifest.yaml:
	helm template \
	    --name cert-manager-webhook-selectel \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag=$(IMAGE_TAG) \
        deploy/cert-manager-webhook-selectel > "$(OUT)/rendered-manifest.yaml"
