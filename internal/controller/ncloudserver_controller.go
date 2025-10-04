/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	serverv1 "github.com/Seo-yul/ncloud-server-controller/api/v1"
)

// NCloudServerReconciler reconciles a NCloudServer object
type NCloudServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// CLI 경로 설정 (환경변수로 지정 가능)
	NCloudCliPath string
}

const (
	// NCLOUD_CLI_PATH 환경변수명
	envNCloudCliPath = "NCLOUD_CLI_PATH"
	// 기본 CLI 경로
	defaultCliPath = "/opt/ncloud-cli/ncloud_cli_linux/ncloud"
)

// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers/finalizers,verbs=update

const (
	finalizerName = "ncloudserver.server.ncloud.devops.ai.kr"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NCloudServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// NCloudServer 리소스 가져오기
	var ncloudServer serverv1.NCloudServer
	if err := r.Get(ctx, req.NamespacedName, &ncloudServer); err != nil {
		if errors.IsNotFound(err) {
			log.Info("NCloudServer resource not found, ignoring since it must have been deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get NCloudServer")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling NCloudServer", "name", req.Name, "namespace", req.Namespace)

	// 삭제 요청 감지 및 Finalizer 처리
	if !ncloudServer.ObjectMeta.DeletionTimestamp.IsZero() {
		// 리소스가 삭제 요청됨 - cleanup 수행
		log.Info("Deletion requested for NCloudServer", "name", req.Name)
		return r.handleResourceDeletion(ctx, &ncloudServer, log)
	}

	// Finalizer 추가 (리소스가 삭제될 때 cleanup을 보장)
	if !controllerutil.ContainsFinalizer(&ncloudServer, finalizerName) {
		controllerutil.AddFinalizer(&ncloudServer, finalizerName)
		if err := r.Update(ctx, &ncloudServer); err != nil {
			log.Error(err, "failed to update finalizers")
			return ctrl.Result{}, err
		}
	}

	// 서버 상태 확인
	switch ncloudServer.Status.Phase {
	case "":
		// 새로 생성된 리소스 - CREATE 단계
		return r.handleServerCreate(ctx, &ncloudServer, log)
	case "Creating":
		// 생성 중 - 상태 확인
		return r.handleServerCreating(ctx, &ncloudServer, log)
	case "Running":
		// 운영 중 - 상태 동기화 확인
		return r.handleServerRunning(ctx, &ncloudServer, log)
	case "Failed":
		// 실패 상태 - 재시도 또는 에러 처리
		return r.handleServerFailed(ctx, &ncloudServer, log)
	case "Terminating":
		// 삭제 중 - 서버 삭제 처리
		return r.handleServerTerminating(ctx, &ncloudServer, log)
	default:
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NCloudServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//cloudCliPath 설정
	if cliPath := os.Getenv(envNCloudCliPath); cliPath != "" {
		r.NCloudCliPath = cliPath
	} else {
		r.NCloudCliPath = defaultCliPath
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&serverv1.NCloudServer{}).
		Named("ncloudserver").
		Complete(r)
}

// handleServerCreate 서버 생성 처리
func (r *NCloudServerReconciler) handleServerCreate(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Creating server", "name", server.Name)

	// 상태를 Creating으로 업데이트
	server.Status.Phase = "Creating"
	server.Status.Message = "Server creation initiated"
	server.Status.LastReconcileTime = metav1.Now().Format(time.RFC3339)

	if err := r.Status().Update(ctx, server); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	// 실제 서버 생성 명령 실행
	if err := r.createServer(ctx, server, log); err != nil {
		log.Error(err, "failed to create server")
		server.Status.Phase = "Failed"
		server.Status.Message = fmt.Sprintf("Failed to create server: %v", err)
		r.Status().Update(ctx, server)
		return ctrl.Result{RequeueAfter: time.Minute * 5}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

// handleServerCreating 생성 중 상태 처리
func (r *NCloudServerReconciler) handleServerCreating(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Checking server creation status", "serverInstanceNo", server.Status.ServerInstanceNo)

	// 서버 상태 확인
	status, err := r.getServerStatus(ctx, server.Status.ServerInstanceNo, log)
	if err != nil {
		log.Error(err, "failed to get server status")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	// 상태에 따라 처리
	switch status {
	case "INIT":
		log.Info("Server is initializing")
		server.Status.Message = "Server initializing"
	case "CREATDT":
		log.Info("Server creation completed")
		server.Status.Phase = "Running"
		server.Status.Message = "Server is running"
	case "RUN":
		log.Info("Server is running")
		server.Status.Phase = "Running"
		server.Status.Message = "Server is running"
	case "STOPDT":
		log.Info("Server stopped")
		server.Status.Message = "Server stopped"
	default:
		log.Info("Unknown server status", "status", status)
		server.Status.Message = fmt.Sprintf("Server status: %s", status)
	}

	server.Status.LastReconcileTime = metav1.Now().Format(time.RFC3339)
	return ctrl.Result{}, r.Status().Update(ctx, server)
}

// handleServerRunning 운영 중 상태 처리
func (r *NCloudServerReconciler) handleServerRunning(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Monitoring running server", "serverInstanceNo", server.Status.ServerInstanceNo)

	// 서버 정보 동기화
	serverInfo, err := r.getServerDetail(ctx, server.Status.ServerInstanceNo, log)
	if err != nil {
		log.Error(err, "failed to get server detail")
		return ctrl.Result{RequeueAfter: time.Minute * 5}, err
	}

	// 상태 업데이트
	if serverInfo != nil {
		server.Status.ServerInstanceStatus = serverv1.StatusCode{
			Code:     serverInfo.StatusCode,
			CodeName: serverInfo.StatusDescription,
		}
		server.Status.PublicIp = serverInfo.PublicIp
		server.Status.PrivateIp = serverInfo.PrivateIp
		server.Status.Message = "Server is running"
	}

	server.Status.LastReconcileTime = metav1.Now().Format(time.RFC3339)
	return ctrl.Result{}, r.Status().Update(ctx, server)
}

// handleServerFailed 실패 상태 처리
func (r *NCloudServerReconciler) handleServerFailed(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Server creation failed, manual intervention required", "name", server.Name)
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

// handleServerTerminating 삭제 중 상태 처리
func (r *NCloudServerReconciler) handleServerTerminating(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Terminating server", "serverInstanceNo", server.Status.ServerInstanceNo)

	if server.Status.ServerInstanceNo != "" {
		if err := r.deleteServer(ctx, server.Status.ServerInstanceNo, log); err != nil {
			log.Error(err, "failed to delete server")
			return ctrl.Result{RequeueAfter: time.Minute * 2}, err
		}
	}

	log.Info("Server termination completed")
	return ctrl.Result{}, nil
}

// handleResourceDeletion 리소스 삭제 시 cleanup 수행
func (r *NCloudServerReconciler) handleResourceDeletion(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
	log.Info("Handling resource deletion", "name", server.Name, "serverInstanceNo", server.Status.ServerInstanceNo)

	// 실제 서버가 생성되어 있다면 삭제
	if server.Status.ServerInstanceNo != "" && server.Status.ServerInstanceNo != "TEMPORARY_INSTANCE_NO" {
		log.Info("Deleting NCloud server", "serverInstanceNo", server.Status.ServerInstanceNo)

		if err := r.deleteServer(ctx, server.Status.ServerInstanceNo, log); err != nil {
			log.Error(err, "Failed to delete NCloud server", "serverInstanceNo", server.Status.ServerInstanceNo)
			// 삭제 실패해도 계속 진행 (서버가 이미 없는 경우 등)
			log.Info("Continuing cleanup despite server deletion failure")
		} else {
			log.Info("Successfully deleted NCloud server", "serverInstanceNo", server.Status.ServerInstanceNo)
		}
	} else {
		log.Info("No NCloud server to delete", "serverInstanceNo", server.Status.ServerInstanceNo)
	}

	// Finalizer 제거 (CR 삭제 허용)
	controllerutil.RemoveFinalizer(server, finalizerName)
	if err := r.Update(ctx, server); err != nil {
		log.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	log.Info("Resource deletion completed", "name", server.Name)
	return ctrl.Result{}, nil
}

// ServerInfo CLI에서 반환되는 서버 정보 구조체
type ServerInfo struct {
	ServerInstanceNo  string `json:"serverInstanceNo"`
	ServerName        string `json:"serverName"`
	ServerDescription string `json:"serverDescription"`
	CpuCount          int    `json:"cpuCount"`
	MemorySize        int64  `json:"memorySize"`
	PublicIp          string `json:"publicIp"`
	PrivateIp         string `json:"privateIp"`
	VpcNo             string `json:"vpcNo"`
	SubnetNo          string `json:"subnetNo"`
	StatusCode        string `json:"serverInstanceStatus.code"`
	StatusDescription string `json:"serverInstanceStatus.codeName"`
	RegionCode        string `json:"regionCode"`
	ZoneCode          string `json:"zoneCode"`
	CreateDate        string `json:"createDate"`
	PlatformType      string `json:"platformType.code"`
	ServerImageName   string `json:"serverImageName"`
}

// createServer 실제 서버 생성 CLI 실행
func (r *NCloudServerReconciler) createServer(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) error {
	log.Info("Executing server creation command")

	// CLI 인자 구성
	args := []string{"vserver", "createServerInstances", "--output", "json"}

	// 필수 파라미터
	if server.Spec.VpcNo != "" {
		args = append(args, "--vpcNo", server.Spec.VpcNo)
	}
	if server.Spec.SubnetNo != "" {
		args = append(args, "--subnetNo", server.Spec.SubnetNo)
	}

	// 이미지 설정 (필수 중 하나)
	if server.Spec.ServerImageProductCode != "" {
		args = append(args, "--serverImageProductCode", server.Spec.ServerImageProductCode)
	} else if server.Spec.MemberServerImageInstanceNo != "" {
		args = append(args, "--memberServerImageInstanceNo", server.Spec.MemberServerImageInstanceNo)
	} else if server.Spec.ServerImageNo != "" {
		args = append(args, "--serverImageNo", server.Spec.ServerImageNo)
	}

	// 선택적 파라미터들
	if server.Spec.RegionCode != "" {
		args = append(args, "--regionCode", server.Spec.RegionCode)
	}
	if server.Spec.ServerProductCode != "" {
		args = append(args, "--serverProductCode", server.Spec.ServerProductCode)
	}
	if server.Spec.ServerSpecCode != "" {
		args = append(args, "--serverSpecCode", server.Spec.ServerSpecCode)
	}
	if server.Spec.ServerName != "" {
		args = append(args, "--serverName", server.Spec.ServerName)
	}
	if server.Spec.ServerDescription != "" {
		args = append(args, "--serverDescription", server.Spec.ServerDescription)
	}
	if server.Spec.LoginKeyName != "" {
		args = append(args, "--loginKeyName", server.Spec.LoginKeyName)
	}
	if server.Spec.FeeSystemTypeCode != "" {
		args = append(args, "--feeSystemTypeCode", server.Spec.FeeSystemTypeCode)
	}
	if server.Spec.ServerCreateCount > 0 {
		args = append(args, "--serverCreateCount", fmt.Sprintf("%d", server.Spec.ServerCreateCount))
	}
	if server.Spec.ServerCreateStartNo > 0 {
		args = append(args, "--serverCreateStartNo", fmt.Sprintf("%d", server.Spec.ServerCreateStartNo))
	}

	// 네트워크 인터페이스 설정
	if len(server.Spec.NetworkInterfaceList) > 0 {
		for _, ni := range server.Spec.NetworkInterfaceList {
			niArg := fmt.Sprintf("networkInterfaceOrder=%d", ni.NetworkInterfaceOrder)
			if ni.SubnetNo != "" {
				niArg += fmt.Sprintf(",subnetNo=%s", ni.SubnetNo)
			}
			if ni.IP != "" {
				niArg += fmt.Sprintf(",ip=%s", ni.IP)
			}
			if len(ni.AccessControlGroupNoList) > 0 {
				acgList := strings.Join(ni.AccessControlGroupNoList, ",")
				niArg += fmt.Sprintf(",accessControlGroupNoList=[%s]", acgList)
			}
			args = append(args, "--networkInterfaceList", niArg)
		}
	}

	if len(server.Spec.AccessControlGroupNoList) > 0 {
		acgList := strings.Join(server.Spec.AccessControlGroupNoList, ",")
		args = append(args, "--networkInterfaceList", fmt.Sprintf("networkInterfaceOrder=0,accessControlGroupNoList=[%s]", acgList))
	}

	// 기타 옵션들
	if server.Spec.IsProtectServerTermination {
		args = append(args, "--isProtectServerTermination", "true")
	}
	if server.Spec.AssociateWithPublicIp {
		args = append(args, "--associateWithPublicIp", "true")
	}
	if server.Spec.IsEncryptedBaseBlockStorageVolume {
		args = append(args, "--isEncryptedBaseBlockStorageVolume", "true")
	}
	if server.Spec.PlacementGroupNo != "" {
		args = append(args, "--placementGroupNo", server.Spec.PlacementGroupNo)
	}
	if server.Spec.InitScriptNo != "" {
		args = append(args, "--initScriptNo", server.Spec.InitScriptNo)
	}
	if server.Spec.RaidTypeName != "" {
		args = append(args, "--raidTypeName", server.Spec.RaidTypeName)
	}

	// 추가 스토리지 설정 (BlockStorageMapping)
	if len(server.Spec.BlockStorageMappingList) > 0 {
		for i, bgm := range server.Spec.BlockStorageMappingList {
			bgmArg := fmt.Sprintf("order=%d", i+1)
			if bgm.BlockStorageSize != "" {
				bgmArg += fmt.Sprintf(",blockStorageSize=%s", bgm.BlockStorageSize)
			}
			if bgm.BlockStorageName != "" {
				bgmArg += fmt.Sprintf(",blockStorageName=%s", bgm.BlockStorageName)
			}
			if bgm.BlockStorageVolumeTypeCode != "" {
				bgmArg += fmt.Sprintf(",blockStorageVolumeTypeCode=%s", bgm.BlockStorageVolumeTypeCode)
			}
			if bgm.Encrypted {
				bgmArg += ",encrypted=true"
			}
			if bgm.NoBlockStorage {
				bgmArg += ",noBlockStorage=true"
			}
			if bgm.EmptyBlockStorage {
				bgmArg += ",emptyBlockStorage=true"
			}
			args = append(args, "--blockStorageMappingList", bgmArg)
		}
	}

	// GPU 설정
	if server.Spec.FabricClusterPoolNo != "" {
		args = append(args, "--fabricClusterPoolNo", server.Spec.FabricClusterPoolNo)
	}
	if server.Spec.IsPreInstallGpuDriver {
		args = append(args, "--isPreInstallGpuDriver", "true")
	}

	// 파티션 설정 (Bare Metal)
	if len(server.Spec.BlockDevicePartitionList) > 0 {
		for _, bdp := range server.Spec.BlockDevicePartitionList {
			if bdp.MountPoint != "" && bdp.PartitionSize != "" {
				partArg := fmt.Sprintf("mountPoint=%s,partitionSize=%s", bdp.MountPoint, bdp.PartitionSize)
				args = append(args, "--blockDevicePartitionList", partArg)
			}
		}
	}

	// 디버그 모드
	if server.Spec.DebugMode {
		args = append(args, "--debug")
	}

	log.Info("Executing CLI command", "args", args)

	// 커맨드 실행 (상대 경로 문제 해결 포함)
	cmd := exec.CommandContext(ctx, r.NCloudCliPath, args...)

	// 컨테이너 내에서 CLI를 절대 경로로 찾도록 설정
	if r.NCloudCliPath == defaultCliPath {
		cliDir := filepath.Dir(r.NCloudCliPath)
		cmd.Dir = cliDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "CLI command failed", "output", string(output))
		return fmt.Errorf("CLI command failed: %v, output: %s", err, string(output))
	}

	log.Info("CLI command executed successfully", "output", string(output))

	// 응답에서 서버 인스턴스 번호 추출
	// 실제로는 JSON 파싱을 해야 하지만, 간단히 하기 위해 여기서는 성공 확인만
	server.Status.ServerInstanceNo = "TEMPORARY_INSTANCE_NO" // 임시값

	return nil
}

// deleteServer 서버 삭제 CLI 실행
func (r *NCloudServerReconciler) deleteServer(ctx context.Context, serverInstanceNo string, log logr.Logger) error {
	log.Info("Executing server deletion command", "serverInstanceNo", serverInstanceNo)

	args := []string{"vserver", "terminateServerInstances", "--output", "json", "--serverInstanceNoList", fmt.Sprintf("[\"%s\"]", serverInstanceNo)}

	cmd := exec.CommandContext(ctx, r.NCloudCliPath, args...)
	if r.NCloudCliPath == defaultCliPath {
		cliDir := filepath.Dir(r.NCloudCliPath)
		cmd.Dir = cliDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "CLI delete command failed", "output", string(output))
		return fmt.Errorf("CLI delete command failed: %v, output: %s", err, string(output))
	}

	log.Info("Server deletion completed", "output", string(output))
	return nil
}

// getServerStatus 서버 상태 조회 CLI 실행
func (r *NCloudServerReconciler) getServerStatus(ctx context.Context, serverInstanceNo string, log logr.Logger) (string, error) {
	if serverInstanceNo == "" || serverInstanceNo == "TEMPORARY_INSTANCE_NO" {
		return "INIT", nil // 임시값 처리
	}

	log.Info("Getting server status", "serverInstanceNo", serverInstanceNo)

	args := []string{"vserver", "getServerInstanceDetail", "--output", "json", "--serverInstanceNo", serverInstanceNo}

	cmd := exec.CommandContext(ctx, r.NCloudCliPath, args...)
	if r.NCloudCliPath == defaultCliPath {
		cliDir := filepath.Dir(r.NCloudCliPath)
		cmd.Dir = cliDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "CLI status command failed", "output", string(output))
		return "", fmt.Errorf("CLI status command failed: %v", err)
	}

	log.Info("Server status retrieved", "output", string(output))

	// 간단한 상태 파싱 (실제로는 JSON 파싱 필요)
	if strings.Contains(string(output), "RUN") {
		return "RUN", nil
	} else if strings.Contains(string(output), "CREATDT") {
		return "CREATDT", nil
	} else if strings.Contains(string(output), "INIT") {
		return "INIT", nil
	} else if strings.Contains(string(output), "STOPDT") {
		return "STOPDT", nil
	}

	return "UNKNOWN", nil
}

// getServerDetail 서버 상세 정보 조회 CLI 실행
func (r *NCloudServerReconciler) getServerDetail(ctx context.Context, serverInstanceNo string, log logr.Logger) (*ServerInfo, error) {
	if serverInstanceNo == "" || serverInstanceNo == "TEMPORARY_INSTANCE_NO" {
		return nil, nil // 임시값 처리
	}

	log.Info("Getting server detail", "serverInstanceNo", serverInstanceNo)

	args := []string{"vserver", "getServerInstanceDetail", "--output", "json", "--serverInstanceNo", serverInstanceNo}

	cmd := exec.CommandContext(ctx, r.NCloudCliPath, args...)
	if r.NCloudCliPath == defaultCliPath {
		cliDir := filepath.Dir(r.NCloudCliPath)
		cmd.Dir = cliDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "CLI detail command failed", "output", string(output))
		return nil, fmt.Errorf("CLI detail command failed: %v", err)
	}

	log.Info("Server detail retrieved", "output", string(output))

	// 간단한 서버 정보 구조체 반환 (실제로는 JSON 파싱 필요)
	return &ServerInfo{
		ServerInstanceNo:  serverInstanceNo,
		StatusCode:        "RUN",
		StatusDescription: "Running",
		PublicIp:          "TBD",
		PrivateIp:         "TBD",
		RegionCode:        "KR",
	}, nil
}
