# Build the manager binary
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and.copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Multi-stage build: include NCloud CLI and Java Runtime
FROM ubuntu:22.04 AS ncloud-cli-builder

# NCloud CLI에 필요한 의존성 설치
RUN apt-get update && apt-get install -y \
    openjdk-8-jre-headless \
    wget \
    curl \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# NCloud CLI 다운로드 및 설정
RUN mkdir -p /opt/ncloud-cli && \
    cd /opt/ncloud-cli && \
    wget -O ncloud-cli.zip "https://github.com/NaverCloudPlatform/ncloud-cli/releases/download/v1.1.26/ncloud-cli-linux.zip" && \
    unzip ncloud-cli.zip && \
    mv ncloud-cli-linux ncloud_cli_linux && \
    rm ncloud-cli.zip && \
    chmod +x ncloud_cli_linux/ncloud

# Final stage: 매니저 바이너리 + NCloud CLI
FROM ubuntu:22.04

# NCloud CLI에 필요한 의존성 설치
RUN apt-get update && apt-get install -y \
    openjdk-8-jre-headless \
    curl \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# 매니저 바이너리 복사
COPY --from=builder /workspace/manager /manager

# NCloud CLI 복사
COPY --from=ncloud-cli-builder /opt/ncloud-cli/ /opt/ncloud-cli/

# CLI 실행 권한 설정
RUN chmod +x /opt/ncloud-cli/ncloud_cli_linux/ncloud

# 작업 디렉토리 설정
WORKDIR /opt/ncloud-cli/ncloud_cli_linux

# 비-root 사용자 설정 (Kubernetes 보안 정책 호환)
RUN groupadd -r manager -g 1000 && useradd --no-log-init -r -g manager -u 1000 manager
USER 1000:1000

# GitHub Packages 라벨 추가 (리포지토리 연결 및 메타데이터)
LABEL org.opencontainers.image.source=https://github.com/Seo-yul/ncloud-server-controller
LABEL org.opencontainers.image.description="NCloud Server Controller - Kubernetes Operator for managing NCloud VPC servers"
LABEL org.opencontainers.image.licenses=MIT

ENTRYPOINT ["/manager"]
