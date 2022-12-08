TEST?=$$(go list ./... | grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=observe
VERSION?=$(shell git describe --tags --always)
TESTARGS?=
SWEEP?=all
SWEEP_DIR?=./observe
OBSERVE_ROOT?=$(HOME)/observe
OBSERVE_DOCS_ROOT?=../observe-docs

default: build

docker-test-gen-gql-client:
	docker run -t --network=host -v `pwd`:/go/src/github.com/observeinc/terraform-provider-observe \
	--rm golang:latest \
		/bin/bash -c "src/github.com/observeinc/terraform-provider-observe/scripts/check-client-generation.sh"

docker-integration:
	docker run -v `pwd`:/go/src/github.com/observeinc/terraform-provider-observe \
	-e OBSERVE_CUSTOMER -e OBSERVE_API_TOKEN -e OBSERVE_DOMAIN -e OBSERVE_USER_EMAIL -e OBSERVE_USER_PASSWORD -e OBSERVE_WORKSPACE \
	--rm golang:latest \
	    /bin/bash -c "cd src/github.com/observeinc/terraform-provider-observe && make testacc"

docker-package:
	docker run --network=host -v `pwd`:/go/src/github.com/observeinc/terraform-provider-observe \
	--rm golang:latest \
	    /bin/bash -c " \
		cd src/github.com/observeinc/terraform-provider-observe && \
		apt-get update && \
		apt-get install -y zip && \
		rm -rf bin && \
		GOOS=darwin GOARCH=amd64 make package && \
		GOOS=darwin GOARCH=arm64 make package && \
		GOOS=linux GOARCH=amd64 make package && \
		GOOS=linux GOARCH=arm64 make package"

docker-sign:
	docker run --network=host -v `pwd`:/root/build \
    --rm ubuntu:20.04 /bin/bash -c "\
		apt-get update && apt-get install --yes gpg && \
		gpg --import /root/build/private.pgp && \
		cd /root/build/bin/$(VERSION) && \
		sha256sum *.zip > terraform-provider-observe_$(VERSION)_SHA256SUMS && \
		gpg --yes \
			--output terraform-provider-observe_$(VERSION)_SHA256SUMS.sig \
			--detach-sign terraform-provider-observe_$(VERSION)_SHA256SUMS"

gen-gql-client:
	cd client/meta && GOOS= GOARCH= go run -mod=mod github.com/Khan/genqlient

copy-gql-schema:
	[ -d "$(OBSERVE_ROOT)" ]
	rm client/internal/meta/schema/*.graphql
	cp -pRd "$(OBSERVE_ROOT)/code/go/src/observe/meta/metagql/schema/"*.graphql client/internal/meta/schema/

generate: gen-gql-client

package: build
	cd bin/$(VERSION); zip -mgq terraform-provider-observe_$(VERSION)_$(GOOS)_$(GOARCH).zip terraform-provider-observe_$(VERSION)

build: generate fmtcheck
	go build -o bin/$(VERSION)/terraform-provider-observe_$(VERSION) -ldflags="-X github.com/observeinc/terraform-provider-observe/version.ProviderVersion=$(VERSION)"

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS)

test: fmtcheck
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC_TERRAFORM_CACHE_DIR=/tmp TF_ACC_TERRAFORM_VERSION=1.1.7 TF_LOG=DEBUG TF_ACC=1 go test $(TEST) -v -json -parallel=5 $(TESTARGS) -timeout 120m

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

docs:
	tfplugindocs generate

docs-sync: docs
	rsync -av \
		--include="index.md" \
		--include="*/" \
		--include="*/dataset.md" \
		--include="*/datastream.md" \
		--include="resources/datastream_token.md" \
		--include="resources/link.md" \
		--include="*/monitor.md" \
		--include="data-sources/workspace.md" \
		--exclude="*" \
		docs/ \
		$(OBSERVE_DOCS)/docs/content/terraform/generated

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

.PHONY: build generate test sweep testacc vet fmt fmtcheck errcheck test-compile website website-test docs
