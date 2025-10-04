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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NCloudServerSpec defines the desired state of NCloudServer.
type NCloudServerSpec struct {
	// NCloud authentication configuration
	Credentials *CredentialsSpec `json:"credentials,omitempty"`

	// Region code where the server will be created
	RegionCode string `json:"regionCode,omitempty"`

	// Server image configuration
	// Exactly one of these must be specified
	ServerImageProductCode      string `json:"serverImageProductCode,omitempty"`      // Image product code (new image)
	MemberServerImageInstanceNo string `json:"memberServerImageInstanceNo,omitempty"` // Custom image instance number
	ServerImageNo               string `json:"serverImageNo,omitempty"`               // Image number (alternative)

	// VPC configuration (Required for VPC environment)
	VpcNo    string `json:"vpcNo"`    // VPC number
	SubnetNo string `json:"subnetNo"` // Subnet number for default network interface

	// Server specification
	ServerProductCode string `json:"serverProductCode,omitempty"` // Product code
	ServerSpecCode    string `json:"serverSpecCode,omitempty"`    // Alternative spec code

	// Storage configuration
	IsEncryptedBaseBlockStorageVolume bool `json:"isEncryptedBaseBlockStorageVolume,omitempty"` // Encrypt base storage

	// Billing configuration
	FeeSystemTypeCode string `json:"feeSystemTypeCode,omitempty"` // MTRAT (hourly) or FXSUM (monthly)

	// Server naming and description
	ServerName        string `json:"serverName,omitempty"`        // Server name (3-30 chars)
	ServerDescription string `json:"serverDescription,omitempty"` // Server description

	// Scaling configuration
	ServerCreateCount   int `json:"serverCreateCount,omitempty"`   // Number of servers to create
	ServerCreateStartNo int `json:"serverCreateStartNo,omitempty"` // Starting sequence number

	// Security configuration
	LoginKeyName               string                          `json:"loginKeyName,omitempty"`               // SSH key name
	NetworkInterfaceList       []NetworkInterfaceConfiguration `json:"networkInterfaceList,omitempty"`       // Network interfaces
	AccessControlGroupNoList   []string                        `json:"accessControlGroupNoList,omitempty"`   // ACG list for default interface
	IsProtectServerTermination bool                            `json:"isProtectServerTermination,omitempty"` // Termination protection

	// Advanced configuration
	PlacementGroupNo string `json:"placementGroupNo,omitempty"` // Physical placement group
	InitScriptNo     string `json:"initScriptNo,omitempty"`     // Initialization script

	// Public IP configuration
	AssociateWithPublicIp bool `json:"associateWithPublicIp,omitempty"` // Auto-assign public IP

	// Bare Metal specific
	RaidTypeName string `json:"raidTypeName,omitempty"` // RAID configuration

	// Partition configuration for Bare Metal
	BlockDevicePartitionList []BlockDevicePartition `json:"blockDevicePartitionList,omitempty"`

	// Additional storage configuration
	BlockStorageMappingList []BlockStorageMapping `json:"blockStorageMappingList,omitempty"`

	// GPU configuration
	FabricClusterPoolNo   string `json:"fabricClusterPoolNo,omitempty"`   // GPU cluster pool
	IsPreInstallGpuDriver bool   `json:"isPreInstallGpuDriver,omitempty"` // Pre-install GPU driver

	// Kubernetes specific annotations for debugging
	DebugMode bool `json:"debugMode,omitempty"` // Enable debug logging
}

// CredentialsSpec defines NCloud authentication configuration
type CredentialsSpec struct {
	// Secret reference containing NCloud credentials
	SecretRef *SecretReference `json:"secretRef,omitempty"`
}

// SecretReference specifies a secret namespace and name
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace of the secret (defaults to same namespace as NCloudServer)
	Namespace string `json:"namespace,omitempty"`

	// Key within the secret containing access key ID (defaults to 'access-key-id')
	AccessKeyIDKey string `json:"accessKeyIDKey,omitempty"`

	// Key within the secret containing secret訪問키 (defaults to 'secret-access-key')
	SecretAccessKeyKey string `json:"secretAccessKeyKey,omitempty"`
}

// NetworkInterfaceConfiguration defines network interface settings
type NetworkInterfaceConfiguration struct {
	NetworkInterfaceOrder    int      `json:"networkInterfaceOrder"`              // Order (0-2, 0 is default)
	NetworkInterfaceNo       string   `json:"networkInterfaceNo,omitempty"`       // Existing interface number
	SubnetNo                 string   `json:"subnetNo,omitempty"`                 // Subnet for new interface
	IP                       string   `json:"ip,omitempty"`                       // Specific IP address
	AccessControlGroupNoList []string `json:"accessControlGroupNoList,omitempty"` // ACG list for this interface
}

// BlockDevicePartition defines partition settings for Bare Metal
type BlockDevicePartition struct {
	MountPoint    string `json:"mountPoint"`    // Mount point (e.g., "/")
	PartitionSize string `json:"partitionSize"` // Size in GiB
}

// BlockStorageMapping defines additional storage configuration
type BlockStorageMapping struct {
	Order                      int    `json:"order,omitempty"`                      // Order
	SnapshotInstanceNo         string `json:"snapshotInstanceNo,omitempty"`         // Snapshot instance number
	BlockStorageSize           string `json:"blockStorageSize,omitempty"`           // Storage size
	BlockStorageName           string `json:"blockStorageName,omitempty"`           // Storage name
	BlockStorageVolumeTypeCode string `json:"blockStorageVolumeTypeCode,omitempty"` // Volume type
	Encrypted                  bool   `json:"encrypted,omitempty"`                  // Encryption
	NoBlockStorage             bool   `json:"noBlockStorage,omitempty"`             // Exclude storage
	EmptyBlockStorage          bool   `json:"emptyBlockStorage,omitempty"`          // Create new storage
}

// NCloudServerStatus defines the observed state of NCloudServer.
type NCloudServerStatus struct {
	// Server instance information from NCloud API
	ServerInstanceNo  string `json:"serverInstanceNo,omitempty"`  // NCloud server instance number
	ServerName        string `json:"serverName,omitempty"`        // Actual server name
	ServerDescription string `json:"serverDescription,omitempty"` // Actual server description

	// Server resources
	CpuCount   int   `json:"cpuCount,omitempty"`   // CPU count
	MemorySize int64 `json:"memorySize,omitempty"` // Memory size in bytes

	// Networking information
	PublicIp  string `json:"publicIp,omitempty"`  // Public IP address
	PrivateIp string `json:"privateIp,omitempty"` // Private IP address
	VpcNo     string `json:"vpcNo,omitempty"`     // VPC number
	SubnetNo  string `json:"subnetNo,omitempty"`  // Subnet number

	// Status information
	ServerInstanceStatus     StatusCode `json:"serverInstanceStatus,omitempty"`     // Current status
	ServerInstanceOperation  StatusCode `json:"serverInstanceOperation,omitempty"`  // Current operation
	ServerInstanceStatusName string     `json:"serverInstanceStatusName,omitempty"` // Status name

	// Image and specification information
	ServerImageProductCode string     `json:"serverImageProductCode,omitempty"` // Image product code
	ServerProductCode      string     `json:"serverProductCode,omitempty"`      // Server product code
	ServerImageName        string     `json:"serverImageName,omitempty"`        // Image name
	ServerInstanceType     StatusCode `json:"serverInstanceType,omitempty"`     // Instance type

	// Platform and hypervisor information
	PlatformType   StatusCode `json:"platformType,omitempty"`   // Platform (Linux, Windows)
	HypervisorType StatusCode `json:"hypervisorType,omitempty"` // Hypervisor type

	// Time information
	CreateDate string `json:"createDate,omitempty"` // Creation timestamp
	Uptime     string `json:"uptime,omitempty"`     // Uptime timestamp

	// Security and access information
	LoginKeyName               string   `json:"loginKeyName,omitempty"`               // SSH key used
	NetworkInterfaceNoList     []string `json:"networkInterfaceNoList,omitempty"`     // Network interfaces
	IsProtectServerTermination bool     `json:"isProtectServerTermination,omitempty"` // Protection status

	// Region and zone information
	RegionCode string `json:"regionCode,omitempty"` // Region code
	ZoneCode   string `json:"zoneCode,omitempty"`   // Zone code

	// Placement and GPU information
	PlacementGroupNo    string `json:"placementGroupNo,omitempty"`    // Placement group
	PlacementGroupName  string `json:"placementGroupName,omitempty"`  // Placement group name
	FabricClusterPoolNo string `json:"fabricClusterPoolNo,omitempty"` // GPU cluster pool

	// Kubernetes reconciliation status
	Phase             string `json:"phase,omitempty"`             // Reconciliation phase
	Message           string `json:"message,omitempty"`           // Status message
	LastReconcileTime string `json:"lastReconcileTime,omitempty"` // Last reconciliation time

	// Observability fields
	ObservedGeneration int64 `json:"observedGeneration,omitempty"` // Last processed generation
}

// StatusCode represents common status codes from NCloud API
type StatusCode struct {
	Code     string `json:"code,omitempty"`
	CodeName string `json:"codeName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NCloudServer is the Schema for the ncloudservers API.
type NCloudServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NCloudServerSpec   `json:"spec,omitempty"`
	Status NCloudServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NCloudServerList contains a list of NCloudServer.
type NCloudServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NCloudServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NCloudServer{}, &NCloudServerList{})
}
