# NCloud Server Controller

[![Release](https://img.shields.io/github/v/release/Seo-yul/ncloud-server-controller)](https://github.com/Seo-yul/ncloud-server-controller/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.33+-blue.svg)](https://kubernetes.io/)

네이버 클라우드 플랫폼 VPC 서버를 Kubernetes Custom Resource로 관리하는 Operator입니다.

## 🚀 주요 기능

- **선언적 서버 관리**: YAML로 NCloud 서버 생성/삭제/관리
- **Secret 기반 인증**: 안전한 API 키 관리
- **멀티 아키텍처 지원**: `linux/amd64`, `linux/arm64`
- **Finalizer 기반 삭제**: 안전한 리소스 정리
- **실시간 상태 동기화**: 서버 상태를 Kubernetes에 반영
- **GitOps 지원**: Git 기반 인프라 관리

## 📋 목차

- [빠른 시작](#빠른-시작)
- [설치](#설치)
- [사용법](#사용법)
- [개발 환경 설정](#개발-환경-설정)
- [배포](#배포)
- [문서](#문서)
- [기여하기](#기여하기)

## 🏃‍♂️ 빠른 시작

### 1. Operator 설치

#### 방법 A: GitHub에서 직접 설치 (권장)

```bash
# CRD 설치
kubectl apply -f https://raw.githubusercontent.com/Seo-yul/ncloud-server-controller/develop/config/crd/bases/server.ncloud.devops.ai.kr_ncloudservers.yaml

# Operator 배포 (Kustomize 사용)
kubectl apply -k https://github.com/Seo-yul/ncloud-server-controller/config/default?ref=develop
```

#### 방법 B: 로컬에서 설치

```bash
# 저장소 클론
git clone https://github.com/Seo-yul/ncloud-server-controller.git
cd ncloud-server-controller

# CRD 설치
kubectl apply -f config/crd/bases/server.ncloud.devops.ai.kr_ncloudservers.yaml

# Operator 배포
kubectl apply -k config/default
```

#### 방법 C: Makefile 사용

```bash
# 저장소 클론
git clone https://github.com/Seo-yul/ncloud-server-controller.git
cd ncloud-server-controller

# Makefile로 설치
make install  # CRD만 설치
make deploy   # 전체 배포
```

### 2. 인증 정보 설정

```bash
# 네임스페이스 생성
kubectl create namespace ncloud-system

# NCloud API 키를 Secret으로 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key-id="YOUR_ACCESS_KEY" \
  --from-literal=secret-access-key="YOUR_SECRET_KEY" \
  --namespace=ncloud-system
```

### 3. 서버 생성

```bash
# 최소 구성으로 서버 생성
kubectl apply -f - <<EOF
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: my-server
  namespace: ncloud-system
spec:
  credentials:
    secretRef:
      name: ncloud-credentials
      namespace: ncloud-system
  
  # 필수 필드
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
EOF
```

### 4. 상태 확인

```bash
# 서버 상태 확인
kubectl get ncloudserver my-server -n ncloud-system

# 상세 정보 확인
kubectl describe ncloudserver my-server -n ncloud-system

# Operator 로그 확인
kubectl logs -f deployment/ncloud-server-controller-manager -n ncloud-system
```

## 📦 설치

### Helm 설치 (권장)

```bash
# Helm 차트 추가 (향후 지원 예정)
helm repo add ncloud-controller https://charts.ncloud.devops.ai.kr
helm install ncloud-controller ncloud-controller/ncloud-server-controller
```

### 수동 설치

```bash
# 1. CRD 설치
kubectl apply -f config/crd/bases/server.ncloud.devops.ai.kr_ncloudservers.yaml

# 2. RBAC 설치
kubectl apply -f config/rbac/

# 3. Operator 배포
kubectl apply -f config/default/
```

### 컨테이너 이미지

```bash
# GitHub Container Registry에서 이미지 가져오기
podman pull ghcr.io/seo-yul/ncloud-server-controller:latest

# 또는 특정 버전
podman pull ghcr.io/seo-yul/ncloud-server-controller:v1.0.0
```

## 💡 사용법

### 기본 사용법

```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: web-server
  namespace: production
spec:
  # 인증 정보
  credentials:
    secretRef:
      name: ncloud-credentials
      namespace: ncloud-system
  
  # 서버 설정
  regionCode: "KR"
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
  serverProductCode: "SVR.VSVR.STAND.C002.M004.NET.SSD.B050.G001"
  
  # 서버 정보
  serverName: "web-server-01"
  serverDescription: "Web server for production"
  
  # 보안 설정
  loginKeyName: "production-ssh-key"
  accessControlGroupNoList:
    - "acg-web-servers"
  
  # 기타 설정
  isProtectServerTermination: true
  associateWithPublicIp: true
  feeSystemTypeCode: "MTRAT"
```

### 고급 사용법

```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: database-cluster
spec:
  credentials:
    secretRef:
      name: ncloud-credentials
      namespace: ncloud-system
  
  # 멀티 서버 생성
  serverCreateCount: 3
  serverCreateStartNo: 1
  serverName: "db-node"
  
  # 고급 네트워킹
  networkInterfaceList:
    - networkInterfaceOrder: 0
      accessControlGroupNoList: ["acg-default"]
    - networkInterfaceOrder: 1
      subnetNo: "subnet-database"
      accessControlGroupNoList: ["acg-database"]
  
  # GPU 설정
  fabricClusterPoolNo: "gpu-cluster-pool-12345"
  isPreInstallGpuDriver: true
  
  # 추가 스토리지
  blockStorageMappingList:
    - order: 1
      blockStorageSize: "500"
      blockStorageName: "data-volume"
      encrypted: true
  
  # 기타 필수 필드
  serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"
  vpcNo: "vpc-67890"
  subnetNo: "subnet-67890"
```

### 예제 파일

- **최소 구성**: `config/samples/server_v1_ncloudserver_minimal.yaml`
- **기본 구성**: `config/samples/server_v1_ncloudserver.yaml`
- **고급 구성**: `config/samples/server_v1_ncloudserver_advanced.yaml`

## 🛠️ 개발 환경 설정

### 필수 도구

```bash
# Go 1.24+ 설치
go version

# kubectl 설치
brew install kubectl  # macOS
# 또는 https://kubernetes.io/docs/tasks/tools/ 에서 다운로드

# Operator SDK 설치
brew install operator-sdk
operator-sdk version
```

### 프로젝트 클론 및 설정

```bash
# 프로젝트 클론
git clone https://github.com/Seo-yul/ncloud-server-controller.git
cd ncloud-server-controller

# 의존성 설치
go mod download

# 코드 생성
make generate

# 매니페스트 생성
make manifests
```

### 로컬 개발

```bash
# 테스트 실행
make test

# 린트 검사
make lint

# 로컬에서 Operator 실행
make run

# 크로스 컴파일 빌드
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make build
```

### 개발 워크플로우

```bash
# 1. 코드 수정
vim internal/controller/ncloudserver_controller.go

# 2. 테스트 실행
make test

# 3. 린트 검사
make lint

# 4. 매니페스트 업데이트
make manifests

# 5. 로컬 테스트
make run
```

## 🚀 배포

### GitHub Actions를 통한 자동 배포

```bash
# 새 버전 태그 생성
git tag v1.0.1
git push origin v1.0.1

# GitHub Actions가 자동으로:
# 1. 멀티 아키텍처 이미지 빌드
# 2. GHCR에 이미지 푸시
# 3. GitHub Release 생성
# 4. 매니페스트 업데이트
```

### 수동 배포

```bash
# 1. 이미지 빌드 및 푸시
make docker-build docker-push IMG=ghcr.io/seo-yul/ncloud-server-controller:v1.0.1

# 2. 매니페스트 업데이트
make deploy IMG=ghcr.io/seo-yul/ncloud-server-controller:v1.0.1
```

### 프로덕션 배포

```bash
# 1. CRD 설치
kubectl apply -f config/crd/bases/server.ncloud.devops.ai.kr_ncloudservers.yaml

# 2. RBAC 설치
kubectl apply -f config/rbac/

# 3. Operator 배포
kubectl apply -f config/default/

# 4. 배포 확인
kubectl get pods -n ncloud-system
kubectl logs -f deployment/ncloud-server-controller-manager -n ncloud-system
```

## 📚 문서

- **[CRD 스펙 정의서](docs/CRD_SPECIFICATION.md)**: NCloudServer CRD의 상세 스펙, 사용 가능한 값들, 예제 매니페스트
- **[API 문서](https://github.com/Seo-yul/ncloud-server-controller/tree/main/api/v1)**: Go 코드 기반 API 문서
- **[예제 모음](config/samples/)**: 다양한 사용 사례별 예제

## 🔧 트러블슈팅

### 일반적인 문제

#### 1. Operator가 시작되지 않음

```bash
# 파드 상태 확인
kubectl get pods -n ncloud-system

# 로그 확인
kubectl logs deployment/ncloud-server-controller-manager -n ncloud-system

# RBAC 권한 확인
kubectl auth can-i get leases --as=system:serviceaccount:ncloud-system:ncloud-server-controller-manager
```

#### 2. 서버 생성 실패

```bash
# NCloudServer 리소스 상태 확인
kubectl describe ncloudserver my-server

# Operator 로그 확인
kubectl logs -f deployment/ncloud-server-controller-manager -n ncloud-system

# Secret 확인
kubectl get secret ncloud-credentials -n ncloud-system -o yaml
```

#### 3. 인증 오류

```bash
# Secret 내용 확인
kubectl get secret ncloud-credentials -n ncloud-system -o jsonpath='{.data.access-key-id}' | base64 -d
kubectl get secret ncloud-credentials -n ncloud-system -o jsonpath='{.data.secret-access-key}' | base64 -d

# Secret 재생성
kubectl delete secret ncloud-credentials -n ncloud-system
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key-id="YOUR_ACCESS_KEY" \
  --from-literal=secret-access-key="YOUR_SECRET_KEY" \
  --namespace=ncloud-system
```

### 디버깅 명령어

```bash
# Operator 로그 실시간 확인
kubectl logs -f deployment/ncloud-server-controller-manager -n ncloud-system

# Custom Resource 이벤트 확인
kubectl describe ncloudserver my-server

# 모든 NCloudServer 리소스 확인
kubectl get ncloudserver --all-namespaces

# Operator 메트릭 확인
kubectl port-forward deployment/ncloud-server-controller-manager 8443:8443 -n ncloud-system
curl -k https://localhost:8443/metrics
```

## 🤝 기여하기

### 기여 방법

1. **Fork** 이 저장소
2. **Feature 브랜치** 생성 (`git checkout -b feature/amazing-feature`)
3. **변경사항 커밋** (`git commit -m 'Add amazing feature'`)
4. **브랜치 푸시** (`git push origin feature/amazing-feature`)
5. **Pull Request** 생성

### 개발 가이드라인

- **코드 스타일**: `make lint` 통과 필수
- **테스트**: 새로운 기능에 대한 테스트 작성
- **문서**: 변경사항에 대한 문서 업데이트
- **커밋 메시지**: [Conventional Commits](https://www.conventionalcommits.org/) 형식 사용

### 이슈 리포트

버그 리포트나 기능 요청은 [GitHub Issues](https://github.com/Seo-yul/ncloud-server-controller/issues)를 통해 제출해주세요.

## 📄 라이선스

이 프로젝트는 [MIT 라이선스](LICENSE) 하에 배포됩니다.

## 🙏 감사의 말

- [Kubernetes Operator SDK](https://sdk.operatorframework.io/) - Operator 개발 프레임워크
- [네이버 클라우드 플랫폼](https://www.ncloud.com/) - 클라우드 인프라 제공
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes 컨트롤러 런타임

## 📊 프로젝트 상태

### ✅ 완료된 기능

- [x] **CRD 정의**: NCloudServer Custom Resource 완성
- [x] **컨트롤러 구현**: 서버 생성/삭제/상태 관리
- [x] **Secret 기반 인증**: 안전한 API 키 관리
- [x] **Finalizer 기반 삭제**: 안전한 리소스 정리
- [x] **멀티 아키텍처 지원**: `linux/amd64`, `linux/arm64`
- [x] **GitHub Actions CI/CD**: 자동 빌드 및 배포
- [x] **문서화**: 완전한 CRD 스펙 정의서
- [x] **예제 매니페스트**: 다양한 사용 사례

### 🚧 진행 중인 작업

- [ ] **JSON 응답 파싱 개선**: NCloud CLI 응답 구조 정확한 파싱
- [ ] **테스트 커버리지 향상**: 현재 22.5% → 목표 80%+
- [ ] **모니터링 및 메트릭**: Prometheus 메트릭 추가
- [ ] **웹훅 검증**: 입력 검증 웹훅 구현

### 🔮 향후 계획

- [ ] **Helm 차트**: 설치 및 관리 자동화
- [ ] **Operator Lifecycle Manager**: OLM 지원
- [ ] **고급 기능**: 자동 스케일링, 백업 관리
- [ ] **멀티 리전 지원**: 여러 리전 동시 관리

---

**문의사항이나 제안사항이 있으시면 [GitHub Issues](https://github.com/Seo-yul/ncloud-server-controller/issues)를 통해 연락해주세요!** 🚀