# NCloudServer CRD 스펙 정의서

네이버 클라우드 플랫폼 VPC 환경에서 사용되는 NCloudServer Custom Resource의 상세 스펙 정의서입니다.

## 목차
- [CRD 개요](#crd-개요)
- [Spec 필드 정의](#spec-필드-정의)
  - [인증 설정](#인증-설정)
  - [필수 필드](#필수-필드-required-fields)
  - [이미지 설정](#이미지-설정-one-of-required)
  - [서버 스펙 설정](#서버-스펙-설정)
  - [네트워킹 설정](#네트워킹-설정)
  - [보안 설정](#보안-설정)
  - [고급 설정](#고급-설정)
  - [저장소 설정](#저장소-설정)
  - [GPU 설정](#gpu-설정)
  - [디버깅 설정](#디버깅-설정)
- [Status 필드 정의](#status-필드-정의)
  - [서버 인스턴스 정보](#서버-인스턴스-정보)
  - [서버 리소스 정보](#서버-리소스-정보)
  - [네트워킹 정보](#네트워킹-정보)
  - [상태 정보](#상태-정보)
  - [플랫폼 정보](#플랫폼-정보)
  - [타임스탬프 정보](#타임스탬프-정보)
  - [Kubernetes 상태 정보](#kubernetes-상태-정보)
  - [보안 및 접근 정보](#보안-및-접근-정보)
  - [리전 및 존 정보](#리전-및-존-정보)
  - [배치 및 GPU 정보](#배치-및-gpu-정보)
- [유효성 검증 규칙](#유효성-검증-규칙)
- [예제 매니페스트](#예제-매니페스트)
- [관련 외부 리소스](#관련-외부-리소스)

## CRD 개요

### API Group 및 Version
- **API Group**: `server.ncloud.devops.ai.kr`
- **Version**: `v1`
- **Kind**: `NCloudServer`

### 기본 구조
```yaml
apiVersion: server.ncloud.devops.ai.kr/v1
kind: NCloudServer
metadata:
  name: my-ncloud-server
spec:
  # 서버 생성 설정
status:
  # 서버 상태 정보
```

## Spec 필드 정의

### 인증 설정

#### `credentials`
**설명**: NCloud 인증 설정  
**타입**: `CredentialsSpec`  
**설명**: NCloud API 접근을 위한 자격 증명 정보입니다.

**예제**:
```yaml
credentials:
  secretRef:
    name: ncloud-credentials
    namespace: ncloud-system
```

#### `CredentialsSpec` 구조
- `secretRef`: Kubernetes Secret 참조

#### `SecretReference` 구조
- `name`: Secret 이름 (필수)
- `namespace`: Secret 네임스페이스 (선택, 기본값: NCloudServer와 같은 네임스페이스)
- `accessKeyIDKey`: Access Key ID가 저장된 키 (기본값: "access-key-id")
- `secretAccessKeyKey`: Secret Access Key가 저장된 키 (기본값: "secret-access-key")

**예제**:
```yaml
credentials:
  secretRef:
    name: ncloud-credentials
    namespace: ncloud-system
    # accessKeyIDKey: "access-key-id"        # 기본값 사용시 생략 가능
    # secretAccessKeyKey: "secret-access-key" # 기본값 사용시 생략 가능
```

**커스텀 키 사용 예제**:
```yaml
credentials:
  secretRef:
    name: my-ncloud-secret
    namespace: production
    accessKeyIDKey: "ncloud-access-key"      # 커스텀 키명
    secretAccessKeyKey: "ncloud-secret-key"  # 커스텀 키명
```

### 필수 필드 (Required Fields)

#### `vpcNo`
**설명**: VPC 번호  
**타입**: `string`  
**제약**: 필수 입력  
**설명**: 서버가 생성될 VPC를 결정합니다. `getVpcList` 액션으로 획득 가능합니다.

**예제**:
```yaml
vpcNo: "vpc-12345678901234567"
```

#### `subnetNo`  
**설명**: 서브넷 번호  
**타입**: `string`  
**제약**: 필수 입력  
**설명**: 기본 네트워크 인터페이스가 사용할 서브넷을 결정합니다. `getSubnetList` 액션으로 획득 가능합니다.

**예제**:
```yaml
subnetNo: "subnet-12345678901234567"
```

### 이미지 설정 (One of Required)

#### `serverImageProductCode`
**설명**: 서버 이미지 상품 코드  
**타입**: `string`  
**제약**: 새 이미지 사용시 필수  
**설명**: 네이버 클라우드에서 제공하는 공식 이미지 상품 코드입니다.

**사용 가능한 값들**:
```yaml
# Linux 운영체제
serverImageProductCode: "SW.VSVR.OS.LNX64.UBNTU.SVR2004.B050"    # Ubuntu 20.04
serverImageProductCode: "SW.VSVR.OS.LNX64.ROCKY.0810.B050"       # Rocky Linux 8.10
serverImageProductCode: "SW.VSVR.OS.LNX64.ROCKY.0808.B050"       # Rocky Linux 8.8
serverImageProductCode: "SW.VSVR.OS.LNX64.ROCKY.0806.B050"       # Rocky Linux 8.6

# Windows 운영체제
serverImageProductCode: "SW.VSVR.OS.WND64.WND.SVR2019EN.B100"    # Windows Server 2019
serverImageProductCode: "SW.VSVR.OS.WND64.WND.SVR2016EN.B100"    # Windows Server 2016

# AI/ML 전용
serverImageProductCode: "SW.VSVR.APP.LNX64.UBNTU.SVR2004.TNSFL.LATEST.B050"  # Ubuntu 20.04 + TensorFlow
serverImageProductCode: "SW.VSVR.APP.LNX64.UBNTU.SVR1804.TNSFL.LATEST.B050"  # Ubuntu 18.04 + TensorFlow

# 데이터베이스 전용
serverImageProductCode: "SW.VSVR.DBMS.WND64.WND.SVR2019STDEN.MSSQL.2022.B100"  # Windows 2019 + SQL Server 2022
serverImageProductCode: "SW.VSVR.DBMS.WND64.WND.SVR2019STDEN.MSSQL.2019.B100"  # Windows 2019 + SQL Server 2019
serverImageProductCode: "SW.VSVR.DBMS.LNX64.ROCKY.0808.TBERO.CSE7.B050"        # Rocky Linux + Tibero CSE
```

#### `memberServerImageInstanceNo`
**설명**: 회원 서버 이미지 인스턴스 번호  
**타입**: `string`  
**제약**: 커스텀 이미지 사용시 필수  
**설명**: 사용자가 직접 생성한 서버 이미지로부터 서버를 생성할 때 사용합니다.

**예제**:
```yaml
memberServerImageInstanceNo: "img-instance-12345678901234567"
```

#### `serverImageNo`
**설명**: 서버 이미지 번호  
**타입**: `string`  
**제약**: 대안 이미지 옵션  
**설명**: 이미지 상품 코드나 회원 서버 이미지와 동시 사용 불가능합니다.

**예제**:
```yaml
serverImageNo: "image-12345678901234567"
```

### 서버 스펙 설정

#### `serverProductCode`
**설명**: 서버 상품 코드  
**타입**: `string`  
**설명**: 생성할 서버의 스펙을 결정합니다. CPU, 메모리, 스토리지 스펙이 포함됩니다.

**획득 방법**: `getServerProductList` 액션으로 획득 가능  
**예제**:
```yaml
serverProductCode: "SVR.VSVR.STAND.C002.M004.NET.SSD.B050.G001"  # 2CPU 4GB 표준 서버
serverProductCode: "SVR.VSVR.HIGHMEM.C004.M008.NET.SSD.B050.G001" # 4CPU 8GB 고메모리 서버
```

#### `serverSpecCode`
**설명**: 서버 사양 코드  
**타입**: `string`  
**설명**: `serverProductCode`의 대안 옵션입니다.

#### `feeSystemTypeCode`
**설명**: 요금제 유형 코드  
**타입**: `string`  
**기본값**: `MTRAT`  
**설명**: 서버 요금 계산 방식을 결정합니다.

**사용 가능한 값**:
```yaml
feeSystemTypeCode: "MTRAT"  # 시간 요금제 (종량제)
feeSystemTypeCode: "FXSUM"  # 월 요금제 (정액제)
```

### 서버 이름 및 설명

#### `serverName`
**설명**: 서버 이름  
**타입**: `string`  
**길이 제한**: 3-30자  
**문자 제한**: 소문자, 숫자, "-" 문자만 허용, 알파벳으로 시작해야 함  
**기본값**: NAVER CLOUD PLATFORM이 자동 할당

**예제**:
```yaml
serverName: "k8s-worker-node-01"
serverName: "web-server-prod"
serverName: "db-server-master"
```

#### `serverDescription`
**설명**: 서버 설명  
**타입**: `string`  
**길이 제한**: 0-1000 바이트

**예제**:
```yaml
serverDescription: "Kubernetes worker node for production cluster"
serverDescription: "Web server for frontend application"
```

### 확장 설정

#### `serverCreateCount`
**설명**: 서버 생성 개수  
**타입**: `integer`  
**범위**: 1-10개  
**기본값**: 1개

**예제**:
```yaml
serverCreateCount: 3  # 한 번에 3대 생성
serverCreateCount: 1   # 단일 서버 생성
```

#### `serverCreateStartNo`
**설명**: 서버 생성 시작 번호  
**

타입**: `integer`  
**범위**: 0-999  
**설명**: 여러 대 생성시 서버명에 붙는 일련번호의 시작 번호입니다.

**예제**:
```yaml
serverCreateCount: 3
serverCreateStartNo: 10
# 결과: k8s-node-10, k8s-node-11, k8s-node-12
```

### 네트워킹 설정

#### `networkInterfaceList`
**설명**: 네트워크 인터페이스 설정 배열  
**타입**: `array[NetworkInterfaceConfiguration]`  
**최대 개수**: 3개 (기본 인터페이스 포함)

```yaml
networkInterfaceList:
  - networkInterfaceOrder: 0                    # 기본 인터페이스 (필수)
    accessControlGroupNoList:
      - "acg-default"
  - networkInterfaceOrder: 1                    # 추가 인터페이스
    subnetNo: "subnet-67890"
    ip: "10.0.1.100"
    accessControlGroupNoList:
      - "acg-web-servers"
      - "acg-database-servers"
```

#### `NetworkInterfaceConfiguration` 구조
- `networkInterfaceOrder`: 인터페이스 순서 (0-2)
- `networkInterfaceNo`: 기존 인터페이스 번호 (옵션)
- `subnetNo`: 서브넷 번호 (새 인터페이스시 필수)
- `ip`: 특정 IP 주소 (옵션)
- `accessControlGroupNoList`: 적용할 ACG 목록

#### `accessControlGroupNoList`
**설명**: 접근 제어 그룹 번호 목록  
**타입**: `array[string]`  
**최대 개수**: 5개

**예제**:
```yaml
accessControlGroupNoList:
  - "acg-12345678901234567"  # 기본 ACG
  - "acg-98765432109876543"  # 웹 서버 ACG
```

### 보안 설정

#### `loginKeyName`
**설명**: 로그인 키 이름  
**타입**: `string`  
**설명**: SSH 접속에 사용될 키 페어입니다.

**획득 방법**: `getLoginKeyList` 액션으로 획득 가능  
**예제**:
```yaml
loginKeyName: "production-ssh-key"
loginKeyName: "my-develop-keypair"
```

#### `isProtectServerTermination`
**설명**: 반납 보호 여부  
**타입**: `boolean`  
**기본값**: `false`  
**설명**: 실수로 서버 반납하는 것을 방지합니다.

**예제**:
```yaml
isProtectServerTermination: true   # 보호 활성화
isProtectServerTermination: false  # 보호 비활성화
```

### 고급 설정

#### `regionCode`
**설명**: 리전 코드  
**타입**: `string`  
**기본값**: getRegionList 조회 결과 첫 번째 리전

**사용 가능한 값**:
```yaml
regionCode: "KR"      # 한국 리전 (권장)
regionCode: "US"      # 미국 리전
regionCode: "SG"      # 싱가포르 리전
regionCode: "JP"      # 일본 리전
regionCode: "HK"      # 홍콩 리전
regionCode: "DE"      # 독일 리전
```

#### `placementGroupNo`
**설명**: 물리 배치 그룹 번호  
**타입**: `string`  
**설명**: 서버 인스턴스의 소속 물리 배치 그룹을 결정합니다.

**획득 방법**: `getPlacementGroupList` 액션으로 획득 가능  
**예제**:
```yaml
placementGroupNo: "place-group-12345678901234567"
```

#### `initScriptNo`
**설명**: 초기화 스크립트 번호  
**타입**: `string`  
**설명**: 서버 최초 부팅시 실행될 초기화 스크립트입니다.

**획득 방법**: `getInitScriptList` 액션으로 획득 가능  
**예제**:
```yaml
initScriptNo: "init-script-12345678901234567"
```

#### `associateWithPublicIp`
**설명**: 공인 IP 자동 할당 여부  
**타입**: `boolean`  
**기본값**: `false`  
**제약**: 서버 생성 개수가 1개일 때만 유효, Public Subnet에서만 가능

**예제**:
```yaml
associateWithPublicIp: true   # 자동 공인 IP 할당
associateWithPublicIp: false  # 수동으로 나중에 할당
```

### 저장소 설정

#### `isEncryptedBaseBlockStorageVolume`
**설명**: 기본 블록 스토리지 볼륨 암호화 여부  
**타입**: `boolean`  
**기본값**: `false`

**예제**:
```yaml
isEncryptedBaseBlockStorageVolume: true   # 암호화 활성화
isEncryptedBaseBlockStorageVolume: false  # 암호화 비활성화
```

#### `blockDevicePartitionList`
**설명**: 파티션 설정 (베어메탈 서버용)  
**타입**: `array[BlockDevicePartition]`

```yaml
blockDevicePartitionList:
  - mountPoint: "/"           # 루트 파티션 (필수)
    partitionSize: "100"      # 100 GiB
  - mountPoint: "/var"        # Var 파티션
    partitionSize: "50"       # 50 GiB
  - mountPoint: "/home"       # Home 파티션
    partitionSize: "200"     # 200 GiB
```

#### `BlockStorageMapping`
**설명**: 추가 스토리지 블록 매핑 (KVM 전용)  
컴플렉스한 스토리지 설정을 위한 구조체입니다.

```yaml
blockStorageMappingList:
  - order: 1
    blockStorageSize: "500"
    blockStorageName: "data-volume"
    blockStorageVolumeTypeCode: "SSD"
    encrypted: true
    emptyBlockStorage: true
  - order: 2
    snapshotInstanceNo: "snapshot-12345678901234567"
    blockStorageSize: "1000"
    blockStorageName: "backup-volume"
    encrypted: false
```

### Bare Metal 전용 설정

#### `raidTypeName`
**설명**: RAID 유형 이름  
**타입**: `string`  
**제약**: 베어메탈 서버 생성시 필수

**획득 방법**: `getRaidList` 액션으로 획득 가능

**예제**:
```yaml
raidTypeName: "RAID1"   # RAID 1 설정
raidTypeName: "RAID0"   # RAID 0 설정
raidTypeName: "RAID5"   # RAID 5 설정
```

### GPU 설정

#### `fabricClusterPoolNo`
**설명**: GPU Fabric Cluster Pool 번호  
**타입**: `string`  
**제약**: KVM GPU 서버에서만 사용 가능

**획득 방법**: `getFabricClusterPoolList` 액션으로 획득 가능

**예제**:
```yaml
fabricClusterPoolNo: "gpu-cluster-pool-12345678901234567"
```

#### `isPreInstallGpuDriver`
**설명**: GPU Driver 사전 설치 여부  
**타입**: `boolean`  
**기본값**: `false`

**예제**:
```yaml
isPreInstallGpuDriver: true   # GPU 드라이버 자동 설치
isPreInstallGpuDriver: false  # 수동 설치 필요
```

### 디버깅 설정

#### `debugMode`
**설명**: 디버그 모드 활성화 여부  
**타입**: `boolean`  
**기본값**: `false`  
**설명**: Operator에서 상세한 로그 출력을 활성화합니다.

**예제**:
```yaml
debugMode: true   # 디버그 로그 활성화
debugMode: false  # 일반 로그 사용
```

## Status 필드 정의

### 서버 인스턴스 정보

#### `serverInstanceNo`
**설명**: NCloud 서버 인스턴스 번호  
**타입**: `string`  
**설명**: 네이버 클라우드에서 할당한 실서버 인스턴스 고유 번호입니다.

**예제**:
```yaml
serverInstanceNo: "397255"
```

#### `serverName`
**설명**: 실제 서버 이름  
**타입**: `string`  
**설명**: NCloud에서 실제 생성된 서버명입니다.

#### `serverDescription`
**설명**: 실제 서버 설명  
**타입**: `string`  
**설명**: NCloud에서 실제 적용된 서버 설명입니다.

### 서버 리소스 정보

#### `cpuCount`
**설명**: CPU 개수  
**타입**: `integer`

**예제**:
```yaml
cpuCount: 2  # 2 CPU 코어
```

#### `memorySize`
**설명**: 메모리 크기  
**타입**: `int64`  
**단위**: 바이트

**예제**:
```yaml
memorySize: 4294967296  # 4GB 메모리
```

### 네트워킹 정보

#### `publicIp`
**설명**: 공인 IP 주소  
**타입**: `string`

**예제**:
```yaml
publicIp: "52.79.123.45"
```

#### `privateIp`
**설명**: 사설 IP 주소  
**타입**: `string`

**예제**:
```yaml
privateIp: "10.113.245.112"
```

#### `vpcNo`
**설명**: VPC 번호  
**타입**: `string`

#### `subnetNo`
**설명**: 서브넷 번호  
**타입**: `string`

### 상태 정보

#### `serverInstanceStatus`
**설명**: 서버 인스턴스 상태  
**타입**: `StatusCode`

**사용 가능한 상태**:
```yaml
serverInstanceStatus:
  code: "CREATDT"      # 생성중
  codeName: "Server CREATDT State"
---
serverInstanceStatus:
  code: "RUN"          # 운영중
  codeName: "Server RUN State"  
---
serverInstanceStatus:
  code: "INIT"         # 초기화중
  codeName: "Server INIT State"
---
serverInstanceStatus:
  code: "STOPDT"       # 정지됨
  codeName: "Server STOPDT State"
```

#### `serverInstanceOperation`
**설명**: 현재 진행중인 작업  
**타입**: `StatusCode`

**사용 가능한 작업**:
```yaml
serverInstanceOperation:
  code: "NULL"         # 작업 없음
  codeName: "Server NULL OP"
---
serverInstanceOperation:
  code: "STARTBT"      # 부팅중
  codeName: "Server STARTBT OP"
---
serverInstanceOperation:
  code: "STOPBT"       # 종료중
  codeName: "Server STOPBT OP"
```

#### `serverInstanceStatusName`
**설명**: 서버 인스턴스 상태명  
**타입**: `string`  
**설명**: 사람이 읽기 쉬운 상태 설명입니다.

**예제**:
```yaml
serverInstanceStatusName: "Server RUN State"
serverInstanceStatusName: "Server CREATDT State"
```

### 플랫폼 정보

#### `platformType`
**설명**: 플랫폼 타입  
**타입**: `StatusCode`

**사용 가능한 타입**:
```yaml
platformType:
  code: "LNX64"       # Linux 64 Bit
  codeName: "Linux 64 Bit"
---
platformType:
  code: "WND64"       # Windows 64 Bit  
  codeName: "Windows 64 Bit"
---
platformType:
  code: "UBS64"       # Ubuntu Server 64 Bit
  codeName: "Ubuntu Server 64 Bit"
```

#### `hypervisorType`
**설명**: 하이퍼바이저 타입  
**타입**: `StatusCode`

#### `serverImageName`
**설명**: 서버 이미지 이름  
**타입**: `string`  
**설명**: 실제 사용된 이미지의 이름입니다.

**예제**:
```yaml
serverImageName: "Ubuntu Server 20.04"
serverImageName: "Rocky Linux 8.10"
```

#### `serverInstanceType`
**설명**: 서버 인스턴스 타입  
**타입**: `StatusCode`  
**설명**: 서버의 인스턴스 타입 정보입니다.

### 타임스탬프 정보

#### `createDate`
**설명**: 생성 시각  
**타입**: `string`  
**형식**: RFC3339 (ISO 8601)

**예제**:
```yaml
createDate: "2020-08-24T09:41:23+0900"
```

#### `uptime`
**설명**: 마지막 부팅 시각  
**타입**: `string`  
**형식**: RFC3339 (ISO 8601)

### Kubernetes 상태 정보

#### `phase`
**설명**: 조정 단계  
**타입**: `string`  
**설명**: Operator에서 관리하는 Kubernetes 상태 단계입니다.

**사용 가능한 단계**:
```yaml
phase: "Pending"      # 대기중 (생성 요청됨)
phase: "Creating"      # 생성중
phase: "Running"       # 서버 운영중
phase: "Failed"         # 실패
phase: "Terminating"   # 삭제중
phase: "Terminated"   # 삭제됨
```

#### `message`
**설명**: 상태 메시지  
**타입**: `string`  
**설명**: 현재 상태에 대한 설명입니다.

**예제**:
```yaml
message: "Server instance created successfully"
message: "Waiting for server to become available"
message: "Failed to create server: Invalid VPC configuration"
```

#### `lastReconcileTime`
**설명**: 마지막 조정 시각  
**타입**: `string`  
**형식**: RFC3339 (ISO 8601)

#### `observedGeneration`
**설명**: 마자막 처리된 Generation  
**타입**: `int64`  
**설명**: Operator가 마지막으로 처리한 Resource Generation 번호입니다.

### 보안 및 접근 정보

#### `loginKeyName`
**설명**: 사용된 로그인 키 이름  
**타입**: `string`  
**설명**: 실제 서버에 적용된 SSH 키 이름입니다.

#### `networkInterfaceNoList`
**설명**: 네트워크 인터페이스 번호 목록  
**타입**: `array[string]`  
**설명**: 서버에 할당된 네트워크 인터페이스 번호들입니다.

#### `isProtectServerTermination`
**설명**: 서버 종료 보호 상태  
**타입**: `boolean`  
**설명**: 서버 종료 보호가 활성화되어 있는지 여부입니다.

### 리전 및 존 정보

#### `regionCode`
**설명**: 리전 코드  
**타입**: `string`  
**설명**: 서버가 생성된 리전 코드입니다.

#### `zoneCode`
**설명**: 존 코드  
**타입**: `string`  
**설명**: 서버가 생성된 가용 영역 코드입니다.

**예제**:
```yaml
zoneCode: "KR-1"  # 한국 리전 1존
zoneCode: "KR-2"  # 한국 리전 2존
```

### 배치 및 GPU 정보

#### `placementGroupNo`
**설명**: 배치 그룹 번호  
**타입**: `string`  
**설명**: 서버가 속한 배치 그룹의 번호입니다.

#### `placementGroupName`
**설명**: 배치 그룹 이름  
**타입**: `string`  
**설명**: 서버가 속한 배치 그룹의 이름입니다.

**예제**:
```yaml
placementGroupName: "production-cluster"
placementGroupName: "development-nodes"
```

#### `fabricClusterPoolNo`
**설명**: GPU 클러스터 풀 번호  
**타입**: `string`  
**설명**: GPU 서버가 속한 클러스터 풀 번호입니다.

## 유효성 검증 규칙

### 필수 필드 검증
- `vpcNo`: 반드시 입력 필요 (omitempty 없음)
- `subnetNo`: 반드시 입력 필요 (omitempty 없음)
- 이미지 설정: `serverImageProductCode`, `memberServerImageInstanceNo`, `serverImageNo` 중 하나는 반드시 입력
- `credentials.secretRef.name`: Secret 이름 필수

### 길이 및 형식 검증
- `serverName`: 3-30자, 소문자/숫자/하이픈만 허용, 알파벳으로 시작
- `serverDescription`: 0-1000 바이트
- `serverCreateCount`: 1-10개
- `serverCreateStartNo`: 0-999

### 조건부 검증
- `associateWithPublicIp: true`시 `serverCreateCount`는 반드시 1개
- 베어메탈 서버 생성시 `raidTypeName` 필수
- GPU 서버 생성시 `fabricClusterPoolNo` 필요

## 예제 매니페스트

### 기본 예제
- **파일**: `config/samples/server_v1_ncloudserver.yaml`
- **용도**: 일반적인 서버 생성
- **특징**: 필수 필드와 주요 선택 필드 포함한 표준 설정

**주요 설정**:
```yaml
credentials:
  secretRef:
    name: ncloud-credentials
    namespace: ncloud-system
    # accessKeyIDKey: "access-key-id"        # 기본값 사용시 생략 가능
    # secretAccessKeyKey: "secret-access-key" # 기본값 사용시 생략 가능
```

### 최소 예제
- **파일**: `config/samples/server_v1_ncloudserver_minimal.yaml`  
- **용도**: 테스트용 서버 생성
- **특징**: 가장 최소한의 필수 필드만 포함

**최소 설정**:
```yaml
credentials:
  secretRef:
    name: ncloud-credentials
    # namespace 생략시 NCloudServer와 같은 네임스페이스 사용
    # accessKeyIDKey, secretAccessKeyKey 생략시 기본값 사용
```

### 고급 예제
- **파일**: `config/samples/server_v1_ncloudserver_advanced.yaml`
- **용도**: 프로덕션 환경용 서버 생성  
- **특징**: 모든 고급 기능 사용 예제

**고급 설정**:
```yaml
credentials:
  secretRef:
    name: ncloud-credentials
    namespace: ncloud-system
    # 커스텀 키 사용 가능
    accessKeyIDKey: "ncloud-access-key"
    secretAccessKeyKey: "ncloud-secret-key"

# 멀티 서버 생성
serverCreateCount: 3
serverCreateStartNo: 1

# 고급 네트워킹
networkInterfaceList:
  - networkInterfaceOrder: 0
    accessControlGroupNoList: ["acg-default"]
  - networkInterfaceOrder: 1
    subnetNo: "subnet-database"
    accessControlGroupNoList: ["acg-database"]

# GPU 및 스토리지 설정
fabricClusterPoolNo: "gpu-cluster-pool-12345"
isPreInstallGpuDriver: true
blockStorageMappingList:
  - order: 1
    blockStorageSize: "500"
    encrypted: true
```

## 관련 외부 리소스

### NCloud CLI 명령어
```bash
# 이미지 상품 목록 조회
ncloud vserver getServerImageProductList

# 서버 상품 목록 조회  
ncloud vserver getServerProductList --serverImageProductCode <이미지코드>

# VPC 목록 조회
ncloud vserver getVpcList

# 서브넷 목록 조회
ncloud vserver getSubnetList --vpcNo <vpc번호>

# ACG 목록 조회
ncloud vserver getAccessControlGroupList

# 로그인 키 목록 조회
ncloud vserver getLoginKeyList
```

### 참고 문서
- [NCloud VPC CLI 가이드](https://cli.ncloud-docs.com/docs/cli-vserver)
- [Server 이미지 상품 목록](https://console.ncloud.com/vserver/server/product)
- [VPC 가이드](https://guide.ncloud-docs.com/docs/vpc-overview)
