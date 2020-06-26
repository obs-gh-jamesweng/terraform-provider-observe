TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=observe
VERSION=$(shell git describe --tags --always)

default: build

docker-integration:
	docker run -v `pwd`:/go/src/github.com/observeinc/terraform-provider-observe \
	-e OBSERVE_CUSTOMER -e OBSERVE_TOKEN -e OBSERVE_DOMAIN -e OBSERVE_USER_EMAIL -e OBSERVE_USER_PASSWORD \
	--rm golang:latest \
	    /bin/bash -c "cd src/github.com/observeinc/terraform-provider-observe && make testacc"

docker-package:
	docker run -v `pwd`:/go/src/github.com/observeinc/terraform-provider-observe \
	--rm golang:latest \
	    /bin/bash -c " \
		cd src/github.com/observeinc/terraform-provider-observe && \
		apt-get update && \
		apt-get install -y zip && \
		rm -rf bin && \
		GOOS=darwin GOARCH=amd64 make package && \
		GOOS=linux GOARCH=amd64 make package"

package: fmtcheck
	go build -o bin/$(GOOS)_$(GOARCH)/terraform-provider-observe_$(VERSION) -ldflags="-X github.com/observeinc/terraform-provider-observe/version.ProviderVersion=$(VERSION)"
	cd bin/$(GOOS)_$(GOARCH); zip -mgq terraform-provider-observe_$(VERSION).zip terraform-provider-observe_$(VERSION)

build: fmtcheck
	go install -ldflags="-X github.com/observeinc/terraform-provider-observe/version.ProviderVersion=$(VERSION)"

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(TEST) -v -sweep=$(SWEEP) $(SWEEPARGS)

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC_TERRAFORM_VERSION=0.12.26 TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test sweep testacc vet fmt fmtcheck errcheck test-compile website website-test
