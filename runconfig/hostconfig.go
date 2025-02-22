package runconfig

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/docker/docker/pkg/blkiodev"
	"github.com/docker/docker/pkg/nat"
	"github.com/docker/docker/pkg/stringutils"
	"github.com/docker/docker/pkg/ulimit"
)

// KeyValuePair is a structure that hold a value for a key.
type KeyValuePair struct {
	Key   string
	Value string
}

// NetworkMode represents the container network stack.
type NetworkMode string

// IsolationLevel represents the isolation level of a container. The supported
// values are platform specific
type IsolationLevel string

// IsDefault indicates the default isolation level of a container. On Linux this
// is the native driver. On Windows, this is a Windows Server Container.
func (i IsolationLevel) IsDefault() bool {
	return strings.ToLower(string(i)) == "default" || string(i) == ""
}

// IpcMode represents the container ipc stack.
type IpcMode string

// IsPrivate indicates whether the container uses it's private ipc stack.
func (n IpcMode) IsPrivate() bool {
	return !(n.IsHost() || n.IsContainer())
}

// IsHost indicates whether the container uses the host's ipc stack.
func (n IpcMode) IsHost() bool {
	return n == "host"
}

// IsContainer indicates whether the container uses a container's ipc stack.
func (n IpcMode) IsContainer() bool {
	parts := strings.SplitN(string(n), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// Valid indicates whether the ipc stack is valid.
func (n IpcMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	case "container":
		if len(parts) != 2 || parts[1] == "" {
			return false
		}
	default:
		return false
	}
	return true
}

// Container returns the name of the container ipc stack is going to be used.
func (n IpcMode) Container() string {
	parts := strings.SplitN(string(n), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// UTSMode represents the UTS namespace of the container.
type UTSMode string

// IsPrivate indicates whether the container uses it's private UTS namespace.
func (n UTSMode) IsPrivate() bool {
	return !(n.IsHost())
}

// IsHost indicates whether the container uses the host's UTS namespace.
func (n UTSMode) IsHost() bool {
	return n == "host"
}

// Valid indicates whether the UTS namespace is valid.
func (n UTSMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// PidMode represents the pid stack of the container.
type PidMode string

// IsPrivate indicates whether the container uses it's private pid stack.
func (n PidMode) IsPrivate() bool {
	return !(n.IsHost())
}

// IsHost indicates whether the container uses the host's pid stack.
func (n PidMode) IsHost() bool {
	return n == "host"
}

// Valid indicates whether the pid stack is valid.
func (n PidMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// DeviceMapping represents the device mapping between the host and the container.
type DeviceMapping struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}

// RestartPolicy represents the restart policies of the container.
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// IsNone indicates whether the container has the "no" restart policy.
// This means the container will not automatically restart when exiting.
func (rp *RestartPolicy) IsNone() bool {
	return rp.Name == "no"
}

// IsAlways indicates whether the container has the "always" restart policy.
// This means the container will automatically restart regardless of the exit status.
func (rp *RestartPolicy) IsAlways() bool {
	return rp.Name == "always"
}

// IsOnFailure indicates whether the container has the "on-failure" restart policy.
// This means the contain will automatically restart of exiting with a non-zero exit status.
func (rp *RestartPolicy) IsOnFailure() bool {
	return rp.Name == "on-failure"
}

// IsUnlessStopped indicates whether the container has the
// "unless-stopped" restart policy. This means the container will
// automatically restart unless user has put it to stopped state.
func (rp *RestartPolicy) IsUnlessStopped() bool {
	return rp.Name == "unless-stopped"
}

// LogConfig represents the logging configuration of the container.
type LogConfig struct {
	Type   string
	Config map[string]string
}

// Resources contains container's resources (cgroups config, ulimits...)
type Resources struct {
	// Applicable to all platforms
	CPUShares int64 `json:"CpuShares"` // CPU shares (relative weight vs. other containers)

	// Applicable to UNIX platforms
	CgroupParent      string // Parent cgroup.
	BlkioWeight       uint16 // Block IO weight (relative weight vs. other containers)
	BlkioWeightDevice []*blkiodev.WeightDevice
	CPUPeriod         int64            `json:"CpuPeriod"` // CPU CFS (Completely Fair Scheduler) period
	CPUQuota          int64            `json:"CpuQuota"`  // CPU CFS (Completely Fair Scheduler) quota
	CpusetCpus        string           // CpusetCpus 0-2, 0,1
	CpusetMems        string           // CpusetMems 0-2, 0,1
	Devices           []DeviceMapping  // List of devices to map inside the container
	KernelMemory      int64            // Kernel memory limit (in bytes)
	Memory            int64            // Memory limit (in bytes)
	MemoryReservation int64            // Memory soft limit (in bytes)
	MemorySwap        int64            // Total memory usage (memory + swap); set `-1` to disable swap
	MemorySwappiness  *int64           // Tuning container memory swappiness behaviour
	Ulimits           []*ulimit.Ulimit // List of ulimits to be set in the container
}

// HostConfig the non-portable Config structure of a container.
// Here, "non-portable" means "dependent of the host we are running on".
// Portable information *should* appear in Config.
type HostConfig struct {
	// Applicable to all platforms
	Binds           []string      // List of volume bindings for this container
	ContainerIDFile string        // File (path) where the containerId is written
	LogConfig       LogConfig     // Configuration of the logs for this container
	NetworkMode     NetworkMode   // Network mode to use for the container
	PortBindings    nat.PortMap   // Port mapping between the exposed port (container) and the host
	RestartPolicy   RestartPolicy // Restart policy to be used for the container
	VolumeDriver    string        // Name of the volume driver used to mount volumes
	VolumesFrom     []string      // List of volumes to take from other container

	// Applicable to UNIX platforms
	CapAdd          *stringutils.StrSlice // List of kernel capabilities to add to the container
	CapDrop         *stringutils.StrSlice // List of kernel capabilities to remove from the container
	DNS             []string              `json:"Dns"`        // List of DNS server to lookup
	DNSOptions      []string              `json:"DnsOptions"` // List of DNSOption to look for
	DNSSearch       []string              `json:"DnsSearch"`  // List of DNSSearch to look for
	ExtraHosts      []string              // List of extra hosts
	GroupAdd        []string              // List of additional groups that the container process will run as
	IpcMode         IpcMode               // IPC namespace to use for the container
	Links           []string              // List of links (in the name:alias form)
	OomKillDisable  bool                  // Whether to disable OOM Killer or not
	PidMode         PidMode               // PID namespace to use for the container
	Privileged      bool                  // Is the container in privileged mode
	PublishAllPorts bool                  // Should docker publish all exposed port for the container
	ReadonlyRootfs  bool                  // Is the container root filesystem in read-only
	SecurityOpt     []string              // List of string values to customize labels for MLS systems, such as SELinux.
	UTSMode         UTSMode               // UTS namespace to use for the container
	ShmSize         *int64                // Total shm memory usage

	// Applicable to Windows
	ConsoleSize [2]int         // Initial console size
	Isolation   IsolationLevel // Isolation level of the container (eg default, hyperv)

	// Contains container's resources (cgroups, ulimits)
	Resources
}

// DecodeHostConfig creates a HostConfig based on the specified Reader.
// It assumes the content of the reader will be JSON, and decodes it.
func DecodeHostConfig(src io.Reader) (*HostConfig, error) {
	decoder := json.NewDecoder(src)

	var w ContainerConfigWrapper
	if err := decoder.Decode(&w); err != nil {
		return nil, err
	}

	hc := w.getHostConfig()
	return hc, nil
}

// SetDefaultNetModeIfBlank changes the NetworkMode in a HostConfig structure
// to default if it is not populated. This ensures backwards compatibility after
// the validation of the network mode was moved from the docker CLI to the
// docker daemon.
func SetDefaultNetModeIfBlank(hc *HostConfig) *HostConfig {
	if hc != nil {
		if hc.NetworkMode == NetworkMode("") {
			hc.NetworkMode = NetworkMode("default")
		}
	}
	return hc
}
