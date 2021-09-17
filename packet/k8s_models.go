package packet

import (
	"net"
	"time"

	"github.com/tinkerbell/boots/k8s/api/v1alpha1"
)

func tinkOsToDiscoveryOS(in *v1alpha1.Metadata_Instance_OperatingSystem) *OperatingSystem {
	if in == nil {
		return nil
	}

	return &OperatingSystem{
		Slug:     in.Slug,
		Distro:   in.Distro,
		Version:  in.Version,
		ImageTag: in.ImageTag,
		OsSlug:   in.OsSlug,
	}
}

func tinkIpToDiscoveryIp(in *v1alpha1.Metadata_Instance_IP) *IP {
	if in == nil {
		return nil
	}
	return &IP{
		Address:    net.ParseIP(in.Address),
		Netmask:    net.ParseIP(in.Netmask),
		Gateway:    net.ParseIP(in.Gateway),
		Family:     int(in.Family),
		Public:     in.Public,
		Management: in.Management,
	}
}

func (d *K8sDiscovery) Instance() *Instance {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Instance != nil {
		return &Instance{
			ID:       d.hw.Spec.Metadata.Instance.Id,
			State:    InstanceState(d.hw.Spec.Metadata.Instance.State),
			Hostname: d.hw.Spec.Metadata.Instance.Hostname,
			AllowPXE: d.hw.Spec.Metadata.Instance.AllowPxe,
			Rescue:   d.hw.Spec.Metadata.Instance.Rescue,
			OS:       tinkOsToDiscoveryOS(d.hw.Spec.Metadata.Instance.OperatingSystem),
			// OSV:           nil,
			AlwaysPXE:     d.hw.Spec.Metadata.Instance.AlwaysPxe,
			IPXEScriptURL: d.hw.Spec.Metadata.Instance.IpxeScriptUrl,
			IPs: func(in []*v1alpha1.Metadata_Instance_IP) []IP {
				resp := []IP{}
				for _, ip := range in {
					resp = append(resp, *tinkIpToDiscoveryIp(ip))
				}
				return resp
			}(d.hw.Spec.Metadata.Instance.Ips),
			UserData: d.hw.Spec.Metadata.Instance.Userdata,
			// servicesVersion:     ServicesVersion{},
			CryptedRootPassword: d.hw.Spec.Metadata.Instance.CryptedRootPassword,
			// PasswordHash:        "",
			Tags:         d.hw.Spec.Metadata.Instance.Tags,
			SSHKeys:      d.hw.Spec.Metadata.Instance.SshKeys,
			NetworkReady: d.hw.Spec.Metadata.Instance.NetworkReady,
			// BootDriveHint:       "",
		}
	}
	return nil
}
func (d *K8sDiscovery) MAC() net.HardwareAddr {
	if len(d.hw.Spec.Interfaces) > 0 && d.hw.Spec.Interfaces[0].DHCP != nil {
		mac, err := net.ParseMAC(d.hw.Spec.Interfaces[0].DHCP.MAC)
		if err != nil {
			return nil
		}
		return mac
	}
	return nil
}

func (d *K8sDiscovery) Mode() string {
	return "hardware"
}

func (d *K8sDiscovery) GetIP(addr net.HardwareAddr) IP {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil && iface.DHCP.MAC != "" && iface.DHCP.IP != nil {
			if addr.String() == iface.DHCP.MAC {
				return IP{
					Address: net.ParseIP(iface.DHCP.IP.Address),
					Netmask: net.ParseIP(iface.DHCP.IP.Netmask),
					Gateway: net.ParseIP(iface.DHCP.IP.Gateway),
					Family:  int(iface.DHCP.IP.Family),
					// TODO not 100% accurate
					Public: !net.ParseIP(iface.DHCP.IP.Address).IsPrivate(),
					// TODO: When should we set this to true?
					Management: false,
				}
			}
		}
	}
	return IP{}
}

func (d *K8sDiscovery) GetMAC(ip net.IP) net.HardwareAddr {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil && iface.DHCP.MAC != "" && iface.DHCP.IP != nil {
			if ip.String() == iface.DHCP.IP.Address {
				mac, err := net.ParseMAC(iface.DHCP.MAC)
				if err != nil {
					return nil
				}
				return mac
			}
		}
	}
	return nil
}

func (d *K8sDiscovery) DnsServers(mac net.HardwareAddr) []net.IP {
	resp := []net.IP{}
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil && iface.DHCP.MAC != "" {
			for _, ns := range iface.DHCP.NameServers {
				resp = append(resp, net.ParseIP(ns))
			}
		}
	}
	return resp
}

func (d *K8sDiscovery) LeaseTime(mac net.HardwareAddr) time.Duration {
	if len(d.hw.Spec.Interfaces) > 0 && d.hw.Spec.Interfaces[0].DHCP != nil {
		return time.Duration(d.hw.Spec.Interfaces[0].DHCP.LeaseTime) * time.Second
	}
	// Default to 24 hours?
	return time.Hour * 24
}

func (d *K8sDiscovery) Hostname() (string, error) {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil && iface.DHCP.Hostname != "" {
			return iface.DHCP.Hostname, nil
		}
	}
	return "", nil
}

func (d *K8sDiscovery) Hardware() Hardware {
	return d
}

func (d *K8sDiscovery) SetMAC(mac net.HardwareAddr) {
	return
}

func NewK8sDiscovery(hw *v1alpha1.Hardware) Discovery {
	return &K8sDiscovery{hw: hw}
}

type K8sDiscovery struct {
	hw *v1alpha1.Hardware
}

var _ Discovery = &K8sDiscovery{}
var _ Hardware = &K8sDiscovery{}

func (d *K8sDiscovery) HardwareAllowWorkflow(mac net.HardwareAddr) bool {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.Netboot != nil && iface.DHCP != nil && mac.String() == iface.DHCP.MAC {
			return *iface.Netboot.AllowWorkflow
		}
	}
	return false
}

func (d *K8sDiscovery) HardwareAllowPXE(mac net.HardwareAddr) bool {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.Netboot != nil && iface.DHCP != nil && mac.String() == iface.DHCP.MAC {
			return *iface.Netboot.AllowPXE
		}
	}
	return false
}

func (d *K8sDiscovery) HardwareArch(mac net.HardwareAddr) string {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil {
			return iface.DHCP.Arch
		}
	}
	return ""
}

func (d *K8sDiscovery) HardwareBondingMode() BondingMode {
	if d.hw.Spec.Metadata != nil {
		return BondingMode(d.hw.Spec.Metadata.BondingMode)
	}
	return BondingMode(0)
}

func (d *K8sDiscovery) HardwareFacilityCode() string {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Facility != nil {
		return d.hw.Spec.Metadata.Facility.FacilityCode
	}
	return ""
}

func (d *K8sDiscovery) HardwareID() HardwareID {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Instance != nil {
		return HardwareID(d.hw.Spec.Metadata.Instance.Id)
	}
	return HardwareID("")
}

func (d *K8sDiscovery) HardwareIPs() []IP {
	resp := []IP{}
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Instance != nil {
		for _, ip := range d.hw.Spec.Metadata.Instance.Ips {
			resp = append(resp, *tinkIpToDiscoveryIp(ip))
		}
	}
	return resp
}

func (d *K8sDiscovery) Interfaces() []Port {
	// TODO: to be updated {}
	return nil
}

func (d *K8sDiscovery) HardwareManufacturer() string {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Manufacturer != nil {
		return d.hw.Spec.Metadata.Manufacturer.Id
	}
	return ""
}

func (d *K8sDiscovery) HardwareProvisioner() string {
	return ""
}

func (d *K8sDiscovery) HardwarePlanSlug() string {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Facility != nil {
		return d.hw.Spec.Metadata.Facility.PlanSlug
	}
	return ""
}

func (d *K8sDiscovery) HardwarePlanVersionSlug() string {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Facility != nil {
		return d.hw.Spec.Metadata.Facility.PlanVersionSlug
	}
	return ""
}

func (d *K8sDiscovery) HardwareState() HardwareState {
	if d.hw.Spec.Metadata != nil {
		return HardwareState(d.hw.Spec.Metadata.State)
	}
	return ""
}

func (d *K8sDiscovery) HardwareOSIEVersion() string {
	return ""
}

func (d *K8sDiscovery) HardwareUEFI(mac net.HardwareAddr) bool {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.DHCP != nil {
			return iface.DHCP.UEFI
		}
	}
	return false
}

func (d *K8sDiscovery) OSIEBaseURL(mac net.HardwareAddr) string {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.Netboot != nil && iface.Netboot.OSIE != nil {
			return iface.Netboot.OSIE.BaseURL
		}
	}
	return ""
}

func (d *K8sDiscovery) KernelPath(mac net.HardwareAddr) string {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.Netboot != nil && iface.Netboot.OSIE != nil {
			return iface.Netboot.OSIE.Kernel
		}
	}
	return ""
}

func (d *K8sDiscovery) InitrdPath(mac net.HardwareAddr) string {
	for _, iface := range d.hw.Spec.Interfaces {
		if iface.Netboot != nil && iface.Netboot.OSIE != nil {
			return iface.Netboot.OSIE.Initrd
		}
	}
	return ""
}

func (d *K8sDiscovery) OperatingSystem() *OperatingSystem {
	if d.hw.Spec.Metadata != nil && d.hw.Spec.Metadata.Instance != nil && d.hw.Spec.Metadata.Instance.OperatingSystem != nil {
		return &OperatingSystem{
			Slug:     d.hw.Spec.Metadata.Instance.OperatingSystem.Slug,
			Distro:   d.hw.Spec.Metadata.Instance.OperatingSystem.Distro,
			Version:  d.hw.Spec.Metadata.Instance.OperatingSystem.Version,
			ImageTag: d.hw.Spec.Metadata.Instance.OperatingSystem.ImageTag,
			OsSlug:   d.hw.Spec.Metadata.Instance.OperatingSystem.OsSlug,
			// Installer:     "",
			// InstallerData: &InstallerData{},
		}
	}
	return nil
}
