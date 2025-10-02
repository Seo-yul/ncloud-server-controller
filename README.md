# 네이버 클라우드 플랫폼 서버 관리 Kubernetes Operator

네이버 클라우드 플랫폼의 서버 인스턴스를 Kubernetes에서 관리할 수 있는 Custom Operator입니다.

## 📋 개요

이 프로젝트는 Go Operator SDK를 사용하여 네이버 클라우드 플랫폼의 서버 인스턴스를 Kubernetes Custom Resource로 관리할 수 있도록 구현합니다.

## 🛠️ 개발 환경 설정

### 1. 필수 도구 설치

```bash
# Go 버전 확인
go version

# 크로스 컴파일을 위한 환경변수 설정 (Linux AMD64/ARM64 배포용)
export GOOS=linux
export GOARCH=amd64  # 또는 arm64
export CGO_ENABLED=0  # 정적 링킹을 위한 설정
```

```bash
# kubectl 설치
brew install kubectl  # macOS
# 또는 https://kubernetes.io/docs/tasks/tools/ 에서 다운로드
```

```bash
# Operator SDK CLI 설치
brew install operator-sdk

# 설치 확인
operator-sdk version
```

```bash
# 네이버 클라우드 CLI 사용.
./ncloud_cli_linux/ncloud

# 설치 확인
./ncloud_cli_linux/ncloud --version
```

### 2. 프로젝트 초기화

```bash
# 현재 디렉토리에서 새 Operator 프로젝트 생성
# --domain: CRD의 API Group 정의 (server.ncloud.devops.ai.kr/v1)
# --repo: Go 모듈 경로 정의 (go.mod의 module name)
operator-sdk init --domain ncloud.devops.ai.kr --repo github.com/Seo-yul/ncloud-server-controller

# 생성되는 파일 예시:
# - go.mod: module github.com/Seo-yul/ncloud-server-controller
# - api/v1/ncloudserver_types.go: Kind = "NCloudServer"
# - controllers/: CRD 그룹 = "server.ncloud.devops.ai.kr/v1"

# Custom Resource 정의
operator-sdk create api --group server --version v1 --kind NCloudServer --resource --controller

# Kustomize 매니페스트 생성 (CRD 배포용)
# 대안: make 명령어로도 가능
make manifests

# 또는 원하는 경우 대화형 입력 포함 생성
# operator-sdk generate kustomize manifests
```

### 3. 프로젝트 구조 확인

실제 생성된 프로젝트 구조:
```
ncloud-server-controller/
├── api/
│   └── v1/
│       ├── groupversion_info.go
│       ├── ncloudserver_types.go
│       └── zz_generated.deepcopy.go
├── config/
│   ├── crd/bases/
│   │   └── server.ncloud.devops.ai.kr_ncloudservers.yaml
│   ├── rbac/
│   ├── manager/
│   └── samples/
│       └── server_v1_ncloudserver.yaml
├── internal/
│   └── controller/
│       ├── ncloudserver_controller.go
│       └── ncloudserver_controller_test.go
├── cmd/
├── main.go
├── Dockerfile
├── Makefile
└── go.mod
```

## 🔧 개발 단계별 명령어

### 단계 1: Custom Resource 정의 수정

```bash
# Custom Resource 타입 정의 수정
vim api/v1/ncloudserver_types.go
```

```bash
# 코드 생성
make generate
```

### 단계 2: 네이버 클라우드 CLI SDK 연동

```bash
# 필요한 Go 모듈 추가
go mod tidy
```

### 단계 3: Controller 로직 구현

```bash
# Controller 파일 수정 (실제 생성된 경로)
vim internal/controller/ncloudserver_controller.go
```

### 단계 4: 로컬 테스트

```bash
# 크로스 플랫폼 빌드 (현재 환경: darwin/arm64 → linux/amd64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make build

# 또는 Makefile에서 기본 설정으로
make build

# 로컬에서 Operator 실행 (개발용 - macOS 환경)
make run

# Linux용 바이너리 직접 실행 (별도 환경에서)
./bin/manager run
```

### 단계 5: CRD 설치 및 테스트

```bash
# CRD 설치
make install

# Sample Custom Resource 생성 (실제 생성된 경로)
kubectl apply -f config/samples/server_v1_ncloudserver.yaml

# Custom Resource 확인
kubectl get ncloudservers

# 로그 확인
kubectl logs -f deployment/ncloudserver-controller-manager -n ncloud-server-system
```

### 단계 6: 클러스터 배포

```bash
# Docker 이미지 빌드 (멀티 아키텍처 지원)
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag your-registry/ncloudserver:v0.0.1 \
  --push .

# 또는 단일 아키텍처
make docker-build docker-push IMG=your-registry/ncloudserver:v0.0.1

# 매니페스트 생성 및 배포
make deploy IMG=your-registry/ncloudserver:v0.0.1
```

## 🔐 네이버 클라우드 인증 설정

### 환경변수 방식

```bash
# 네이버 클라우드 API 키 설정
export NCLOUD_ACCESS_KEY="your_access_key"
export NCLOUD_SECRET_KEY="your_secret_key"

# 한국 리전 기본 설정
export NCLOUD_DEFAULT_REGION="KR"
```

### Kubernetes Secret 방식

```bash
# Namespace 생성
kubectl create namespace ncloud-server-system

# 네이버 클라우드 인증 정보를 Secret으로 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key="your_access_key" \
  --from-literal=secret-key="your_secret_key" \
  -n ncloud-server-system
```

## 📊 API 명령어 참고

### 서버 생성
```bash
ncloud vserver createServerInstances \
  --regionCode KR \
  --serverName "operator-managed-server" \
  --serverImageProductCode "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050" \
  --serverProductCode "SVR.VSVR.HICPU.C002.M004.NET.SSD.B050.G002" \
  --loginKeyName "k8s-key"
```

### 서버 목록 조회
```bash
ncloud vserver getServerInstanceList --regionCode KR
```

### 서버 삭제
```bash
ncloud vserver eliminateServerInstances \
  --regionCode KR \
  --serverInstanceNoList "server-instance-no"
```

### 서버 상태 확인
```bash
ncloud vserver getServerInstanceDetail \
  --regionCode KR \
  --serverInstanceNo "server-instance-no"
```

## 🧪 테스트 명령어

```bash
# 전체 테스트 실행 (현재 환경)
make test

# 특정 기능 테스트
make test-local FOCUS="should create NCloudServer"

# Linux 환경을 위한 크로스 컴파일 테스트
GOOS=linux GOARCH=amd64 make test

# 정적 분석
make lint

# 보안 스캔
make security-scan

# 아키텍처별 빌드 테스트
for arch in amd64 arm64 ; do
  echo "Building for linux/$arch"
  GOOS=linux GOARCH=$arch CGO_ENABLED=0 go build -o bin/manager-linux-$arch ./main.go
done
```

## 🚀 배포 및 운영

### Minikube 로컬 테스트

```bash
# Minikube 시작
minikube start

# 이미지 로드
minikube load ncloudserver:latest

# Operator 배포
kubectl apply -f https://raw.githubusercontent.com/yoon/ncloud-server-controller/main/config/samples/compute_v1_ncloudserver.yaml
```

### 프로덕션 배포

```bash
# OLM (Operator Lifecycle Manager) 번들 생성
make bundle

# OperatorHub에 퍼블리시 (선택사항)
make bundle-push
```

## 📁 실제 생성된 Custom Resource 예시

**현재 생성된 Sample 파일** (`config/samples/server_v1_ncloudserver.yaml`):
```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  labels:
    app.kubernetes.io/name: ncloud-server-controller
    app.kubernetes.io/managed-by: kustomize
  name: ncloudserver-sample
spec:
  # TODO(user): Add fields here
```

**향후 구현할 목표 Custom Resource**:
```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: web-server-01
  namespace: default
spec:
  serverName: "kubernetes-web-node"
  region: "KR"
  zone: "KR-2"
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  serverProductCode: "SVR.VSVR.STAND.C002.M008.NET.SSD.B050.G002"
  loginKeyName: "k8s-cluster-key"
  replicas: 3
status:
  phase: "Creating"
  serverInstanceNumbers: []
  message: "Server instances are being created"
```

## 🔍 디버깅 및 트러블슈팅

```bash
# Operator 로그 실시간 확인
kubectl logs -f -l control-plane=controller-manager -n ncloud-server-system

# Custom Resource 이벤트 확인
kubectl describe ncloudserver web-server-01

# 네이버 클라우드 API 호출 테스트 (실제 CLI 사용)
./ncloud_cli_linux/ncloud vserver getRegionList
```

## 📚 참고 자료

- [Kubernetes Operator SDK 문서](https://sdk.operatorframework.io/)
- [네이버 클라우드 플랫폼 CLI 가이드](https://cli.ncloud-docs.com/docs/guide)
- [Go Operator 튜토리얼](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)

## 📝 개발 로드맵

- [x] ✅ **개발 환경 설정 완료** (Go 1.23.1 darwin/arm64)
- [x] ✅ **Operator SDK 프로젝트 초기화 완료**
- [x] ✅ **기본 CRD 정의 완료** (`server.ncloud.devops.ai.kr/v1`)
- [x] ✅ **API 및 Controller 스켈레톤 생성** (`NCloudServer` Kind)
- [x] ✅ **CRD 매니페스트 생성 완료**
- [ ] 네이버 클라우드 CLI 연동
- [ ] 서버 생성/삭제 로직 구현
- [ ] 상태 동기화 구현
- [ ] 웹훅 검증 추가
- [ ] 모니터링 및 메트릭 추가
- [ ] 테스트 커버리지 향상
- [ ] 문서화 완성
