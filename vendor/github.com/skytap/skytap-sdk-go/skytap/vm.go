package skytap

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
)

const timestampFormat = "2006/01/02 15:04:05 -0700"

// Default URL paths
const (
	vmsLegacyBasePath = "/configurations/"
	vmsBasePath       = "/v2/configurations/"
	vmsPath           = "/vms/"
)

// VMsService is the contract for the services provided on the Skytap VM resource
type VMsService interface {
	List(ctx context.Context, environmentID string) (*VMListResult, error)
	Get(ctx context.Context, environmentID string, id string) (*VM, error)
	Create(ctx context.Context, environmentID string, opts *CreateVMRequest) (*VM, error)
	Update(ctx context.Context, environmentID string, id string, vm *UpdateVMRequest) (*VM, error)
	Delete(ctx context.Context, environmentID string, id string) error
}

// VMsServiceClient is the VMsService implementation
type VMsServiceClient struct {
	client *Client
}

// VM describes a virtual machines in the environment. It is legal to have 0 entries in this array
type VM struct {
	ID                     *string      `json:"id"`
	Name                   *string      `json:"name"`
	Runstate               *VMRunstate  `json:"runstate"`
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

// Hardware describes the VM's hardware configuration
type Hardware struct {
	CPUs                 *int    `json:"cpus"`
	SupportsMulticore    *bool   `json:"supports_multicore"`
	CpusPerSocket        *int    `json:"cpus_per_socket"`
	RAM                  *int    `json:"ram"`
	SVMs                 *int    `json:"svms"`
	GuestOS              *string `json:"guestOS"`
	MaxCPUs              *int    `json:"max_cpus"`
	MinRAM               *int    `json:"min_ram"`
	MaxRAM               *int    `json:"max_ram"`
	VncKeymap            *string `json:"vnc_keymap"`
	UUID                 *int    `json:"uuid"`
	Disks                []Disk  `json:"disks"`
	Storage              *int    `json:"storage"`
	Upgradable           *bool   `json:"upgradable"`
	InstanceType         *string `json:"instance_type"`
	TimeSyncEnabled      *bool   `json:"time_sync_enabled"`
	RTCStartTime         *string `json:"rtc_start_time"`
	CopyPasteEnabled     *bool   `json:"copy_paste_enabled"`
	NestedVirtualization *bool   `json:"nested_virtualization"`
	Architecture         *string `json:"architecture"`
}

// Disk describes the VM's hard drive configuration
type Disk struct {
	ID         *string `json:"id"`
	Size       *int    `json:"size"`
	Type       *string `json:"type"`
	Controller *string `json:"controller"`
	LUN        *string `json:"lun"`
	Name       *string `json:"name,omitempty"`
}

// Note describes a note on the VM
type Note struct {
	ID        *string `json:"id"`
	UserID    *int    `json:"user_id"`
	User      *User   `json:"user"`
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
	Text      *string `json:"text"`
}

// User describes the user who made the note
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

// Label describes a label attached to the VM
type Label struct {
	ID                       *string `json:"id"`
	Value                    *string `json:"value"`
	LabelCategory            *string `json:"label_category"`
	LabelCategoryID          *string `json:"label_category_id"`
	LabelCategorySingleValue *bool   `json:"label_category_single_value"`
}

// Credential describes credentials stored on the VM and available from the Credentials page in the UI
type Credential struct {
	ID   *string `json:"id"`
	Text *string `json:"text"`
}

// Container describes the containers running on the VM. If null, the VM is not a container host.
// To make the VM a container host, see Make the VM a container host.
// If the VM is a container host, this object contains the following fields:
type Container struct {
	ID              *int        `json:"id"`
	CID             *string     `json:"cid"`
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
	VMRunstate      *VMRunstate `json:"vm_runstate"`
	ConfigurationID *int        `json:"configuration_id"`
}

// VMRunstate enumerates the possible VM running states
type VMRunstate string

// The VM running states
const (
	VMRunstateStopped   VMRunstate = "stopped"
	VMRunstateSuspended VMRunstate = "suspended"
	VMRunstateRunning   VMRunstate = "running"
	VMRunstateReset     VMRunstate = "reset"
	VMRunstateHalted    VMRunstate = "halted"
	VMRunstateBusy      VMRunstate = "busy"
)

// Architecture is the system architecture
type Architecture int

// The architecture types
const (
	ArchitectureX86   Architecture = 0
	ArchitecturePower Architecture = 1
)

// CreateVMRequest describes the create the VM data
type CreateVMRequest struct {
	TemplateID string
	VMID       string
}

// createVMRequestAPI describes the create the VM data accepted by the API
type createVMRequestAPI struct {
	TemplateID string   `json:"template_id"`
	VMID       []string `json:"vm_ids"`
}

// UpdateVMRequest describes the update the VM data
type UpdateVMRequest struct {
	Name     *string         `json:"name,omitempty"`
	Runstate *VMRunstate     `json:"runstate,omitempty"`
	Hardware *UpdateHardware `json:"hardware,omitempty"`
}

// UpdateHardware describes the update data to update the VM cpu, ram and disks
type UpdateHardware struct {
	CPUs        *int         `json:"cpus,omitempty"`
	RAM         *int         `json:"ram,omitempty"`
	UpdateDisks *UpdateDisks `json:"disks,omitempty"`
}

// UpdateDisks defines a list of new disks
type UpdateDisks struct {
	NewDisks           []int                 `json:"new,omitempty"`
	RemoveDisks        map[string]RemoveDisk `json:"existing,omitempty"`
	DiskIdentification []DiskIdentification  `json:"identification,omitempty"`
}

// DiskIdentification defines a new size and name combination
type DiskIdentification struct {
	ID   *string
	Size *int
	Name *string
}

// RemoveDisk defines the disk to remove
type RemoveDisk struct {
	ID   *string `json:"id"`
	Size *int    `json:"size"`
}

// VMListResult is the listing request specific struct
type VMListResult struct {
	Value []VM
}

// List the vms
func (s *VMsServiceClient) List(ctx context.Context, environmentID string) (*VMListResult, error) {
	path := s.buildPath(false, environmentID, "") + "/vms"

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var vmListResponse VMListResult
	_, err = s.client.do(ctx, req, &vmListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &vmListResponse, nil
}

// Get a vm
func (s *VMsServiceClient) Get(ctx context.Context, environmentID string, id string) (*VM, error) {
	path := s.buildPath(false, environmentID, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var vm VM
	_, err = s.client.do(ctx, req, &vm)
	if err != nil {
		return nil, err
	}

	return &vm, nil
}

// Create a vm - returns an Environment
func (s *VMsServiceClient) Create(ctx context.Context, environmentID string, opts *CreateVMRequest) (*VM, error) {
	path := s.buildPath(true, environmentID, "")

	apiOpts := createVMRequestAPI{
		TemplateID: opts.TemplateID,
		VMID:       []string{opts.VMID},
	}

	req, err := s.client.newRequest(ctx, "PUT", path, apiOpts)
	if err != nil {
		return nil, err
	}

	var createdEnvironment Environment
	_, err = s.client.do(ctx, req, &createdEnvironment)
	if err != nil {
		return nil, err
	}

	// The create method returns an environment. The ID of the VM is not specified.
	// It is necessary to retrieve the most recently created vm.
	createdVM := mostRecentVM(createdEnvironment.VMs)

	return createdVM, nil
}

// Update a vm
func (s *VMsServiceClient) Update(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	if opts.Runstate != nil {
		return s.changeRunstate(ctx, environmentID, id, opts)
	} else if opts.Hardware != nil && opts.Hardware.UpdateDisks == nil {
		return s.changeNameCPURAM(ctx, environmentID, id, opts)
	}
	return s.updateHardware(ctx, environmentID, id, opts)
}

// Delete a vm
func (s *VMsServiceClient) Delete(ctx context.Context, environmentID string, id string) error {
	path := s.buildPath(false, environmentID, id)

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

// mostRecentVM returns the mose recent VM given a list of VMs
func mostRecentVM(vms []VM) *VM {
	sort.Slice(vms, func(i, j int) bool {
		time1, _ := time.Parse(timestampFormat, *vms[i].CreatedAt)
		time2, _ := time.Parse(timestampFormat, *vms[j].CreatedAt)
		return time1.After(time2)
	})
	return &vms[0]
}

func (s *VMsServiceClient) updateHardware(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	path := s.buildPath(false, environmentID, id)

	diskIdentification := opts.Hardware.UpdateDisks.DiskIdentification
	opts.Hardware.UpdateDisks.DiskIdentification = nil

	currentVM, err := s.Get(ctx, environmentID, id)
	// if started stop
	runstate := currentVM.Runstate
	if *runstate == VMRunstateRunning {
		_, err = s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateStopped)})
		if err != nil {
			return nil, err
		}
		err = s.waitForRunstate(&ctx, environmentID, id, VMRunstateStopped)
		if err != nil {
			return nil, err
		}
	}

	removes := buildRemoveList(currentVM, diskIdentification)

	requestCreate, err := s.client.newRequest(ctx, "PUT", path, opts)
	if err != nil {
		return nil, err
	}

	var updatedVM VM
	_, err = s.client.do(ctx, requestCreate, &updatedVM)
	if err != nil {
		return nil, err
	}

	// wait until disks as expected.
	expected := len(currentVM.Hardware.Disks) + len(opts.Hardware.UpdateDisks.NewDisks)
	if len(updatedVM.Hardware.Disks) != expected {
		vm, err := s.waitForDisks(&ctx, environmentID, id, expected)
		if err != nil {
			return nil, err
		}
		if vm == nil {
			return nil, fmt.Errorf("unable to retrieve the VM with the expected (%d) number of disks", expected)
		}
		updatedVM = *vm
	}

	matchUpExistingDisks(&updatedVM, diskIdentification, removes)
	matchUpNewDisks(&updatedVM, diskIdentification, removes)

	disks := updatedVM.Hardware.Disks

	if len(removes) > 0 {
		// delete phase
		opts.Hardware.CPUs = nil
		opts.Hardware.RAM = nil
		opts.Hardware.UpdateDisks.NewDisks = nil
		opts.Hardware.UpdateDisks.RemoveDisks = removes
		requestDelete, err := s.client.newRequest(ctx, "PUT", path, opts)
		if err != nil {
			return nil, err
		}
		_, err = s.client.do(ctx, requestDelete, &updatedVM)
		if err != nil {
			return nil, err
		}

		// wait until disks as expected.
		expected = len(disks) - len(opts.Hardware.UpdateDisks.RemoveDisks)
		if len(updatedVM.Hardware.Disks) != expected {
			vm, err := s.waitForDisks(&ctx, environmentID, id, expected)
			if err != nil {
				return nil, err
			}
			if vm == nil {
				return nil, fmt.Errorf("unable to retrieve the VM with the expected (%d) number of disks", expected)
			}
			updatedVM = *vm
		}

		// update new list of disks
		updateFinalDiskList(updatedVM, disks)
	}

	// if stopped start
	if *runstate == VMRunstateRunning {
		_, err = s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateRunning)})
		if err != nil {
			return nil, err
		}
	}

	return &updatedVM, nil
}

func (s *VMsServiceClient) changeRunstate(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	path := s.buildPath(false, environmentID, id)

	opts.Name = nil
	opts.Hardware = nil
	requestCreate, err := s.client.newRequest(ctx, "PUT", path, opts)
	if err != nil {
		return nil, err
	}

	var updatedVM VM
	_, err = s.client.do(ctx, requestCreate, &updatedVM)
	if err != nil {
		return nil, err
	}
	return &updatedVM, nil
}

func (s *VMsServiceClient) changeNameCPURAM(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	path := s.buildPath(false, environmentID, id)

	currentVM, err := s.Get(ctx, environmentID, id)
	// if started stop
	runstate := currentVM.Runstate
	if *runstate == VMRunstateRunning {
		_, err := s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateStopped)})
		if err != nil {
			return nil, err
		}
		err = s.waitForRunstate(&ctx, environmentID, id, VMRunstateStopped)
		if err != nil {
			return nil, err
		}
	}

	requestCreate, err := s.client.newRequest(ctx, "PUT", path, opts)
	if err != nil {
		return nil, err
	}

	var updatedVM VM
	_, err = s.client.do(ctx, requestCreate, &updatedVM)
	if err != nil {
		return nil, err
	}

	// if stopped start
	if *runstate == VMRunstateRunning {
		_, err := s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateRunning)})
		if err != nil {
			return nil, err
		}
	}

	return &updatedVM, nil
}

func matchUpExistingDisks(vm *VM, identifications []DiskIdentification, ignored map[string]RemoveDisk) {
	for idx := range vm.Hardware.Disks {
		// ignore os disk
		if idx > 0 {
			for _, id := range identifications {
				if id.ID != nil && *id.ID == *vm.Hardware.Disks[idx].ID {
					if _, ok := ignored[*id.ID]; !ok {
						vm.Hardware.Disks[idx].Name = id.Name
						break
					}
				}
			}
		}
	}
}

func matchUpNewDisks(vm *VM, identifications []DiskIdentification, ignored map[string]RemoveDisk) {
	for _, id := range identifications {
		if id.ID == nil || *id.ID == "" {
			for idx, disk := range vm.Hardware.Disks {
				// ignore os disk
				if idx > 0 {
					// a new disk
					if _, ok := ignored[*disk.ID]; !ok {
						if disk.Name == nil {
							vm.Hardware.Disks[idx].Name = id.Name
							break
						}
					}
				}
			}
		}
	}
}

// wait until new disks added
func (s *VMsServiceClient) waitForDisks(ctx *context.Context, environmentID string, id string, expectedDiskCount int) (*VM, error) {
	var makeRequest = true
	var vm *VM
	var err error
	for i := 0; i < s.client.retryCount+1 && makeRequest; i++ {
		vm, err = s.Get(*ctx, environmentID, id)

		if err != nil {
			break
		}

		makeRequest = len(vm.Hardware.Disks) != expectedDiskCount

		if makeRequest {
			seconds := s.client.retryAfter
			log.Printf("[INFO] retrying after %d second(s)\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	}
	return vm, nil
}

// wait for runstate
func (s *VMsServiceClient) waitForRunstate(ctx *context.Context, environmentID string, id string, runstate VMRunstate) error {
	var makeRequest = true
	var err error
	for i := 0; i < s.client.retryCount+1 && makeRequest; i++ {
		vm, err := s.Get(*ctx, environmentID, id)

		if err != nil {
			break
		}

		makeRequest = *vm.Runstate != runstate

		if makeRequest {
			seconds := s.client.retryAfter
			log.Printf("[INFO] retrying after %d second(s)\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	}
	return err
}

func (s *VMsServiceClient) buildPath(legacy bool, environmentID string, vmID string) string {
	var path string
	if legacy {
		path = vmsLegacyBasePath
	} else {
		path = vmsBasePath
	}
	path = path + environmentID
	if vmID != "" {
		path += vmsPath + vmID
	}
	return path
}

func updateFinalDiskList(vm VM, disks []Disk) {
	for idx := range vm.Hardware.Disks {
		// forget os disk
		if idx > 0 {
			for _, name := range disks {
				if *name.ID == *vm.Hardware.Disks[idx].ID {
					vm.Hardware.Disks[idx].Name = name.Name
					break
				}
			}
		}
	}
}

// clear the removes from this phase
func buildRemoveList(vm *VM, diskIDs []DiskIdentification) map[string]RemoveDisk {
	removes := make(map[string]RemoveDisk, 0)
	for idx, disk := range vm.Hardware.Disks {
		if idx > 0 {
			found := false
			for _, diskID := range diskIDs {
				if diskID.ID != nil && *diskID.ID == *disk.ID {
					found = true
					break
				}
			}
			if !found {
				removes[*disk.ID] = RemoveDisk{
					ID: strToPtr(*disk.ID),
				}
			}
		}
	}
	return removes
}
