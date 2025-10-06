# NCloud Server Controller Helm Chart

이 Helm 차트는 NCloud VPC 서버를 관리하는 Kubernetes Operator를 배포합니다.

## 설치

### 사전 요구사항

- Kubernetes 1.19+
- Helm 3.0+
- NCloud API 자격 증명

### 기본 설치

```bash
# OCI 레지스트리에서 직접 설치 (권장)
helm install ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller

# 특정 버전 설치
helm install ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller --version 1.0.0

# 네임스페이스 지정하여 설치
helm install ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller \
  --namespace ncloud-system --create-namespace
```

### 사용자 정의 값으로 설치

```bash
# values.yaml 파일 생성
cat > my-values.yaml << EOF
operator:
  image:
    tag: "v1.0.0"
  replicaCount: 2
  resources:
    limits:
      cpu: 1000m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi

namespace:
  name: "my-ncloud-system"

metrics:
  enabled: true
  serviceMonitor:
    enabled: true
EOF

# 사용자 정의 값으로 설치
helm install ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller -f my-values.yaml
```

## 설정

### 주요 설정 옵션

| 매개변수 | 설명 | 기본값 |
|---------|------|--------|
| `operator.image.repository` | 컨테이너 이미지 저장소 | `ghcr.io/seo-yul/ncloud-server-controller` |
| `operator.image.tag` | 컨테이너 이미지 태그 | `$(VERSION)` (현재 1.0.0) |
| `operator.replicaCount` | Operator 복제본 수 | `1` |
| `operator.resources.limits.cpu` | CPU 제한 | `500m` |
| `operator.resources.limits.memory` | 메모리 제한 | `128Mi` |
| `namespace.name` | 네임스페이스 이름 | `ncloud-system` |
| `rbac.create` | RBAC 리소스 생성 여부 | `true` |
| `crds.install` | CRD 설치 여부 | `true` |
| `metrics.enabled` | 메트릭 서비스 활성화 | `true` |
| `metrics.serviceMonitor.enabled` | ServiceMonitor 생성 여부 | `false` |

### 고급 설정

#### 자동 스케일링

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

#### 네트워크 정책

```yaml
networkPolicy:
  enabled: true
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8080
```

#### Pod 중단 예산

```yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

## 사용법

### 1. NCloud 자격 증명 설정

```bash
# NCloud API 자격 증명을 포함한 Secret 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=accessKeyID=YOUR_ACCESS_KEY \
  --from-literal=secretAccessKey=YOUR_SECRET_KEY \
  -n ncloud-system
```

### 2. NCloudServer 리소스 생성

```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: my-server
  namespace: ncloud-system
spec:
  vpcNo: "vpc-12345"
  subnetNo: "subnet-12345"
  serverImageProductCode: "SPSW0LINUX000046"
  serverProductCode: "SPSVRSTAND000004"
  loginKeyName: "my-key"
  credentials:
    secretRef:
      name: ncloud-credentials
```

### 3. 리소스 확인

```bash
# NCloudServer 리소스 확인
kubectl get ncloudservers -n ncloud-system

# Operator 로그 확인
kubectl logs -n ncloud-system -l app.kubernetes.io/name=ncloud-server-controller -f
```

## 업그레이드

```bash
# 차트 업그레이드
helm upgrade ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller

# 특정 버전으로 업그레이드
helm upgrade ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller --version 1.1.0

# 이미지 태그만 변경하여 업그레이드
helm upgrade ncloud-server-controller oci://ghcr.io/seo-yul/ncloud-server-controller --set operator.image.tag=v1.1.0
```

## 제거

```bash
# Helm 릴리스 제거
helm uninstall ncloud-server-controller

# CRD도 함께 제거하려면
helm uninstall ncloud-server-controller --no-hooks
kubectl delete crd ncloudservers.server.ncloud.devops.ai.kr
```

## 문제 해결

### 일반적인 문제

1. **Operator가 시작되지 않음**
   ```bash
   kubectl describe pod -n ncloud-system -l app.kubernetes.io/name=ncloud-server-controller
   kubectl logs -n ncloud-system -l app.kubernetes.io/name=ncloud-server-controller
   ```

2. **RBAC 권한 오류**
   ```bash
   kubectl auth can-i create ncloudservers --as=system:serviceaccount:ncloud-system:ncloud-server-controller-manager
   ```

3. **이미지 풀 오류**
   ```bash
   kubectl get events -n ncloud-system --sort-by='.lastTimestamp'
   ```

### 로그 확인

```bash
# Operator 로그
kubectl logs -n ncloud-system -l app.kubernetes.io/name=ncloud-server-controller -f

# 특정 Pod 로그
kubectl logs -n ncloud-system deployment/ncloud-server-controller-manager -f
```

## 개발

### 로컬 개발

```bash
# 차트 템플릿 렌더링 테스트
helm template ncloud-server-controller ./helm/ncloud-server-controller

# 린트 검사
helm lint ./helm/ncloud-server-controller

# 테스트 실행
helm test ncloud-server-controller
```

### Helm 차트 생성 및 배포

#### 로컬에서 Helm 차트 생성 및 업로드

```bash
# 1단계: 기본 빌드
make manifests && make generate && make build

# 2단계: 이미지 빌드 (Podman 사용)
make podman-multiarch-build

# 3단계: Helm 차트 생성 및 푸시
make helm-generate && make helm-package && make helm-lint && make helm-push
```

#### 한 번에 모든 작업 수행

```bash
# 완전한 Helm 차트 생성 및 업로드 (권장)
make helm-all-with-push

# 또는 단계별 실행
make helm-all  # 생성 + 패키징 + 린팅 (업로드 제외)
make helm-push # OCI 레지스트리에 업로드
```

#### Helm 차트 테스트

```bash
# Helm 차트 설치 테스트 (Dry Run)
make helm-dry-run

# 로컬에서 Helm 차트 린팅
make helm-lint

# Helm 차트 템플릿 렌더링 테스트
helm template ncloud-controller helm/ncloud-server-controller/
```

### 차트 패키징

```bash
# 차트 패키지 생성
helm package ./helm/ncloud-server-controller

# OCI 레지스트리에 업로드
helm push ncloud-server-controller-1.0.0.tgz oci://ghcr.io/seo-yul
```

## 기여

이 차트에 기여하려면:

1. 이슈를 생성하여 변경사항을 논의
2. 포크를 생성하고 기능 브랜치 생성
3. 변경사항을 커밋하고 푸시
4. Pull Request 생성

## 라이선스

MIT License
