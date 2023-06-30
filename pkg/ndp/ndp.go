package ndp

import (
	"log"
	"net"
	"os/exec"

	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/yu1745/ndp_proxy/pkg/utils"
)

type CustomPacketDataSource struct {
	fd int
}

func NewCustomPacketDataSource(fd int) *CustomPacketDataSource {
	return &CustomPacketDataSource{
		fd: fd,
	}
}

func (p *CustomPacketDataSource) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	buf := make([]byte, 1500)
	n, rAddr, err := syscall.Recvfrom(p.fd, buf, 0)
	if err != nil {
		return nil, gopacket.CaptureInfo{}, err
	}
	println(net.IP(rAddr.(*syscall.SockaddrInet6).Addr[:]).String())
	return buf[:n], gopacket.CaptureInfo{CaptureLength: n, Length: n}, nil
}

func Listen(ipNet net.IPNet, interfaceName string) {
	srcIP, err := utils.GetLinkLocalIPByInterfaceName(interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	dstMac, err := utils.GetMacByInterfaceName(interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	handle, err := pcap.OpenLive(interfaceName, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	filter := "icmp6 and ip6[40] == 135" // 135: Neighbor Solicitation
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		log.Println("receive " + packet.String())
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
		ns := packet.Layer(layers.LayerTypeICMPv6NeighborSolicitation).(*layers.ICMPv6NeighborSolicitation)

		if !ipNet.Contains(ns.TargetAddress) {
			log.Printf("prefix not contains %s\n", ns.TargetAddress.String())
			continue
		}
		ipLayer := packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)

		ethernetLayer_ := &layers.Ethernet{
			SrcMAC:       dstMac,
			DstMAC:       ethernetLayer.SrcMAC,
			EthernetType: layers.EthernetTypeIPv6,
		}

		ipLayer_ := &layers.IPv6{
			Version:    6,
			HopLimit:   255,
			SrcIP:      srcIP,
			DstIP:      ipLayer.SrcIP,
			NextHeader: layers.IPProtocolICMPv6,
		}

		icmpv6Layer_ := &layers.ICMPv6{
			TypeCode: layers.CreateICMPv6TypeCode(layers.ICMPv6TypeNeighborAdvertisement, 0),
		}
		icmpv6Layer_.SetNetworkLayerForChecksum(ipLayer_)
		NeighborAdvertisementLayer_ := &layers.ICMPv6NeighborAdvertisement{
			Flags:         0b01000000,
			TargetAddress: ns.TargetAddress,
			Options: layers.ICMPv6Options{
				layers.ICMPv6Option{
					Type: layers.ICMPv6OptTargetAddress,
					Data: []byte(dstMac),
				},
			},
		}
		buffer := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		}
		err = gopacket.SerializeLayers(buffer, opts, ethernetLayer_, ipLayer_, icmpv6Layer_, NeighborAdvertisementLayer_)
		if err != nil {
			log.Fatal(err)
		}
		handle.WritePacketData(buffer.Bytes())
		log.Println("send " + gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default).String())
	}
}

func allowNonLocalBind() error {
	err := exec.Command("sysctl", "net.ipv6.ip_nonlocal_bind=1").Run()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	err := allowNonLocalBind()
	if err != nil {
		log.Fatal(err)
	}
}
