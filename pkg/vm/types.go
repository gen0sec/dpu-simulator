package vm

import (
	"fmt"
	"os"
	"strings"

	"libvirt.org/go/libvirt"

	"github.com/ovn-kubernetes/dpu-simulator/pkg/config"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/log"
	"github.com/ovn-kubernetes/dpu-simulator/pkg/platform"
)

// VMManager manages libvirt virtual machines and networks
type VMManager struct {
	conn       *libvirt.Connect
	config     *config.Config
	hostDistro *platform.Distro
	// hostExec runs tools on the libvirt host (qemu-img, ovs-vsctl, wget, etc.).
	hostExec platform.CommandExecutor
	// hostSpec is the resolved per-host virtualization profile (machine type,
	// emulator path, firmware, and feature flags) reused across VM XML creation.
	hostSpec archSpec
}

// NewVMManager creates a new VMManager with the given config, connecting to libvirt.
// hostExec must be non-nil; it runs host-local commands (qemu-img, ovs-vsctl, etc.).
// Callers typically use platform.NewLocalExecutor(); tests may pass a stub.
func NewVMManager(cfg *config.Config, hostExec platform.CommandExecutor) (*VMManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if hostExec == nil {
		return nil, fmt.Errorf("hostExec is nil")
	}

	distro, err := platform.GetHostDistro()
	if err != nil {
		return nil, fmt.Errorf("failed to detect host distro: %w", err)
	}
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	hostname, err := conn.GetHostname()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	log.Debug("✓ Connected to libvirt: %s", hostname)

	// Target arch defaults to the host arch (KVM-accelerated). DPU_SIM_TARGET_ARCH
	// overrides it to emulate a different arch under TCG — e.g. an aarch64
	// BlueField-style DPU on an x86_64 host.
	targetArch := distro.Architecture
	tcg := false
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("DPU_SIM_TARGET_ARCH"))); v != "" {
		switch v {
		case "aarch64", "arm64":
			targetArch = platform.AARCH64
		case "x86_64", "amd64":
			targetArch = platform.X86_64
		default:
			conn.Close()
			return nil, fmt.Errorf("invalid DPU_SIM_TARGET_ARCH %q (want aarch64|x86_64)", v)
		}
		tcg = targetArch != distro.Architecture
		if tcg {
			log.Info("⚠ DPU_SIM_TARGET_ARCH=%s differs from host %s → TCG software emulation (slow)", targetArch, distro.Architecture)
		}
	}

	hostSpec, err := hostArchSpec(targetArch, tcg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to determine host virtualization settings: %w", err)
	}

	return &VMManager{
		conn:       conn,
		config:     cfg,
		hostDistro: distro,
		hostExec:   hostExec,
		hostSpec:   hostSpec,
	}, nil
}

// Close closes the libvirt connection
func (m *VMManager) Close() error {
	if m.conn != nil {
		_, err := m.conn.Close()
		return err
	}
	return nil
}

// VMState represents the state of a virtual machine
type VMState int

const (
	VMStateUnknown VMState = iota
	VMStateRunning
	VMStateBlocked
	VMStatePaused
	VMStateShutdown
	VMStateShutoff
	VMStateCrashed
)

// String returns string representation of VM state
func (s VMState) String() string {
	switch s {
	case VMStateRunning:
		return "Running"
	case VMStateBlocked:
		return "Blocked"
	case VMStatePaused:
		return "Paused"
	case VMStateShutdown:
		return "Shutdown"
	case VMStateShutoff:
		return "Shut off"
	case VMStateCrashed:
		return "Crashed"
	default:
		return "Unknown"
	}
}

// InterfaceInfo represents VM interface information
type InterfaceInfo struct {
	Name   string
	Hwaddr string
	Addrs  []string
}

// VMInfo represents comprehensive VM information
type VMInfo struct {
	Name      string
	State     VMState
	IP        string
	VCPUs     uint
	MemoryMB  uint64
	MaxMemory uint64
}
