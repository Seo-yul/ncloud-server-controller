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
	"encoding/json"
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
	defaultCliPath = "/app/ncloud-cli/ncloud_cli_linux/ncloud"

	// 서버 상태 상수
	serverPhaseCreating    = "Creating"
	serverPhaseRunning     = "Running"
	serverPhaseFailed      = "Failed"
	serverPhaseTerminating = "Terminating"

	// 서버 상태 메시지 상수
	serverStatusRunning = "Server is running"
	serverStatusInit    = "INIT"

	// 임시 인스턴스 번호
	tempInstanceNo = "TEMPORARY_INSTANCE_NO"
)

// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=server.ncloud.devops.ai.kr,resources=ncloudservers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

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
	if !ncloudServer.DeletionTimestamp.IsZero() {
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
	case serverPhaseCreating:
		// 생성 중 - 상태 확인
		return r.handleServerCreating(ctx, &ncloudServer, log)
	case serverPhaseRunning:
		// 운영 중 - 상태 동기화 확인
		return r.handleServerRunning(ctx, &ncloudServer, log)
	case serverPhaseFailed:
		// 실패 상태 - 재시도 또는 에러 처리
		return r.handleServerFailed(ctx, &ncloudServer, log)
	case serverPhaseTerminating:
		// 삭제 중 - 서버 삭제 처리
		return r.handleServerTerminating(ctx, &ncloudServer, log)
	default:
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NCloudServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// cloudCliPath 설정
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

	if err := r.updateStatus(ctx, server, log); err != nil {
		return ctrl.Result{}, err
	}

	// 실제 서버 생성 명령 실행
	if err := r.createServer(ctx, server, log); err != nil {
		log.Error(err, "failed to create server")
		server.Status.Phase = "Failed"
		server.Status.Message = fmt.Sprintf("Failed to create server: %v", err)
		if updateErr := r.Status().Update(ctx, server); updateErr != nil {
			log.Error(updateErr, "failed to update status after server creation failure")
			return ctrl.Result{}, updateErr
		}
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
	case serverStatusInit:
		log.Info("Server is initializing")
		server.Status.Message = "Server initializing"
	case "CREATDT":
		log.Info("Server creation completed")
		server.Status.Phase = "Running"
		server.Status.Message = serverStatusRunning
	case "RUN":
		log.Info("Server is running")
		server.Status.Phase = "Running"
		server.Status.Message = serverStatusRunning
	case "STOPDT":
		log.Info("Server stopped")
		server.Status.Message = "Server stopped"
	default:
		log.Info("Unknown server status", "status", status)
		server.Status.Message = fmt.Sprintf("Server status: %s", status)
	}

	server.Status.LastReconcileTime = metav1.Now().Format(time.RFC3339)
	if err := r.updateStatus(ctx, server, log); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
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
		server.Status.Message = serverStatusRunning
	}

	server.Status.LastReconcileTime = metav1.Now().Format(time.RFC3339)
	if err := r.updateStatus(ctx, server, log); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleServerFailed 실패 상태 처리
func (r *NCloudServerReconciler) handleServerFailed(_ context.Context, server *serverv1.NCloudServer, log logr.Logger) (ctrl.Result, error) {
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
	if server.Status.ServerInstanceNo != "" && server.Status.ServerInstanceNo != tempInstanceNo {
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

// updateStatus 안전한 Status 업데이트 헬퍼 함수
func (r *NCloudServerReconciler) updateStatus(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) error {
	if err := r.Status().Update(ctx, server); err != nil {
		log.Error(err, "failed to update status")
		return err
	}
	return nil
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

// ServerStatusResponse 서버 상태 조회 응답 구조체
type ServerStatusResponse struct {
	GetServerInstanceListResponse struct {
		RequestID          string `json:"requestId"`
		ReturnCode         string `json:"returnCode"`
		ReturnMessage      string `json:"returnMessage"`
		ServerInstanceList []struct {
			ServerInstanceNo         string `json:"serverInstanceNo"`
			ServerInstanceStatus     string `json:"serverInstanceStatus.code"`
			ServerInstanceStatusName string `json:"serverInstanceStatusName"`
		} `json:"serverInstanceList"`
	} `json:"getServerInstanceListResponse"`
}

// ServerInstance 서버 인스턴스 정보 구조체
type ServerInstance struct {
	ServerInstanceNo           string   `json:"serverInstanceNo"`
	ServerName                 string   `json:"serverName"`
	ServerInstanceStatusName   string   `json:"serverInstanceStatusName"`
	ServerInstanceOperation    string   `json:"serverInstanceOperation"`
	ServerInstanceStatus       string   `json:"serverInstanceStatus"`
	PlatformType               string   `json:"platformType"`
	LoginKeyName               string   `json:"loginKeyName"`
	IsFeeChargingMonitoring    bool     `json:"isFeeChargingMonitoring"`
	PublicIp                   string   `json:"publicIp"`
	PrivateIp                  string   `json:"privateIp"`
	ServerImageName            string   `json:"serverImageName"`
	ServerInstanceType         string   `json:"serverInstanceType"`
	RegionCode                 string   `json:"regionCode"`
	ZoneCode                   string   `json:"zoneCode"`
	VpcNo                      string   `json:"vpcNo"`
	SubnetNo                   string   `json:"subnetNo"`
	NetworkInterfaceNoList     []string `json:"networkInterfaceNoList"`
	PlacementGroupName         string   `json:"placementGroupName"`
	FabricClusterPoolNo        string   `json:"fabricClusterPoolNo"`
	IsProtectServerTermination bool     `json:"isProtectServerTermination"`
}

// CreateServerResponse 서버 생성 응답 구조체
type CreateServerResponse struct {
	CreateServerInstancesResponse struct {
		RequestID          string           `json:"requestId"`
		ReturnCode         string           `json:"returnCode"`
		ReturnMessage      string           `json:"returnMessage"`
		ServerInstanceList []ServerInstance `json:"serverInstanceList"`
	} `json:"createServerInstancesResponse"`
}

// buildServerCreationArgs CLI 인자 구성
func (r *NCloudServerReconciler) buildServerCreationArgs(server *serverv1.NCloudServer) []string {
	args := []string{"vserver", "createServerInstances", "--output", "json"}

	// 필수 파라미터
	args = r.addRequiredServerArgs(args, server)

	// 이미지 설정
	args = r.addImageServerArgs(args, server)

	// 선택적 파라미터들
	args = r.addOptionalServerArgs(args, server)

	// 네트워크 인터페이스 설정
	args = r.addNetworkInterfaceServerArgs(args, server)

	// 블록 스토리지 매핑
	args = r.addBlockStorageServerArgs(args, server)

	// 파티션 설정
	args = r.addPartitionServerArgs(args, server)

	return args
}

// addRequiredServerArgs 필수 파라미터 추가
func (r *NCloudServerReconciler) addRequiredServerArgs(args []string, server *serverv1.NCloudServer) []string {
	if server.Spec.VpcNo != "" {
		args = append(args, "--vpcNo", server.Spec.VpcNo)
	}
	if server.Spec.SubnetNo != "" {
		args = append(args, "--subnetNo", server.Spec.SubnetNo)
	}
	return args
}

// addImageServerArgs 이미지 설정 추가
func (r *NCloudServerReconciler) addImageServerArgs(args []string, server *serverv1.NCloudServer) []string {
	if server.Spec.ServerImageProductCode != "" {
		args = append(args, "--serverImageProductCode", server.Spec.ServerImageProductCode)
	} else if server.Spec.MemberServerImageInstanceNo != "" {
		args = append(args, "--memberServerImageInstanceNo", server.Spec.MemberServerImageInstanceNo)
	} else if server.Spec.ServerImageNo != "" {
		args = append(args, "--serverImageNo", server.Spec.ServerImageNo)
	}
	return args
}

// addOptionalServerArgs 선택적 파라미터 추가
func (r *NCloudServerReconciler) addOptionalServerArgs(args []string, server *serverv1.NCloudServer) []string {
	optionalParams := map[string]string{
		"regionCode":          server.Spec.RegionCode,
		"serverProductCode":   server.Spec.ServerProductCode,
		"serverSpecCode":      server.Spec.ServerSpecCode,
		"serverName":          server.Spec.ServerName,
		"serverDescription":   server.Spec.ServerDescription,
		"loginKeyName":        server.Spec.LoginKeyName,
		"feeSystemTypeCode":   server.Spec.FeeSystemTypeCode,
		"placementGroupNo":    server.Spec.PlacementGroupNo,
		"raidTypeName":        server.Spec.RaidTypeName,
		"initScriptNo":        server.Spec.InitScriptNo,
		"fabricClusterPoolNo": server.Spec.FabricClusterPoolNo,
	}

	for param, value := range optionalParams {
		if value != "" {
			args = append(args, "--"+param, value)
		}
	}

	// 숫자 파라미터들
	if server.Spec.ServerCreateCount > 0 {
		args = append(args, "--serverCreateCount", fmt.Sprintf("%d", server.Spec.ServerCreateCount))
	}
	if server.Spec.ServerCreateStartNo > 0 {
		args = append(args, "--serverCreateStartNo", fmt.Sprintf("%d", server.Spec.ServerCreateStartNo))
	}

	// Boolean 파라미터들
	if server.Spec.IsProtectServerTermination {
		args = append(args, "--isProtectServerTermination", "true")
	}
	if server.Spec.AssociateWithPublicIp {
		args = append(args, "--associateWithPublicIp", "true")
	}
	if server.Spec.IsEncryptedBaseBlockStorageVolume {
		args = append(args, "--isEncryptedBaseBlockStorageVolume", "true")
	}

	return args
}

// addNetworkInterfaceServerArgs 네트워크 인터페이스 설정 추가
func (r *NCloudServerReconciler) addNetworkInterfaceServerArgs(args []string, server *serverv1.NCloudServer) []string {
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

	// 접근 제어 그룹 (기본 네트워크 인터페이스)
	if len(server.Spec.AccessControlGroupNoList) > 0 {
		acgList := strings.Join(server.Spec.AccessControlGroupNoList, ",")
		args = append(args, "--networkInterfaceList", fmt.Sprintf("networkInterfaceOrder=0,accessControlGroupNoList=[%s]", acgList))
	}

	return args
}

// addBlockStorageServerArgs 블록 스토리지 매핑 추가
func (r *NCloudServerReconciler) addBlockStorageServerArgs(args []string, server *serverv1.NCloudServer) []string {
	for _, mapping := range server.Spec.BlockStorageMappingList {
		if mapping.BlockStorageName != "" {
			args = append(args, "--blockStorageName", mapping.BlockStorageName)
		}
		if mapping.BlockStorageSize != "" {
			args = append(args, "--blockStorageSize", mapping.BlockStorageSize)
		}
		if mapping.BlockStorageVolumeTypeCode != "" {
			args = append(args, "--blockStorageVolumeTypeCode", mapping.BlockStorageVolumeTypeCode)
		}
		if mapping.EmptyBlockStorage {
			args = append(args, "--emptyBlockStorage")
		}
		if mapping.Encrypted {
			args = append(args, "--encrypted")
		}
		if mapping.NoBlockStorage {
			args = append(args, "--noBlockStorage")
		}
	}
	return args
}

// addPartitionServerArgs 파티션 설정 추가
func (r *NCloudServerReconciler) addPartitionServerArgs(args []string, server *serverv1.NCloudServer) []string {
	for _, partition := range server.Spec.BlockDevicePartitionList {
		if partition.MountPoint != "" {
			args = append(args, "--mountPoint", partition.MountPoint)
		}
		if partition.PartitionSize != "" {
			args = append(args, "--partitionSize", partition.PartitionSize)
		}
	}
	return args
}

// executeServerCreationCLI CLI 명령 실행
func (r *NCloudServerReconciler) executeServerCreationCLI(args []string, log logr.Logger) ([]byte, error) {
	log.Info("Executing CLI command", "args", args)

	cmd := exec.Command(r.NCloudCliPath, args...)
	if r.NCloudCliPath == defaultCliPath {
		cliDir := filepath.Dir(r.NCloudCliPath)
		cmd.Dir = cliDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, "Failed to execute CLI command", "output", string(output))
		return nil, fmt.Errorf("failed to execute CLI command: %w", err)
	}

	log.Info("CLI command executed successfully", "output", string(output))
	return output, nil
}

// parseServerCreationResponse 서버 생성 응답 파싱
func (r *NCloudServerReconciler) parseServerCreationResponse(output []byte, log logr.Logger) (*ServerInstance, error) {
	var response CreateServerResponse
	if err := json.Unmarshal(output, &response); err != nil {
		log.Error(err, "Failed to parse CLI response", "output", string(output))
		return nil, fmt.Errorf("failed to parse CLI response: %w", err)
	}

	// 응답 검증
	if response.CreateServerInstancesResponse.ReturnCode != "0" {
		log.Error(nil, "CLI command failed", "returnCode", response.CreateServerInstancesResponse.ReturnCode, "returnMessage", response.CreateServerInstancesResponse.ReturnMessage)
		return nil, fmt.Errorf("CLI command failed: %s", response.CreateServerInstancesResponse.ReturnMessage)
	}

	// 서버 인스턴스 정보 추출
	if len(response.CreateServerInstancesResponse.ServerInstanceList) == 0 {
		log.Error(nil, "No server instances returned from CLI")
		return nil, fmt.Errorf("no server instances returned from CLI")
	}

	return &response.CreateServerInstancesResponse.ServerInstanceList[0], nil
}

// updateServerCreationStatus 서버 생성 상태 업데이트
func (r *NCloudServerReconciler) updateServerCreationStatus(ctx context.Context, server *serverv1.NCloudServer, serverInstance *ServerInstance, log logr.Logger) error {
	server.Status.Phase = serverPhaseCreating
	server.Status.Message = "Server is being created"
	server.Status.ServerInstanceNo = serverInstance.ServerInstanceNo
	server.Status.ServerName = serverInstance.ServerName
	server.Status.ServerInstanceStatusName = serverInstance.ServerInstanceStatusName
	server.Status.ServerInstanceOperation = serverv1.StatusCode{Code: serverInstance.ServerInstanceOperation}
	server.Status.ServerInstanceStatus = serverv1.StatusCode{Code: serverInstance.ServerInstanceStatus}
	server.Status.PlatformType = serverv1.StatusCode{Code: serverInstance.PlatformType}
	server.Status.LoginKeyName = serverInstance.LoginKeyName
	server.Status.PublicIp = serverInstance.PublicIp
	server.Status.PrivateIp = serverInstance.PrivateIp
	server.Status.ServerImageName = serverInstance.ServerImageName
	server.Status.ServerInstanceType = serverv1.StatusCode{Code: serverInstance.ServerInstanceType}
	server.Status.RegionCode = serverInstance.RegionCode
	server.Status.ZoneCode = serverInstance.ZoneCode
	server.Status.VpcNo = serverInstance.VpcNo
	server.Status.SubnetNo = serverInstance.SubnetNo
	server.Status.NetworkInterfaceNoList = serverInstance.NetworkInterfaceNoList
	server.Status.PlacementGroupName = serverInstance.PlacementGroupName
	server.Status.FabricClusterPoolNo = serverInstance.FabricClusterPoolNo
	server.Status.IsProtectServerTermination = serverInstance.IsProtectServerTermination

	if err := r.Status().Update(ctx, server); err != nil {
		log.Error(err, "Failed to update server status")
		return fmt.Errorf("failed to update server status: %w", err)
	}

	log.Info("Server creation initiated", "serverInstanceNo", serverInstance.ServerInstanceNo)
	return nil
}

// createServer 실제 서버 생성 CLI 실행
// createServer 실제 서버 생성 CLI 실행
func (r *NCloudServerReconciler) createServer(ctx context.Context, server *serverv1.NCloudServer, log logr.Logger) error {
	log.Info("Executing server creation command")

	// CLI 인자 구성
	args := r.buildServerCreationArgs(server)

	// CLI 실행
	output, err := r.executeServerCreationCLI(args, log)
	if err != nil {
		return err
	}

	// 응답 파싱 및 검증
	serverInstance, err := r.parseServerCreationResponse(output, log)
	if err != nil {
		return err
	}

	// 상태 업데이트
	return r.updateServerCreationStatus(ctx, server, serverInstance, log)
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

	// 구조화된 JSON 파싱
	var response ServerStatusResponse
	if err := json.Unmarshal(output, &response); err != nil {
		log.Error(err, "Failed to parse server status response", "output", string(output))
		// JSON 파싱 실패 시 기존 방식으로 폴백
		return r.parseServerStatusFallback(output, log)
	}

	// 응답 검증
	if response.GetServerInstanceListResponse.ReturnCode != "0" {
		log.Error(nil, "Server status command failed", "returnCode", response.GetServerInstanceListResponse.ReturnCode, "returnMessage", response.GetServerInstanceListResponse.ReturnMessage)
		return "", fmt.Errorf("server status command failed: %s", response.GetServerInstanceListResponse.ReturnMessage)
	}

	// 서버 인스턴스 상태 추출
	if len(response.GetServerInstanceListResponse.ServerInstanceList) == 0 {
		log.Error(nil, "No server instances found in status response")
		return "UNKNOWN", nil
	}

	status := response.GetServerInstanceListResponse.ServerInstanceList[0].ServerInstanceStatus
	log.Info("Server status parsed", "status", status)
	return status, nil
}

// parseServerStatusFallback JSON 파싱 실패 시 폴백 함수
func (r *NCloudServerReconciler) parseServerStatusFallback(output []byte, log logr.Logger) (string, error) {
	log.Info("Using fallback parsing for server status")

	// 기존 문자열 매칭 방식
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
