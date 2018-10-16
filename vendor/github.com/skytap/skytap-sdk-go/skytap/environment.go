package skytap

import (
	"context"
	"fmt"
)

const (
	environmentLegacyBasePath = "/configurations"
	environmentBasePath       = "/v2/configurations"
)

type EnvironmentsService interface {
	List(ctx context.Context) (*EnvironmentListResult, error)
	Get(ctx context.Context, id string) (*Environment, error)
	Create(ctx context.Context, createEnvironmentRequest *CreateEnvironmentRequest) (*Environment, error)
	Update(ctx context.Context, id string, updateEnvironmentRequest *UpdateEnvironmentRequest) (*Environment, error)
	Delete(ctx context.Context, id string) error
}

// Environment service implementation
type EnvironmentsServiceClient struct {
	client *Client
}

// Environment resource struct definitions
type Environment struct {
	ID                      *string              `json:"id"`
	URL                     *string              `json:"url"`
	Name                    *string              `json:"name"`
	Description             *string              `json:"description"`
	Errors                  []string             `json:"errors"`
	ErrorDetails            []string             `json:"error_details"`
	Runstate                *EnvironmentRunstate `json:"runstate"`
	RateLimited             *bool                `json:"rate_limited"`
	LastRun                 *string              `json:"last_run"`
	SuspendOnIdle           *int                 `json:"suspend_on_idle"`
	SuspendAtTime           *string              `json:"suspend_at_time"`
	OwnerURL                *string              `json:"owner_url"`
	OwnerName               *string              `json:"owner_name"`
	OwnerID                 *string              `json:"owner_id"`
	VMCount                 *int                 `json:"vm_count"`
	Storage                 *int                 `json:"storage"`
	NetworkCount            *int                 `json:"network_count"`
	CreatedAt               *string              `json:"created_at"`
	Region                  *string              `json:"region"`
	RegionBackend           *string              `json:"region_backend"`
	Svms                    *int                 `json:"svms"`
	CanSaveAsTemplate       *bool                `json:"can_save_as_template"`
	CanCopy                 *bool                `json:"can_copy"`
	CanDelete               *bool                `json:"can_delete"`
	CanChangeState          *bool                `json:"can_change_state"`
	CanShare                *bool                `json:"can_share"`
	CanEdit                 *bool                `json:"can_edit"`
	LabelCount              *int                 `json:"label_count"`
	LabelCategoryCount      *int                 `json:"label_category_count"`
	CanTag                  *bool                `json:"can_tag"`
	Tags                    []Tag                `json:"tags"`
	TagList                 *string              `json:"tag_list"`
	Alerts                  []string             `json:"alerts"`
	PublishedServiceCount   *int                 `json:"published_service_count"`
	PublicIPCount           *int                 `json:"public_ip_count"`
	AutoSuspendDescription  *string              `json:"auto_suspend_description"`
	Stages                  []Stage              `json:"stages"`
	StagedExecution         *StagedExecution     `json:"staged_execution"`
	SequencingEnabled       *bool                `json:"sequencing_enabled"`
	NoteCount               *int                 `json:"note_count"`
	ProjectCountForUser     *int                 `json:"project_count_for_user"`
	ProjectCount            *int                 `json:"project_count"`
	PublishSetCount         *int                 `json:"publish_set_count"`
	ScheduleCount           *int                 `json:"schedule_count"`
	VpnCount                *int                 `json:"vpn_count"`
	OutboundTraffic         *bool                `json:"outbound_traffic"`
	Routable                *bool                `json:"routable"`
	Vms                     []Vm                 `json:"vms"`
	Networks                []Network            `json:"networks"`
	ContainersCount         *int                 `json:"containers_count"`
	ContainerHostsCount     *int                 `json:"container_hosts_count"`
	PlatformErrors          []string             `json:"platform_errors"`
	SvmsByArchitecture      *SvmsByArchitecture  `json:"svms_by_architecture"`
	AllVmsSupportSuspend    *bool                `json:"all_vms_support_suspend"`
	ShutdownOnIdle          *int                 `json:"shutdown_on_idle"`
	ShutdownAtTime          *string              `json:"shutdown_at_time"`
	AutoShutdownDescription *string              `json:"auto_shutdown_description"`
}

type Tag struct {
	Id    *string `json:"id"`
	Value *string `json:"value"`
}

type Stage struct {
	DelayAfterFinishSeconds *int     `json:"delay_after_finish_seconds"`
	Index                   *int     `json:"index"`
	VMIds                   []string `json:"vm_ids"`
}

type StagedExecution struct {
	ActionType                          *string  `json:"action_type"`
	CurrentStageDelayAfterFinishSeconds *int     `json:"current_stage_delay_after_finish_seconds"`
	CurrentStageIndex                   *int     `json:"current_stage_index"`
	CurrentStageFinishedAt              *string  `json:"current_stage_finished_at"`
	VMIds                               []string `json:"vm_ids"`
}

type Vm struct {
	ID                     *string      `json:"id"`
	Name                   *string      `json:"name"`
	Runstate               *VmRunstate  `json:"runstate"`
	RateLimited            *bool        `json:"rate_limited"`
	Hardware               *Hardware    `json:"hardware"`
	Error                  *bool        `json:"error"`
	ErrorDetails           *bool        `json:"error_details"`
	AssetID                *string      `json:"asset_id"`
	HardwareVersion        *int         `json:"hardware_version"`
	MaxHardwareVersion     *int         `json:"max_hardware_version"`
	Interfaces             []Interface  `json:"interfaces"`
	Notes                  []Note       `json:"notes"`
	Labels                 []Label      `json:"labels"`
	Credentials            []Credential `json:"credentials"`
	DesktopResizable       *bool        `json:"desktop_resizable"`
	LocalMouseCursor       *bool        `json:"local_mouse_cursor"`
	MaintenanceLockEngaged *bool        `json:"maintenance_lock_engaged"`
	RegionBackend          *string      `json:"region_backend"`
	CreatedAt              *string      `json:"created_at"`
	SupportsSuspend        *bool        `json:"supports_suspend"`
	CanChangeObjectState   *bool        `json:"can_change_object_state"`
	Containers             []Container  `json:"containers"`
	ConfigurationURL       *string      `json:"configuration_url"`
}

type Hardware struct {
	Cpus                 *int    `json:"cpus"`
	SupportsMulticore    *bool   `json:"supports_multicore"`
	CpusPerSocket        *int    `json:"cpus_per_socket"`
	RAM                  *int    `json:"ram"`
	Svms                 *int    `json:"svms"`
	GuestOS              *string `json:"guestOS"`
	MaxCpus              *int    `json:"max_cpus"`
	MinRAM               *int    `json:"min_ram"`
	MaxRAM               *int    `json:"max_ram"`
	VncKeymap            *string `json:"vnc_keymap"`
	UUID                 *int    `json:"uuid"`
	Disks                []Disk  `json:"disks"`
	Storage              *int    `json:"storage"`
	Upgradable           *bool   `json:"upgradable"`
	InstanceType         *string `json:"instance_type"`
	TimeSyncEnabled      *bool   `json:"time_sync_enabled"`
	RtcStartTime         *string `json:"rtc_start_time"`
	CopyPasteEnabled     *bool   `json:"copy_paste_enabled"`
	NestedVirtualization *bool   `json:"nested_virtualization"`
	Architecture         *string `json:"architecture"`
}

type Disk struct {
	ID         *string `json:"id"`
	Size       *int    `json:"size"`
	Type       *string `json:"type"`
	Controller *string `json:"controller"`
	Lun        *string `json:"lun"`
}

type Interface struct {
	ID                  *string              `json:"id"`
	IP                  *string              `json:"ip"`
	Hostname            *string              `json:"hostname"`
	Mac                 *string              `json:"mac"`
	ServicesCount       *int                 `json:"services_count"`
	Services            []Service            `json:"services"`
	PublicIpsCount      *int                 `json:"public_ips_count"`
	PublicIps           []map[string]string  `json:"public_ips"`
	VMID                *string              `json:"vm_id"`
	VMName              *string              `json:"vm_name"`
	Status              *string              `json:"status"`
	NetworkID           *string              `json:"network_id"`
	NetworkName         *string              `json:"network_name"`
	NetworkURL          *string              `json:"network_url"`
	NetworkType         *string              `json:"network_type"`
	NetworkSubnet       *string              `json:"network_subnet"`
	NicType             *string              `json:"nic_type"`
	SecondaryIps        []SecondaryIp        `json:"secondary_ips"`
	PublicIPAttachments []PublicIpAttachment `json:"public_ip_attachments"`
}

type Service struct {
	ID           *string `json:"id"`
	InternalPort *int    `json:"internal_port"`
	ExternalIP   *string `json:"external_ip"`
	ExternalPort *int    `json:"external_port"`
}

type SecondaryIp struct {
	ID      *string `json:"id"`
	Address *string `json:"address"`
}

type PublicIpAttachment struct {
	ID                    *int    `json:"id"`
	PublicIPAttachmentKey *int    `json:"public_ip_attachment_key"`
	Address               *string `json:"address"`
	ConnectType           *int    `json:"connect_type"`
	Hostname              *string `json:"hostname"`
	DNSName               *string `json:"dns_name"`
	PublicIPKey           *string `json:"public_ip_key"`
}

type Note struct {
	ID        *string `json:"id"`
	UserID    *int    `json:"user_id"`
	User      *User   `json:"user"`
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
	Text      *string `json:"text"`
}

type User struct {
	ID        *string `json:"id"`
	URL       *string `json:"url"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	LoginName *string `json:"login_name"`
	Email     *string `json:"email"`
	Title     *string `json:"title"`
	Deleted   *bool   `json:"deleted"`
}

type Label struct {
	ID                       *string `json:"id"`
	Value                    *string `json:"value"`
	LabelCategory            *string `json:"label_category"`
	LabelCategoryID          *string `json:"label_category_id"`
	LabelCategorySingleValue *bool   `json:"label_category_single_value"`
}

type Credential struct {
	ID   *string `json:"id"`
	Text *string `json:"text"`
}

type Container struct {
	ID              *int        `json:"id"`
	Cid             *string     `json:"cid"`
	Name            *string     `json:"name"`
	Image           *string     `json:"image"`
	CreatedAt       *string     `json:"created_at"`
	LastRun         *string     `json:"last_run"`
	CanChangeState  *bool       `json:"can_change_state"`
	CanDelete       *bool       `json:"can_delete"`
	Status          *string     `json:"status"`
	Privileged      *bool       `json:"privileged"`
	VMID            *int        `json:"vm_id"`
	VMName          *string     `json:"vm_name"`
	VMRunstate      *VmRunstate `json:"vm_runstate"`
	ConfigurationID *int        `json:"configuration_id"`
}

type Network struct {
	ID                  *string         `json:"id"`
	URL                 *string         `json:"url"`
	Name                *string         `json:"name"`
	NetworkType         *string         `json:"network_type"`
	Subnet              *string         `json:"subnet"`
	SubnetAddr          *string         `json:"subnet_addr"`
	SubnetSize          *int            `json:"subnet_size"`
	Gateway             *string         `json:"gateway"`
	PrimaryNameserver   *string         `json:"primary_nameserver"`
	SecondaryNameserver *string         `json:"secondary_nameserver"`
	Region              *string         `json:"region"`
	Domain              *string         `json:"domain"`
	VpnAttachments      []VpnAttachment `json:"vpn_attachments"`
	Tunnelable          *bool           `json:"tunnelable"`
	Tunnels             []Tunnel        `json:"tunnels"`
}

type VpnAttachment struct {
	ID        *string               `json:"id"`
	Connected *bool                 `json:"connected"`
	Network   *VpnAttachmentNetwork `json:"network"`
	Vpn       *Vpn                  `json:"vpn"`
}

type VpnAttachmentNetwork struct {
	ID              *string `json:"id"`
	Subnet          *string `json:"subnet"`
	NetworkName     *string `json:"network_name"`
	ConfigurationID *string `json:"configuration_id"`
}

type Vpn struct {
	ID            *string `json:"id"`
	Name          *string `json:"name"`
	Enabled       *bool   `json:"enabled"`
	NatEnabled    *bool   `json:"nat_enabled"`
	RemoteSubnets *string `json:"remote_subnets"`
	RemotePeerIP  *string `json:"remote_peer_ip"`
	CanReconnect  *bool   `json:"can_reconnect"`
}

type Tunnel struct {
	ID            *string  `json:"id"`
	Status        *string  `json:"status"`
	Error         *string  `json:"error"`
	SourceNetwork *Network `json:"source_network"`
	TargetNetwork *Network `json:"target_network"`
}

type SvmsByArchitecture struct {
	X86   *int `json:"x86"`
	Power *int `json:"power"`
}

type EnvironmentRunstate string

const (
	EnvironmentRunstateStopped   EnvironmentRunstate = "stopped"
	EnvironmentRunstateSuspended EnvironmentRunstate = "suspended"
	EnvironmentRunstateRunning   EnvironmentRunstate = "running"
)

type VmRunstate string

const (
	VmRunstateStopped   VmRunstate = "stopped"
	VmRunstateSuspended VmRunstate = "suspended"
	VmRunstateRunning   VmRunstate = "running"
	VmRunstateReset     VmRunstate = "reset"
	VmRunstateHalted    VmRunstate = "halted"
)

type Architecture int

const (
	ArchitectureX86   Architecture = 0
	ArchitecturePower Architecture = 1
)

// Request specific structs
type EnvironmentListResult struct {
	Value []Environment
}

type CreateEnvironmentRequest struct {
	TemplateId      *string `json:"template_id,omitempty"`
	ProjectId       *string `json:"project_id,omitempty"`
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Owner           *string `json:"owner,omitempty"`
	OutboundTraffic *bool   `json:"outbound_traffic,omitempty"`
	Routable        *bool   `json:"routable,omitempty"`
	SuspendOnIdle   *int    `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int    `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string `json:"shutdown_at_time,omitempty"`
}

type UpdateEnvironmentRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Owner           *string `json:"owner,omitempty"`
	OutboundTraffic *bool   `json:"outbound_traffic,omitempty"`
	Routable        *bool   `json:"routable,omitempty"`
	SuspendOnIdle   *int    `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int    `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string `json:"shutdown_at_time,omitempty"`
}

func (s *EnvironmentsServiceClient) List(ctx context.Context) (*EnvironmentListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", environmentBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var environmentsListResponse EnvironmentListResult
	_, err = s.client.do(ctx, req, &environmentsListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &environmentsListResponse, nil
}

func (s *EnvironmentsServiceClient) Get(ctx context.Context, id string) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

func (s *EnvironmentsServiceClient) Create(ctx context.Context, request *CreateEnvironmentRequest) (*Environment, error) {
	req, err := s.client.newRequest(ctx, "POST", environmentLegacyBasePath, request)
	if err != nil {
		return nil, err
	}

	var createdEnvironment Environment
	_, err = s.client.do(ctx, req, &createdEnvironment)
	if err != nil {
		return nil, err
	}

	updateOpts := &UpdateEnvironmentRequest{
		Name:            request.Name,
		Description:     request.Description,
		Owner:           request.Owner,
		OutboundTraffic: request.OutboundTraffic,
		Routable:        request.Routable,
		SuspendOnIdle:   request.SuspendOnIdle,
		SuspendAtTime:   request.SuspendAtTime,
		ShutdownOnIdle:  request.ShutdownOnIdle,
		ShutdownAtTime:  request.ShutdownAtTime,
	}

	// update environment after creation to establish the resource information.
	environment, err := s.Update(ctx, String(createdEnvironment.ID), updateOpts)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (s *EnvironmentsServiceClient) Update(ctx context.Context, id string, updateEnvironment *UpdateEnvironmentRequest) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "PUT", path, updateEnvironment)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

func (s *EnvironmentsServiceClient) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", environmentLegacyBasePath, id)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	_, err = s.client.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
