# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=1.0.0)
# - use environment variables to overwrite this value (e.g export VERSION=1.0.0)
VERSION ?= 1.0.0

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# ghcr.io/seo-yul/ncloud-server-controller-bundle:$VERSION and ghcr.io/seo-yul/ncloud-server-controller-catalog:$VERSION.
IMAGE_TAG_BASE ?= ghcr.io/seo-yul/ncloud-server-controller

# Helm Chart Configuration
HELM_CHART_NAME ?= ncloud-server-controller
HELM_CHART_VERSION ?= 1.0.0
HELM_REGISTRY ?= ghcr.io/seo-yul
HELM_CHART_PACKAGE ?= $(HELM_CHART_NAME)-$(HELM_CHART_VERSION).tgz

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Set the Operator SDK version to use. By default, what is installed on the system is used.
# This is useful for CI or a project to utilize a specific version of the operator-sdk toolkit.
OPERATOR_SDK_VERSION ?= v1.41.1
# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/seo-yul/ncloud-server-controller:v$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Automatically detects Docker or Podman based on environment:
# - GitHub Actions: Always use Docker
# - Local development: Prefer Podman if available, fallback to Docker
# - Manual override: Set CONTAINER_TOOL=docker or CONTAINER_TOOL=podman
ifeq ($(GITHUB_ACTIONS),true)
CONTAINER_TOOL ?= docker
else
CONTAINER_TOOL ?= $(shell command -v podman >/dev/null 2>&1 && echo podman || echo docker)
endif

# Container tool detection info
CONTAINER_TOOL_INFO := $(shell echo "Using $(CONTAINER_TOOL) as container tool")

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: container-info
container-info: ## Show container tool information
	@echo "🐳 Container Tool Information"
	@echo "============================"
	@echo "Detected tool: $(CONTAINER_TOOL)"
	@echo "GitHub Actions: $(GITHUB_ACTIONS)"
	@echo ""
	@echo "Available build commands:"
	@echo "  make docker-build-single  - Build single architecture Docker image"
	@echo "  make docker-build-multi   - Build multi-architecture Docker image"
	@echo "  make podman-build-single  - Build single architecture Podman image"
	@echo "  make podman-build-multi   - Build multi-architecture Podman image"
	@echo ""
	@echo "Single architecture examples:"
	@echo "  make docker-build-single PLATFORM=linux/amd64"
	@echo "  make docker-build-single PLATFORM=linux/arm64"
	@echo "  make podman-build-single PLATFORM=linux/amd64"
	@echo "  make podman-build-single PLATFORM=linux/arm64"
	@echo ""
	@echo "Multi-architecture examples:"
	@echo "  make docker-build-multi PLATFORMS=linux/amd64,linux/arm64"
	@echo "  make podman-build-multi PLATFORMS=linux/amd64,linux/arm64"
	@echo ""
	@echo "Environment detection:"
	@echo "  - GitHub Actions: Always uses Docker"
	@echo "  - Local development: Prefers Podman if available, falls back to Docker"
	@echo "  - Manual override: Set CONTAINER_TOOL=docker or CONTAINER_TOOL=podman"

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest-local ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

.PHONY: test-coverage
test-coverage: test ## Generate test coverage report
	go tool cover -html=cover.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage percentage:"
	go tool cover -func=cover.out | tail -1

# TODO(user): To use a different vendor for e2e tests, modify the setup under 'tests/e2e'.
# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# CertManager is installed by default; skip with:
# - CERT_MANAGER_INSTALL_SKIP=true
KIND_CLUSTER ?= ncloud-server-controller-test-e2e

.PHONY: setup-test-e2e
setup-test-e2e: ## Set up a Kind cluster for e2e tests if it does not exist
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "Kind is not installed. Please install Kind manually."; \
		exit 1; \
	}
	@case "$$($(KIND) get clusters)" in \
		*"$(KIND_CLUSTER)"*) \
			echo "Kind cluster '$(KIND_CLUSTER)' already exists. Skipping creation." ;; \
		*) \
			echo "Creating Kind cluster '$(KIND_CLUSTER)'..."; \
			$(KIND) create cluster --name $(KIND_CLUSTER) ;; \
	esac

.PHONY: test-e2e
test-e2e: setup-test-e2e manifests generate fmt vet ## Run the e2e tests. Expected an isolated environment using Kind.
	@echo "=== Building controller image for E2E testing ==="
	@echo "Using container tool: $(CONTAINER_TOOL)"
	$(CONTAINER_TOOL) build -t ncloud-server-controller:e2e-test .
	@echo "=== Loading image into Kind cluster ==="
	$(KIND) load docker-image ncloud-server-controller:e2e-test --name $(KIND_CLUSTER)
	@echo "=== Running E2E tests ==="
	KIND_CLUSTER=$(KIND_CLUSTER) go test ./test/e2e/ -v -ginkgo.v -timeout=30m
	$(MAKE) cleanup-test-e2e

.PHONY: test-e2e-debug
test-e2e-debug: setup-test-e2e manifests generate fmt vet ## Run the e2e tests with debug information.
	@echo "Debugging E2E test environment..."
	@echo "Kind cluster: $(KIND_CLUSTER)"
	@echo "Project image: ncloud-server-controller:e2e-test"
	@echo "Namespace: ncloud-system"
	@echo "Checking Kind cluster status..."
	$(KIND) get clusters
	@echo "Checking loaded images..."
	$(KIND) get nodes --name $(KIND_CLUSTER) | xargs -I {} docker exec {} crictl images
	KIND_CLUSTER=$(KIND_CLUSTER) go test ./test/e2e/ -v -ginkgo.v
	$(MAKE) cleanup-test-e2e

.PHONY: cleanup-test-e2e
cleanup-test-e2e: ## Tear down the Kind cluster used for e2e tests
	@$(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

.PHONY: security-scan
security-scan: ## Run security scan using gosec
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		echo "✅ gosec found, running scan..."; \
		gosec ./...; \
	else \
		echo "⚠️ gosec not found. Installing gosec..."; \
		GOPROXY=https://proxy.golang.org,direct go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		if [ $$? -eq 0 ]; then \
			echo "✅ gosec installed successfully"; \
			$(GOBIN)/gosec ./...; \
		else \
			echo "❌ Failed to install gosec"; \
			echo "💡 Trying alternative installation method..."; \
			go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest || echo "❌ gosec installation failed"; \
		fi; \
	fi

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

.PHONY: run-local
run-local: ## Run controller locally with development settings
	@echo "Running controller locally..."
	@echo "Make sure you have kubectl configured to point to your cluster"
	go run ./cmd/main.go --leader-elect=false

.PHONY: dev-setup
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	$(MAKE) manifests
	$(MAKE) generate
	$(MAKE) setup-hooks
	@echo "Development environment ready!"

.PHONY: setup-hooks
setup-hooks: ## Set up Git hooks for pre-commit checks
	@echo "Setting up Git hooks..."
	@if [ -f "scripts/setup-hooks.sh" ]; then \
		./scripts/setup-hooks.sh; \
	else \
		echo "❌ setup-hooks.sh not found"; \
		exit 1; \
	fi

.PHONY: local-check
local-check: ## Run all local checks (fmt, vet, lint, test)
	@echo "🔍 Running local checks..."
	@echo "📝 Formatting code..."
	@$(MAKE) fmt
	@echo "🔍 Running go vet..."
	@$(MAKE) vet
	@echo "🔍 Running linter..."
	@$(MAKE) lint
	@echo "🧪 Running tests..."
	@$(MAKE) test
	@echo "✅ All local checks passed!"

##@ Container Build Commands

# PLATFORM defines the target platform for single architecture builds
# Examples: linux/amd64, linux/arm64, linux/s390x, linux/ppc64le
PLATFORM ?= linux/amd64

# PLATFORMS defines the target platforms for multi-architecture builds
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: docker-build-single
docker-build-single: ## Build single architecture Docker image (use PLATFORM=linux/arm64 to specify architecture)
	@echo "🐳 Building single architecture Docker image..."
	@echo "Platform: $(PLATFORM)"
	@echo "Image: $(IMG)"
	docker build --platform $(PLATFORM) -t $(IMG) .
	@echo "✅ Single architecture build completed: $(IMG)"

.PHONY: docker-build-multi
docker-build-multi: ## Build multi-architecture Docker image using Buildx
	@echo "🐳 Building multi-architecture Docker image..."
	@echo "Platforms: $(PLATFORMS)"
	@echo "Image: $(IMG)"
	# Create and use buildx builder
	docker buildx create --name ncloud-server-controller-builder --use || true
	# Build and push multi-arch image
	docker buildx build --push --platform=$(PLATFORMS) --tag $(IMG) .
	# Cleanup
	docker buildx rm ncloud-server-controller-builder || true
	@echo "✅ Multi-architecture build completed: $(IMG)"

.PHONY: podman-build-single
podman-build-single: ## Build single architecture Podman image (use PLATFORM=linux/arm64 to specify architecture)
	@echo "🐳 Building single architecture Podman image..."
	@echo "Platform: $(PLATFORM)"
	@echo "Image: $(IMG)"
	podman build --platform $(PLATFORM) -t $(IMG) .
	@echo "✅ Single architecture build completed: $(IMG)"

.PHONY: podman-build-multi
podman-build-multi: ## Build multi-architecture Podman image
	@echo "🐳 Building multi-architecture Podman image..."
	@echo "Platforms: $(PLATFORMS)"
	@echo "Image: $(IMG)"
	# Remove existing manifest and images if they exist
	@echo "Cleaning up existing manifest and images (if any)..."
	-podman manifest rm $(IMG) 2>/dev/null || true
	$(eval ARCHS=$(shell echo $(PLATFORMS) | tr ',' ' '))
	@for arch in $(ARCHS); do \
		echo "Removing existing image for platform: $$arch"; \
		podman rmi $(IMG)-$${arch##*/} 2>/dev/null || true; \
	done
	# Create manifest list
	podman manifest create $(IMG)
	# Build and push for each architecture
	$(eval ARCHS=$(shell echo $(PLATFORMS) | tr ',' ' '))
	@for arch in $(ARCHS); do \
		echo "Building for platform: $$arch"; \
		podman build --platform $$arch --tag $(IMG)-$${arch##*/} . && \
		echo "Pushing $$arch image" && \
		podman push $(IMG)-$${arch##*/} && \
		echo "Adding $$arch to manifest" && \
		podman manifest add $(IMG) docker://$(IMG)-$${arch##*/}; \
	done
	# Push manifest list
	@echo "Pushing multi-architecture manifest"
	podman manifest push $(IMG) docker://$(IMG)
	@echo "✅ Multi-architecture build completed: $(IMG)"

.PHONY: docker-push
docker-push: ## Push Docker image
	docker push $(IMG)

.PHONY: podman-push
podman-push: ## Push Podman image
	podman push $(IMG)


.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml

.PHONY: helm-template
helm-template: ## Template Helm chart
	helm template test-release helm/ncloud-server-controller --dry-run

##@ Release Management

.PHONY: release-info
release-info: ## Show release information
	@echo "📋 Release Information:"
	@echo "  Version: v$(VERSION)"
	@echo "  Image: $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "  Helm Chart: $(HELM_CHART_PACKAGE)"
	@echo "  Registry: $(HELM_REGISTRY)"

.PHONY: release-check
release-check: ## Check if release prerequisites are met
	@echo "🔍 Checking release prerequisites..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ GITHUB_TOKEN environment variable is required"; \
		exit 1; \
	fi
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "❌ GitHub CLI (gh) is required"; \
		exit 1; \
	fi
	@if ! gh auth status >/dev/null 2>&1; then \
		echo "❌ GitHub CLI authentication required"; \
		exit 1; \
	fi
	@echo "✅ All prerequisites met"

.PHONY: release-tag
release-tag: release-check ## Create Git tag for release
	@echo "🏷️ Creating Git tag v$(VERSION)..."
	@if git tag -l | grep -q "^v$(VERSION)$$"; then \
		echo "⚠️ Tag v$(VERSION) already exists"; \
		echo "💡 To recreate, run: git tag -d v$(VERSION) && git push origin :refs/tags/v$(VERSION)"; \
		exit 1; \
	fi
	@git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@git push origin "v$(VERSION)"
	@echo "✅ Git tag v$(VERSION) created and pushed"

.PHONY: release-image-tag
release-image-tag: ## Tag existing image with release version
	@echo "🐳 Tagging existing image with release version..."
	@echo "Using container tool: $(CONTAINER_TOOL)"
	@echo "Source: $(IMAGE_TAG_BASE):develop"
	@echo "Target: $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "Target: $(IMAGE_TAG_BASE):latest"
	@echo "💡 Manual tagging commands:"
	@echo "  $(CONTAINER_TOOL) pull $(IMAGE_TAG_BASE):develop"
	@echo "  $(CONTAINER_TOOL) tag $(IMAGE_TAG_BASE):develop $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "  $(CONTAINER_TOOL) tag $(IMAGE_TAG_BASE):develop $(IMAGE_TAG_BASE):latest"
	@echo "  $(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "  $(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):latest"

.PHONY: release-image-tag-auto
release-image-tag-auto: ## Automatically tag and push release images
	@echo "🐳 Automatically tagging and pushing release images..."
	@echo "Using container tool: $(CONTAINER_TOOL)"
	@echo "Source: $(IMAGE_TAG_BASE):develop"
	@echo "Target: $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "Target: $(IMAGE_TAG_BASE):latest"
	@echo "📥 Pulling source image..."
	$(CONTAINER_TOOL) pull $(IMAGE_TAG_BASE):develop
	@echo "🏷️ Tagging images..."
	$(CONTAINER_TOOL) tag $(IMAGE_TAG_BASE):develop $(IMAGE_TAG_BASE):v$(VERSION)
	$(CONTAINER_TOOL) tag $(IMAGE_TAG_BASE):develop $(IMAGE_TAG_BASE):latest
	@echo "📤 Pushing tagged images..."
	$(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):v$(VERSION)
	$(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):latest
	@echo "✅ Release images tagged and pushed successfully"

.PHONY: release-github
release-github: release-check ## Create GitHub release
	@echo "🚀 Creating GitHub release for version $(VERSION)..."
	@if gh release view "v$(VERSION)" >/dev/null 2>&1; then \
		echo "⚠️ Release v$(VERSION) already exists"; \
		echo "💡 To recreate, run: gh release delete v$(VERSION)"; \
		exit 1; \
	fi
	@echo "Creating release for tag v$(VERSION)..."
	@gh release create "v$(VERSION)" \
		--title "Release v$(VERSION)" \
		--notes "Release v$(VERSION) of NCloud Server Controller

	## 🚀 What's New
	- NCloud Server Controller Operator v$(VERSION)
	- Multi-architecture support (linux/amd64, linux/arm64)
	- Helm chart included
	
	## 📦 Installation
	\`\`\`bash
	# Using Helm
	helm install ncloud-server-controller oci://$(HELM_REGISTRY)/$(HELM_CHART_NAME) --version v$(VERSION)
	
	# Using kubectl
	kubectl apply -k https://github.com/seo-yul/ncloud-server-controller/config/default?ref=v$(VERSION)
	\`\`\`
	
	## 🐳 Docker Images
	- \`$(IMAGE_TAG_BASE):v$(VERSION)\`
- \`$(IMAGE_TAG_BASE):latest\`" \
		--latest
	@echo "✅ GitHub release created successfully"

.PHONY: release-helm
release-helm: helm-package ## Package Helm chart for release
	@echo "📦 Packaging Helm chart for release..."
	@echo "Chart: $(HELM_CHART_PACKAGE)"
	@if [ -f "$(HELM_CHART_PACKAGE)" ]; then \
		echo "✅ Helm chart packaged: $(HELM_CHART_PACKAGE)"; \
	else \
		echo "❌ Helm chart package not found"; \
		exit 1; \
	fi

.PHONY: release-upload-assets
release-upload-assets: release-helm ## Upload Helm chart to GitHub release
	@echo "📤 Uploading Helm chart to GitHub release..."
	@if ! gh release view "v$(VERSION)" >/dev/null 2>&1; then \
		echo "❌ Release v$(VERSION) not found. Run 'make release-github' first"; \
		exit 1; \
	fi
	@if [ ! -f "$(HELM_CHART_PACKAGE)" ]; then \
		echo "❌ Helm chart package not found: $(HELM_CHART_PACKAGE)"; \
		exit 1; \
	fi
	@gh release upload "v$(VERSION)" "$(HELM_CHART_PACKAGE)"
	@echo "✅ Helm chart uploaded to release v$(VERSION)"

.PHONY: release
release: release-info release-check release-tag release-image-tag-auto release-github release-upload-assets ## Complete release process
	@echo "🎉 Release v$(VERSION) completed successfully!"
	@echo ""
	@echo "📋 Release Summary:"
	@echo "  Version: v$(VERSION)"
	@echo "  Git Tag: v$(VERSION)"
	@echo "  Image: $(IMAGE_TAG_BASE):v$(VERSION)"
	@echo "  Helm Chart: $(HELM_CHART_PACKAGE)"
	@echo "  GitHub Release: https://github.com/seo-yul/ncloud-server-controller/releases/tag/v$(VERSION)"
	@echo ""
	@echo "💡 Next Steps:"
	@echo "  1. Tag the Docker image: make release-image-tag"
	@echo "  2. Push Docker images: $(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):v$(VERSION) && $(CONTAINER_TOOL) push $(IMAGE_TAG_BASE):latest"
	@echo "  3. Test the release: helm install test-release oci://$(HELM_REGISTRY)/$(HELM_CHART_NAME) --version v$(VERSION)"

.PHONY: release-quick
release-quick: release-check release-github release-upload-assets ## Quick release (skip Git tag creation)
	@echo "⚡ Quick release v$(VERSION) completed!"
	@echo "💡 Note: Git tag was not created. Run 'make release-tag' if needed."

.PHONY: github-release
github-release: release-github ## Alias for release-github (backward compatibility)

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@echo "Cleaning build artifacts..."
	rm -rf dist/
	rm -f cover.out coverage.html
	rm -f Dockerfile.cross
	@echo "Cleaning bin directory (excluding k8s binaries)..."
	@find bin/ -type f ! -path "*/k8s/*" -delete 2>/dev/null || true
	@echo "Cleanup complete"

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.18.0
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.1.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest-local
setup-envtest-local: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: operator-sdk
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (, $(shell which operator-sdk 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which operator-sdk)
endif
endif

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_TOOL) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = $(LOCALBIN)/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.55.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool $(CONTAINER_TOOL) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

##@ CI/CD Pipeline

.PHONY: ci-build
ci-build: manifests generate fmt vet lint test ## Complete CI build pipeline
	@echo "CI build pipeline completed successfully"

.PHONY: ci-test
ci-test: test test-coverage ## Complete CI test pipeline
	@echo "CI test pipeline completed successfully"

.PHONY: ci-release
ci-release: ci-build helm-all release-github release-upload-assets ## Complete CI release pipeline
	@echo "CI release pipeline completed successfully"

.PHONY: pre-commit
pre-commit: fmt vet lint test ## Run pre-commit checks
	@echo "Pre-commit checks completed successfully"

.PHONY: all-checks
all-checks: pre-commit security-scan ## Run all quality checks
	@echo "All quality checks completed successfully"

##@ Local Testing Pipeline

.PHONY: setup-envtest
setup-envtest: ## Setup envtest binaries for CI/CD
	@echo "🔧 Setting up envtest binaries..."
	@mkdir -p $${HOME:-/home/runner}/kubebuilder/bin
	@if command -v setup-envtest >/dev/null 2>&1; then \
		echo "✅ setup-envtest found, installing binaries..."; \
		setup-envtest use 1.29.0 --bin-dir $${HOME:-/home/runner}/kubebuilder/bin; \
		echo "✅ envtest binaries installed in $${HOME:-/home/runner}/kubebuilder/bin"; \
	else \
		echo "⚠️ setup-envtest not found, installing..."; \
		go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest; \
		echo "✅ setup-envtest installed, now installing binaries..."; \
		setup-envtest use 1.29.0 --bin-dir $${HOME:-/home/runner}/kubebuilder/bin; \
		echo "✅ envtest binaries installed in $${HOME:-/home/runner}/kubebuilder/bin"; \
	fi
	@echo "🔍 Verifying installation..."
	@ls -la $${HOME:-/home/runner}/kubebuilder/bin/ || echo "Directory not found"
	@find $${HOME:-/home/runner}/kubebuilder/bin -name "etcd" -type f 2>/dev/null || echo "etcd not found"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@echo "KUBEBUILDER_ASSETS: $(KUBEBUILDER_ASSETS)"
	go test -short ./internal/controller/... -v

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	go test ./test/integration/... -v

.PHONY: test-all
test-all: ## Run all tests (unit + integration + e2e)
	@echo "🧪 Running all tests..."
	@echo "1️⃣ Unit tests..."
	@$(MAKE) test-unit
	@echo "2️⃣ Integration tests..."
	@$(MAKE) test-integration
	@echo "3️⃣ E2E tests..."
	@$(MAKE) test-e2e
	@echo "✅ All tests completed!"

.PHONY: test-quick
test-quick: ## Quick test (unit tests only, no coverage)
	@echo "⚡ Running quick tests..."
	go test -short ./internal/... ./api/... -v -timeout=30s

.PHONY: test-coverage-full
test-coverage-full: ## Generate comprehensive test coverage report
	@echo "📊 Generating comprehensive test coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## Run benchmark tests
	@echo "🏃 Running benchmark tests..."
	go test -bench=. -benchmem ./internal/...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "🏁 Running tests with race detection..."
	go test -race ./internal/... ./api/...

.PHONY: test-stress
test-stress: ## Run stress tests
	@echo "💪 Running stress tests..."
	go test -count=10 -race ./internal/...

.PHONY: validate-local
validate-local: ## Complete local validation pipeline
	@echo "🔍 Running complete local validation..."
	@echo "1️⃣ Code formatting..."
	@$(MAKE) fmt
	@echo "2️⃣ Static analysis..."
	@$(MAKE) vet
	@echo "3️⃣ Linting..."
	@$(MAKE) lint
	@echo "4️⃣ Security scan..."
	@$(MAKE) security-scan
	@echo "5️⃣ Unit tests..."
	@$(MAKE) test-unit
	@echo "6️⃣ Dependencies check..."
	@$(MAKE) check-deps
	@echo "✅ Local validation completed successfully!"

.PHONY: dev-test
dev-test: ## Development testing pipeline (fast feedback)
	@echo "🚀 Running development test pipeline..."
	@echo "1️⃣ Quick formatting..."
	@$(MAKE) fmt
	@echo "2️⃣ Quick vet..."
	@$(MAKE) vet
	@echo "3️⃣ Quick tests..."
	@$(MAKE) test-quick
	@echo "✅ Development tests completed!"

.PHONY: ci-local
ci-local: ## Run full CI pipeline locally
	@echo "🏗️ Running full CI pipeline locally..."
	@echo "1️⃣ Build pipeline..."
	@$(MAKE) ci-build
	@echo "2️⃣ Test pipeline..."
	@$(MAKE) ci-test
	@echo "3️⃣ Quality checks..."
	@$(MAKE) all-checks
	@echo "✅ Full CI pipeline completed locally!"

##@ Version Management

.PHONY: version-info
version-info: ## Show current version information
	@echo "📦 Version Information"
	@echo "===================="
	@echo "Current version: v$(VERSION)"
	@echo "Image tag base: $(IMAGE_TAG_BASE)"
	@echo "Current image: $(IMG)"
	@echo ""
	@echo "Version management:"
	@echo "  make version-patch  - Create patch release"
	@echo "  make version-minor  - Create minor release"
	@echo "  make version-major  - Create major release"
	@echo "  make version-info   - Show this information"

.PHONY: version-patch
version-patch: ## Create patch release (bug fixes)
	@echo "🔧 Creating patch release..."
	@./scripts/version.sh patch

.PHONY: version-minor
version-minor: ## Create minor release (new features)
	@echo "✨ Creating minor release..."
	@./scripts/version.sh minor

.PHONY: version-major
version-major: ## Create major release (breaking changes)
	@echo "🚀 Creating major release..."
	@./scripts/version.sh major

.PHONY: version-check
version-check: ## Check version consistency across files
	@echo "🔍 Checking version consistency..."
	@echo "Makefile VERSION: v$(VERSION)"
	@echo "Helm Chart version: $$(grep '^version:' helm/ncloud-server-controller/Chart.yaml | cut -d' ' -f2)"
	@echo "Helm Chart appVersion: $$(grep '^appVersion:' helm/ncloud-server-controller/Chart.yaml | cut -d' ' -f2)"
	@echo "Latest git tag: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'No tags found')"


##@ Container Registry Management

.PHONY: cleanup-manifests
cleanup-manifests: ## Clean up existing manifests from GHCR using API
	@echo "🧹 Cleaning up existing manifests from GHCR..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ GITHUB_TOKEN environment variable is required"; \
		exit 1; \
	fi
	@echo "📦 Package name: $(shell echo $(IMG) | cut -d: -f1 | sed 's|.*/||')"
	@echo "🏷️  Tags to clean: $(TAGS)"
	@echo "🔧 Processing tags for cleanup..."
	@for tag in $(TAGS); do \
		if [ -n "$$tag" ]; then \
			tag_name="$${tag##*:}"; \
			echo "🗑️  Cleaning up tag: $$tag_name"; \
			gh api DELETE "/user/packages/container/$(shell echo $(IMG) | cut -d: -f1 | sed 's|.*/||')" \
				-H "Accept: application/vnd.github+json" \
				-H "X-GitHub-Api-Version: 2022-11-28" \
				--silent || echo "⚠️  Failed to delete package or package not found"; \
		fi; \
	done
	@echo "⏳ Waiting for GHCR cleanup to complete..."
	@sleep 30
	@echo "✅ GHCR cleanup completed"



##@ Helm Chart Management

.PHONY: helm-generate
helm-generate: ## Generate Helm chart from current kustomize resources (CRD only)
	@echo "🔧 Generating Helm chart from kustomize resources..."
	mkdir -p helm/ncloud-server-controller/templates
	# Generate CRDs (make install과 동일)
	# Copy NCloudServer CRD directly from source
	cp config/crd/bases/server.ncloud.devops.ai.kr_ncloudservers.yaml helm/ncloud-server-controller/templates/ncloudserver-crd.yaml
	# Update Chart.yaml with Helm chart version (separate from image version)
	@echo "📝 Updating Chart.yaml with Helm chart version $(HELM_CHART_VERSION)..."
	@echo "apiVersion: v2" > helm/ncloud-server-controller/Chart.yaml
	@echo "name: $(HELM_CHART_NAME)" >> helm/ncloud-server-controller/Chart.yaml
	@echo "description: A Helm chart for NCloud Server Controller Operator" >> helm/ncloud-server-controller/Chart.yaml
	@echo "type: application" >> helm/ncloud-server-controller/Chart.yaml
	@echo "version: $(HELM_CHART_VERSION)" >> helm/ncloud-server-controller/Chart.yaml
	@echo "appVersion: \"$(VERSION)\"" >> helm/ncloud-server-controller/Chart.yaml
	@echo "icon: https://raw.githubusercontent.com/Seo-yul/ncloud-server-controller/main/docs/icon.png" >> helm/ncloud-server-controller/Chart.yaml
	@echo "✅ Helm chart generated successfully with Helm chart version $(HELM_CHART_VERSION) and app version $(VERSION)"
	@echo "⚠️  Note: Individual templates are manually maintained and not overwritten"

.PHONY: helm-package
helm-package: helm-generate ## Package the operator as a Helm chart
	@echo "📦 Packaging Helm chart..."
	helm package helm/ncloud-server-controller/
	@echo "✅ Helm chart packaged as $(HELM_CHART_PACKAGE)"

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	@echo "🔍 Linting Helm chart..."
	helm lint helm/ncloud-server-controller/
	@echo "✅ Helm chart linting completed"

.PHONY: helm-push
helm-push: helm-package ## Push Helm chart to OCI registry
	@echo "🚀 Pushing Helm chart to OCI registry..."
	@if helm push $(HELM_CHART_PACKAGE) oci://$(HELM_REGISTRY); then \
		echo "✅ Helm chart pushed to $(HELM_REGISTRY)/$(HELM_CHART_NAME):$(HELM_CHART_VERSION)"; \
	else \
		echo "⚠️  Failed to push to OCI registry. Chart packaged locally: $(HELM_CHART_PACKAGE)"; \
		echo "💡 You can manually push with: helm push $(HELM_CHART_PACKAGE) oci://$(HELM_REGISTRY)"; \
		exit 1; \
	fi

.PHONY: helm-dry-run
helm-dry-run: ## Dry run Helm installation
	@echo "🧪 Dry run Helm installation..."
	helm install ncloud-server-controller helm/ncloud-server-controller/ --dry-run --debug \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2) \
		--namespace ncloud-system --create-namespace

.PHONY: helm-install
helm-install: ## Install the operator using Helm
	@echo "📋 Installing NCloud Server Controller using Helm..."
	helm install ncloud-server-controller helm/ncloud-server-controller/ \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2) \
		--namespace ncloud-system --create-namespace
	@echo "✅ NCloud Server Controller installed successfully"

.PHONY: helm-install-from-registry
helm-install-from-registry: ## Install the operator using Helm chart from OCI registry
	@echo "📋 Installing NCloud Server Controller from OCI registry..."
	helm install ncloud-server-controller oci://$(HELM_REGISTRY)/$(HELM_CHART_NAME) \
		--version $(HELM_CHART_VERSION) \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2) \
		--namespace ncloud-system --create-namespace
	@echo "✅ NCloud Server Controller installed from registry"

.PHONY: helm-install-with-values
helm-install-with-values: ## Install the operator using Helm chart from OCI registry with local values
	@echo "📋 Installing NCloud Server Controller with custom values..."
	helm install ncloud-server-controller oci://$(HELM_REGISTRY)/$(HELM_CHART_NAME) \
		--version $(HELM_CHART_VERSION) \
		-f helm/ncloud-server-controller/values.yaml \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2) \
		--namespace ncloud-system --create-namespace
	@echo "✅ NCloud Server Controller installed with custom values"

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade the operator using Helm
	@echo "🔄 Upgrading NCloud Server Controller using Helm..."
	helm upgrade ncloud-server-controller helm/ncloud-server-controller/ \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2) \
		--namespace ncloud-system
	@echo "✅ NCloud Server Controller upgraded successfully"

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall the operator using Helm
	@echo "🗑️ Uninstalling NCloud Server Controller using Helm..."
	helm uninstall ncloud-server-controller --namespace ncloud-system
	@echo "✅ NCloud Server Controller uninstalled successfully"

.PHONY: helm-clean
helm-clean: ## Clean up generated Helm chart package
	@echo "🧹 Cleaning up Helm chart package..."
	rm -f $(HELM_CHART_PACKAGE)
	@echo "✅ Helm chart package cleaned up"

.PHONY: helm-all
helm-all: helm-generate helm-package helm-lint ## Run all Helm operations (generate, package, lint)
	@echo "🎉 All Helm operations completed successfully!"
	@echo "📦 Chart packaged as: $(HELM_CHART_PACKAGE)"
	@echo "💡 To push to registry: make helm-push"

.PHONY: helm-all-with-push
helm-all-with-push: helm-generate helm-package helm-lint helm-push ## Run all Helm operations including push
	@echo "🎉 All Helm operations including push completed successfully!"
