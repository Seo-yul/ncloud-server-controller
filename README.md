# 네이버 클라우드 플랫폼 서버 관리 Kubernetes Operator

네이버 클라우드 플랫폼의 서버 인스턴스를 Kubernetes에서 관리할 수 있는 Custom Operator

## 개요

이 프로젝트는 Go Operator SDK를 사용하여 네이버 클라우드 플랫폼의 서버 인스턴스를 Kubernetes Custom Resource로 관리할 수 있도록 구현한다.

## 개발 환경 설정

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

## 개발 단계별 명령어

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
kubectl logs -f deployment/ncloud-server-controller-manager -n ncloud-system
```

### 단계 6: 클러스터 배포

**Docker 사용시:**
```bash
# Docker 멀티 아키텍처 빌드
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag your-registry/ncloudserver:v0.0.1 \
  --push .

# 또는 단일 아키텍처
make docker-build docker-push IMG=your-registry/ncloudserver:v0.0.1
```

**Podman 사용시 (Makefile 통합 - 권장):**
```bash
# 간편한 멀티 아키텍처 빌드 (Makefile 타겟 사용)
make podman-multiarch-build IMG=your-registry/ncloudserver:v0.0.1

# 빌드 후 정리 (선택사항)
make podman-cleanup IMG=your-registry/ncloudserver:v0.0.1

**또는 수동 단계별 빌드:**
# 1. 먼저 manifest 생성
podman manifest create your-registry/ncloudserver:v0.0.1

# 2. 각 아키텍처별로 개별 빌드 및 푸시
# AMD64
podman build \
  --platform linux/amd64 \
  --tag your-registry/ncloudserver:v0.0.1-amd64 \
  .

# ARM64  
podman build \
  --platform linux/arm64 \
  --tag your-registry/ncloudserver:v0.0.1-arm64 \
  .

# 3. 개별 이미지 푸시
podman push your-registry/ncloudserver:v0.0.1-amd64
podman push your-registry/ncloudserver:v0.0.1-arm64

# 4. manifest에 이미지 추가
podman manifest add your-registry/ncloudserver:v0.0.1 docker://your-registry/ncloudserver:v0.0.1-amd64
podman manifest add your-registry/ncloudserver:v0.0.1 docker://your-registry/ncloudserver:v0.0.1-arm64

# 5. 멀티 아키텍처 manifest 푸시
podman manifest push your-registry/ncloudserver:v0.0.1 docker://your-registry/ncloudserver:v0.0.1
```

**배포:**
```bash
# 매니페스트 생성 및 배포
make deploy IMG=your-registry/ncloudserver:v0.0.1
```

## 🔐 네이버 클라우드 인증 설정

### 환경변수 방식

```bash
# 네이버 클라우드 API 키 설정 (참고용 - 권장사항은 Secret 사용)
export NCLOUD_ACCESS_KEY="your_access_key"
export NCLOUD_SECRET_KEY="your_secret_key"

# 한국 리전 기본 설정
export NCLOUD_DEFAULT_REGION="KR"
```

**⚠️ 주의**: 환경변수 방식보다는 **Secret 기반 방식이 더 안전하고 권장됩니다**.

### Kubernetes Secret 방식

```bash
# Namespace 생성
kubectl create namespace ncloud-system

# 네이버 클라우드 인증 정보를 Secret으로 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key-id="your_access_key" \
  --from-literal=secret-access-key="your_secret_key" \
  -n ncloud-system
```

## API 명령어 참고

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
ncloud vserver terminateServerInstances \
  --regionCode KR \
  --serverInstanceNoList "server-instance-no"
```

### 서버 상태 확인
```bash
ncloud vserver getServerInstanceDetail \
  --regionCode KR \
  --serverInstanceNo "server-instance-no"
```

## 테스트 명령어

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

## 배포 및 운영

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

## 실제 생성된 Custom Resource 예시

**실제 구현된 Sample 파일 예시** (`config/samples/server_v1_ncloudserver_minimal.yaml`):
```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: ncloud-server-minimal
spec:
  # 인증 정보 설정
  credentials:
    secretRef:
      name: ncloud-credentials
      # 네임스페이스 생략시 NCloudServer와 같은 네임스페이스 사용

  # 필수 필드들
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
```

**고급 구성 예시** (`config/samples/server_v1_ncloudserver.yaml`):
```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: ncloud-server-sample
  labels:
    app.kubernetes.io/name: ncloud-server-controller
spec:
  credentials:
    secretRef:
      name: ncloud-credentials
      namespace: ncloud-system
  
  regionCode: "KR"
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
  serverProductCode: "SVR.VSVR.STAND.C002.M004.NET.SSD.B050.G001"
  
  serverName: "k8s-worker-node-01"
  loginKeyName: "my-ssh-key"
  accessControlGroupNoList: ["acg-12345"]
  
  isProtectServerTermination: true
  associateWithPublicIp: true
  feeSystemTypeCode: "MTRAT"
  
status:
  phase: "Creating"  # Operator에서 자동 관리
  message: "Server creation initiated"
  lastReconcileTime: "2025-10-02T17:42:25Z"
```

## 디버깅 및 트러블슈팅

```bash
# Operator 로그 실시간 확인
kubectl logs -f -l control-plane=controller-manager -n ncloud-system

# Custom Resource 이벤트 확인
kubectl describe ncloudserver web-server-01

# 네이버 클라우드 API 호출 테스트 (로컬 CLI 사용)
# ※ 주의: ncloud_cli_linux/ 폴더는 .gitignore에 포함되어 GitHub에 업로드되지 않음
./ncloud_cli_linux/ncloud vserver getRegionList
```

## NCloudServer CRD 스펙 정의

이 Operator는 네이버 클라우드 플랫폼의 **VPC 환경**에서 서버를 관리합니다.

### 🔐 인증 정보 관리 (필수 설정)

이 Operator는 **Secret 기반의 안전한 인증 정보 관리**를 지원합니다.

#### 1. 인증용 Secret 생성

```bash
# 방법 1: 직접 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key-id="YOUR_ACCESS_KEY" \
  --from-literal=secret-access-key="YOUR_SECRET_KEY" \
  --namespace=ncloud-system

# 방법 2: 예제 파일 사용
kubectl apply -f config/examples/ncloud-credentials-secret.yaml
```

#### 2. NCloudServer 리소스에 인증 정보 설정

```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: my-server
spec:
  # 인증 정보 설정 (권장)
  credentials:
    secretRef:
      name: ncloud-credentials
      namespace: ncloud-system
      # 다음 필드들은 선택사항 (기본값 사용 가능)
      # accessKeyIDKey: "access-key-id"
      # secretAccessKeyKey: "secret-access-key"
  
  # 기타 서버 설정...
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
```

### 빠른 시작

**전체 CRD 스펙 정의와 사용 가능한 모든 필드, 값들, 예제에 대한 상세 정보는 [CRD 스펙 정의서](docs/CRD_SPECIFICATION.md)를 참고하자.**

### 예제 파일
- `config/samples/server_v1_ncloudserver_minimal.yaml`: 최소 구성 (인증 정보 포함)
- `config/samples/server_v1_ncloudserver.yaml`: 기본 예제 (표준 구성)
- `config/samples/server_v1_ncloudserver_advanced.yaml`: 고급 구성 (모든 기능)

### 사용 예제

```bash
# 1. 인증 정보 설정
kubectl apply -f config/examples/ncloud-credentials-secret.yaml

# 2. 서버 리소스 생성

# 최소 구성으로 간단한 서버 생성
kubectl apply -f config/samples/server_v1_ncloudserver_minimal.yaml

# 표준 구성으로 서버 생성
kubectl apply -f config/samples/server_v1_ncloudserver.yaml

# 고급 기능이 포함된 서버 생성
kubectl apply -f config/samples/server_v1_ncloudserver_advanced.yaml

# 3. 리소스 확인
kubectl get ncloudserver
kubectl describe ncloudserver ncloud-server-minimal
```

### 리소스 삭제 프로세스

이 Operator는 **Finalizer 기반의 안전한 리소스 삭제**를 지원합니다:

```bash
# 리소스 삭제 요청
kubectl delete ncloudserver my-server

# 삭제 프로세스:
# 1. Operator가 NCloud 서버 삭제 실행
# 2. 삭제 완료 확인
# 3. CRD 리소스 완전 삭제
```

**정리하면**: 리소스를 삭제하면 **실제 NCloud 서버도 함께 안전하게 삭제**됩니다.

## 문서

- **[CRD 스펙 정의서](docs/CRD_SPECIFICATION.md)**: NCloudServer CRD의 상세 스펙, 사용 가능한 값들, 예제 매니페스트

## 참고 자료

- [Kubernetes Operator SDK 문서](https://sdk.operatorframework.io/)
- [네이버 클라우드 플랫폼 VPC CLI 가이드](https://cli.ncloud-docs.com/docs/cli-vserver)
- [Go Operator 튜토리얼](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)

## 개발 로드맵

- [x] **개발 환경 설정 완료** (Go 1.23.1 darwin/arm64)
- [x] **Operator SDK 프로젝트 초기화 완료**
- [x] **NCloud VPC CLI 파라미터 조사 완료** (vserver 명령어 분석)
- [x] **VPC 환경 CRD 스펙 정의 완료** (`server.ncloud.devops.ai.kr/v1`)
- [x] **API 및 Controller 스켈레톤 생성** (`NCloudServer` Kind)
- [x] **CRD 매니페스트 생성 완료**
- [x] **예제 매니페스트 작성 완료** (minimal, basic, advanced)
- [x] **CRD 스펙 정의서 작성 완료** (`docs/CRD_SPECIFICATION.md`)
- [x] **NCloud VPC CLI 연동 완료** (경로 문제 해결 및 Docker 통합)
- [x] **서버 생성/삭제 로직 구현 완료** (reconciliation 상태 관리 포함)
- [x] **상태 동기화 구현 완료** (Phase: Creating → Running → Failed/Terminating)
- [x] **Docker 빌드 설정 완료** (CLI 포함 컨테이너 이미지)
- [x] **Secret 기반 인증 정보 관리 구현 완료** (credentials 필드 및 Secret 참조)
- [x] **Finalizer 기반 리소스 삭제 프로세스 구현 완료** (안전한 cleanup)
- [x] **리소스 이름 중복 제거 및 네임스페이스 통일 완료** (`ncloud-system`)
- [x] **매니페스트 완전성 및 검토 완료** (모든 CRD 및 샘플 검증)
- [ ] JSON 응답 파싱 개선 (NCloud CLI 응답 구조 정확한 파싱 - 현재 strings.Contains 수준)
- [ ] 모니터링 및 메트릭 추가
- [ ] 테스트 커버리지 향상
- [ ] 프로덕션 배포 가이드 완성
