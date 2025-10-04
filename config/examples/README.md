# NCloud Server Controller 예제들

이 디렉토리는 NCloud Server Operator의 다양한 사용 예제들을 제공합니다.

## 네임스페이스 정보

이 Operator는 기본적으로 `ncloud-system` 네임스페이스에서 실행됩니다.

## 인증 정보 설정

### 1. Secret 생성

```bash
# 방법 1: CLI로 직접 생성
kubectl create secret generic ncloud-credentials \
  --from-literal=access-key-id="YOUR_ACCESS_KEY" \
  --from-literal=secret-access-key="YOUR_SECRET_KEY" \
  --namespace=ncloud-system

# 방법 2: YAML 파일 사용
kubectl apply -f ncloud-credentials-secret.yaml
```

### 2. 권한 확인

```bash
# Secret 생성 확인
kubectl get secrets -n ncloud-system

# 특정 Secret 상세 정보
kubectl describe secret ncloud-credentials -n ncloud-system
```

## 샘플 리소스들

### 기본 서버 생성

```bash
# 최소 구성으로 서버 생성
kubectl apply -f ../samples/server_v1_ncloudserver_minimal.yaml

# 표준 구성으로 서버 생성  
kubectl apply -f ../samples/server_v1_ncloudserver.yaml

# 고급 구성으로 서버 생성
kubectl apply -f ../samples/server_v1_ncloudserver_advanced.yaml
```

### 리소스 확인

```bash
# 생성된 서버 확인
kubectl get ncloudserver

# 특정 서버 상세 정보
kubectl describe ncloudserver ncloud-server-sample

# 서버 상태 확인
kubectl get ncloudserver -o custom-columns=NAME:.metadata.name,PHASE:.status.phase
```

### 리소스 삭제

```bash
# 특정 서버 삭제
kubectl delete ncloudserver ncloud-server-sample

# 모든 서버 삭제
kubectl delete ncloudserver --all
```

## 네임스페이스별 사용법

### 개발 환경과 운영 환경 분리

```bash
# 개발팀 네임스페이스 생성
kubectl create namespace ncloud-dev

# 개발용 Secret 생성
kubectl create secret generic ncloud-dev-credentials \
  --from-literal=access-key-id="DEV_ACCESS_KEY" \
  --from-literal=secret-access-key="DEV_SECRET_KEY" \
  --namespace=ncloud-dev

# 운영팀 네임스페이스 생성  
kubectl create namespace ncloud-prod

# 운영용 Secret 생성
kubectl create secret generic ncloud-prod-credentials \
  --from-literal=access-key-id="PROD_ACCESS_KEY" \
  --from-literal=secret-access-key="PROD_SECRET_KEY" \
  --namespace=ncloud-prod
```

## 문제 해결

### 포드 상태 확인

```bash
# Operator 포드 상태 확인
kubectl get pods -n ncloud-system

# Operator 로그 확인
kubectl logs -f deployment/ncloud-server-controller-controller-manager -n ncloud-system
```

### 인증 오류 해결

```bash
# Secret 존재 확인
kubectl get secret ncloud-credentials -n ncloud-system

# Secret 내용 확인 (Base64 디코딩)
kubectl get secret ncloud-credentials -n ncloud-system -o jsonpath='{.data.access-key-id}' | base64 -d
```
