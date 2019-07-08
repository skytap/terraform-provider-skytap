package skytap

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
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
	NewDisks           []int                   `json:"new,omitempty"`
	ExistingDisks      map[string]ExistingDisk `json:"existing,omitempty"`
	DiskIdentification []DiskIdentification    `json:"identification,omitempty"`
	OSSize             *int                    `json:"os_size,omitempty"`
}

// DiskIdentification defines a new size and name combination
type DiskIdentification struct {
	ID   *string
	Size *int
	Name *string
}

// ExistingDisk defines the disk to change
type ExistingDisk struct {
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
	_, err = s.client.do(ctx, req, &vmListResponse.Value, nil, nil)
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
	_, err = s.client.do(ctx, req, &vm, nil, nil)
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

	var environment Environment
	_, err = s.client.do(ctx, req, &environment, envRunStateNotBusy(environmentID), opts)
	if err != nil {
		return nil, err
	}

	// The create method returns an environment. The ID of the VM is not specified.
	// It is necessary to retrieve the most recently created vm.
	createdVM, err := mostRecentVM(&environment)
	if err != nil {
		return nil, err
	}

	return createdVM, nil
}

// Update a vm
func (s *VMsServiceClient) Update(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	if opts.Runstate != nil && opts.Hardware == nil {
		return s.changeRunstate(ctx, environmentID, id, opts)
	} else if opts.Hardware == nil || opts.Hardware.UpdateDisks == nil || opts.Hardware.UpdateDisks.DiskIdentification == nil {
		return nil, fmt.Errorf("expecting the DiskIdentification list to be populated")
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

	_, err = s.client.do(ctx, req, nil, envRunStateNotBusyWithVM(environmentID, id), nil)
	if err != nil {
		return err
	}

	return nil
}

// mostRecentVM returns the mose recent VM given a list of VMs
func mostRecentVM(environment *Environment) (*VM, error) {
	vms := environment.VMs
	if len(vms) == 0 {
		return nil, fmt.Errorf("could not find any VMs in environment (%s)", *environment.ID)
	}
	sort.Slice(vms, func(i, j int) bool {
		time1, _ := time.Parse(timestampFormat, *vms[i].CreatedAt)
		time2, _ := time.Parse(timestampFormat, *vms[j].CreatedAt)
		return time1.After(time2)
	})
	return &vms[0], nil
}

func (s *VMsServiceClient) updateHardware(ctx context.Context, environmentID string, id string, opts *UpdateVMRequest) (*VM, error) {
	path := s.buildPath(false, environmentID, id)

	osDiskSize := opts.Hardware.UpdateDisks.OSSize
	opts.Hardware.UpdateDisks.OSSize = nil
	diskIdentification := opts.Hardware.UpdateDisks.DiskIdentification
	opts.Hardware.UpdateDisks.DiskIdentification = nil

	vm, err := s.Get(ctx, environmentID, id)
	if err != nil {
		return nil, err
	}
	// if started stop
	runstate := *vm.Runstate
	if runstate == VMRunstateRunning {
		_, err = s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateStopped)})
		if err != nil {
			return nil, err
		}
	}

	removes := buildRemoveList(vm, diskIdentification)
	updates := buildUpdateList(vm, diskIdentification)
	addOSDiskResize(osDiskSize, vm, updates)
	if len(updates) > 0 {
		opts.Hardware.UpdateDisks.ExistingDisks = updates
	} else if len(opts.Hardware.UpdateDisks.NewDisks) == 0 {
		opts.Hardware.UpdateDisks = nil
	}

	state := vmRequestRunStateStopped(environmentID, id)
	state.diskIdentification = diskIdentification
	if opts.Hardware.UpdateDisks != nil || opts.Hardware.RAM != nil || opts.Hardware.CPUs != nil || opts.Name != nil {
		requestCreate, err := s.client.newRequest(ctx, "PUT", path, opts)
		if err != nil {
			return nil, err
		}
		_, err = s.client.do(ctx, requestCreate, &vm, state, opts)
		if err != nil {
			return nil, err
		}

		vm, err = s.Get(ctx, environmentID, id)
		if err != nil {
			return nil, err
		}
	}
	matchUpExistingDisks(vm, diskIdentification, removes)
	matchUpNewDisks(vm, diskIdentification, removes)

	disks := vm.Hardware.Disks

	if len(removes) > 0 {
		// delete phase
		opts.Hardware.CPUs = nil
		opts.Hardware.RAM = nil
		if opts.Hardware.UpdateDisks == nil {
			opts.Hardware.UpdateDisks = &UpdateDisks{}
		}
		opts.Hardware.UpdateDisks.NewDisks = nil
		opts.Hardware.UpdateDisks.ExistingDisks = removes
		requestDelete, err := s.client.newRequest(ctx, "PUT", path, opts)
		if err != nil {
			return nil, err
		}
		_, err = s.client.do(ctx, requestDelete, &vm, state, opts)
		if err != nil {
			return nil, err
		}

		vm, err = s.Get(ctx, environmentID, id)
		if err != nil {
			return nil, err
		}
		updateFinalDiskList(vm, disks)
	}

	// if stopped start
	if runstate == VMRunstateRunning {
		_, err = s.changeRunstate(ctx, environmentID, id, &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateRunning)})
		if err != nil {
			return nil, err
		}
		vm, err = s.Get(ctx, environmentID, id)
		if err != nil {
			return nil, err
		}
		updateFinalDiskList(vm, disks)
	}

	return vm, nil
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
	_, err = s.client.do(ctx, requestCreate, &updatedVM, envRunStateNotBusyWithVM(environmentID, id), opts)
	if err != nil {
		return nil, err
	}
	return &updatedVM, nil
}

func matchUpExistingDisks(vm *VM, identifications []DiskIdentification, ignored map[string]ExistingDisk) {
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

func matchUpNewDisks(vm *VM, identifications []DiskIdentification, ignored map[string]ExistingDisk) {
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

func updateFinalDiskList(vm *VM, disks []Disk) {
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

func buildRemoveList(vm *VM, diskIDs []DiskIdentification) map[string]ExistingDisk {
	removes := make(map[string]ExistingDisk, 0)
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
				removes[*disk.ID] = ExistingDisk{
					ID: strToPtr(*disk.ID),
				}
			}
		}
	}
	return removes
}

func buildUpdateList(vm *VM, diskIDs []DiskIdentification) map[string]ExistingDisk {
	updates := make(map[string]ExistingDisk, 0)
	for idx, disk := range vm.Hardware.Disks {
		if idx > 0 {
			var changed *ExistingDisk
			for _, diskID := range diskIDs {
				if diskID.ID != nil && *diskID.ID == *disk.ID && diskID.Size != nil && *diskID.Size > *disk.Size {
					changed = &ExistingDisk{
						ID:   disk.ID,
						Size: diskID.Size,
					}
					break
				}
			}
			if changed != nil {
				updates[*disk.ID] = *changed
			}
		}
	}
	return updates
}

func addOSDiskResize(osDiskSize *int, vm *VM, updates map[string]ExistingDisk) {
	if osDiskSize != nil && (*vm.Hardware.Disks[0].Size) < *osDiskSize {
		updates[*vm.Hardware.Disks[0].ID] = ExistingDisk{
			ID:   vm.Hardware.Disks[0].ID,
			Size: osDiskSize,
		}
	}
}

func (payload *CreateVMRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	env, err := c.Environments.Get(ctx, *state.environmentID)
	if err != nil {
		return requestNotAsExpected, false
	}
	logEnvironmentStatus(env)
	if *env.Runstate != EnvironmentRunstateBusy {
		return "", true
	}
	return "VM environment not ready", false
}

func (payload *UpdateVMRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	vm, err := c.VMs.Get(ctx, *state.environmentID, *state.vmID)
	if err != nil {
		return requestNotAsExpected, false
	}
	logVMStatus(vm)
	if payload.Runstate != nil && payload.Hardware == nil {
		if *payload.Runstate == *vm.Runstate {
			return "", true
		}
		return "VM not ready", false
	}
	actual := payload.buildComparison(vm, state.diskIdentification)
	if payload.string() == actual.string() {
		return "", true
	}
	return "VM not ready", false
}

func (payload *UpdateVMRequest) buildComparison(vm *VM, diskIdentification []DiskIdentification) *UpdateVMRequest {
	update := &UpdateVMRequest{}

	if payload.Name != nil {
		update.Name = vm.Name
	}
	if payload.Runstate != nil {
		update.Runstate = vm.Runstate
	}
	if payload.Hardware != nil {
		update.Hardware = &UpdateHardware{}
		if payload.Hardware.CPUs != nil {
			update.Hardware.CPUs = vm.Hardware.CPUs
		}
		if payload.Hardware.RAM != nil {
			update.Hardware.RAM = vm.Hardware.RAM
		}
		if payload.Hardware.UpdateDisks != nil {
			update.Hardware.UpdateDisks = payload.buildDiskStructure(vm, diskIdentification)
		}
	}
	if payload.Hardware.CPUs == nil && payload.Hardware.RAM == nil && payload.Hardware.UpdateDisks == nil {
		payload.Hardware = nil
	}
	return update
}

func (payload *UpdateVMRequest) buildDiskStructure(vm *VM, diskIdentification []DiskIdentification) *UpdateDisks {
	if diskIdentification == nil {
		log.Println("[ERROR] SDK cannot compare disks because the disk identification structure is empty.")
		return nil
	}
	if vm.Hardware.Disks == nil {
		return nil
	}
	existing := payload.buildVMExistingDisks(vm.Hardware.Disks)
	newDisks := payload.buildVMNewDisks(vm.Hardware.Disks, diskIdentification)

	update := &UpdateDisks{}
	if existing != nil {
		update.ExistingDisks = existing
	}
	if newDisks != nil {
		update.NewDisks = newDisks
	}
	if update.ExistingDisks == nil && update.NewDisks == nil {
		return nil
	}
	return update
}

func (payload *UpdateVMRequest) buildVMExistingDisks(disks []Disk) map[string]ExistingDisk {
	var existingDiskPayload map[string]ExistingDisk
	if payload.Hardware != nil &&
		payload.Hardware.UpdateDisks != nil &&
		payload.Hardware.UpdateDisks.ExistingDisks != nil {
		existingDiskPayload = payload.Hardware.UpdateDisks.ExistingDisks
	}
	existingDisks := make(map[string]ExistingDisk)
	disks = disks[0:]
	for _, disk := range disks {
		if _, ok := existingDiskPayload[*disk.ID]; ok {
			existingDisks[*disk.ID] = ExistingDisk{
				ID:   disk.ID,
				Size: disk.Size,
			}
		}
	}
	for id, disk := range existingDiskPayload {
		if disk.Size == nil {
			existingDisks[id] = disk
		}
	}
	if len(existingDisks) == 0 {
		return nil
	}
	return existingDisks
}

func (payload *UpdateVMRequest) buildVMNewDisks(disks []Disk, identification []DiskIdentification) []int {
	if payload.Hardware == nil ||
		payload.Hardware.UpdateDisks == nil ||
		payload.Hardware.UpdateDisks.NewDisks == nil {
		return nil
	}
	newDisks := make([]int, 0)
	disks = disks[1:]
	for _, disk := range disks {
		if identification == nil {
			newDisks = append(newDisks, *disk.Size)
		} else {
			found := false
			for _, diskID := range identification {
				if diskID.ID != nil && *diskID.ID == *disk.ID {
					found = true
					break
				}
			}
			if !found {
				newDisks = append(newDisks, *disk.Size)
			}
		}
	}
	if len(newDisks) == 0 {
		return nil
	}
	sort.Ints(payload.Hardware.UpdateDisks.NewDisks)
	sort.Ints(newDisks)
	return newDisks
}

func (payload *UpdateVMRequest) string() string {
	name := ""
	runState := ""
	hardware := ""

	if payload.Name != nil {
		name = *payload.Name
	}
	if payload.Runstate != nil {
		runState = string(*payload.Runstate)
	}
	if payload.Hardware != nil {
		hardware = payload.Hardware.string()
	}
	s := fmt.Sprintf("%s%s%s",
		name,
		runState,
		hardware)
	log.Printf("[DEBUG] SDK update vm payload: %s", s)
	return s
}

func (payload *UpdateHardware) string() string {
	cpus := ""
	ram := ""
	updateDisks := ""

	if payload.CPUs != nil {
		cpus = fmt.Sprintf("%d", *payload.CPUs)
	}
	if payload.RAM != nil {
		ram = fmt.Sprintf("%d", *payload.RAM)
	}
	if payload.UpdateDisks != nil {
		updateDisks = payload.UpdateDisks.string()
	}
	return fmt.Sprintf("%s%s%s",
		cpus,
		ram,
		updateDisks)
}

func (payload *UpdateDisks) string() string {
	osSize := ""
	disksExisting := ""
	disksNew := ""

	if payload.OSSize != nil {
		osSize = fmt.Sprintf("%d", *payload.OSSize)
	}
	if payload.ExistingDisks != nil {
		disks := make([]string, 0)
		for _, disk := range payload.ExistingDisks {
			id := ""
			size := ""
			if disk.ID != nil {
				id = *disk.ID
			}
			if disk.Size != nil {
				size = fmt.Sprintf("%d", *disk.Size)
			}
			disks = append(disks, fmt.Sprintf("%s:%s", id, size))
		}
		sort.Strings(disks)
		disksExisting = strings.Join(disks, ",")
	}
	if payload.NewDisks != nil {
		disks := make([]string, 0)
		for _, disk := range payload.NewDisks {
			disks = append(disks, fmt.Sprintf("%d", disk))
		}
		disksNew = strings.Join(disks, ",")
	}
	return fmt.Sprintf("%s%s%s",
		osSize,
		disksExisting,
		disksNew)
}

func logVMStatus(vm *VM) {
	if vm.RateLimited != nil && *vm.RateLimited {
		log.Printf("[INFO] SDK VM rate limiting detected\n")
	}
}
