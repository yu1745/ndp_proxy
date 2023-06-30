package utils

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/yu1745/ndp_proxy/pkg/C"
)

var fc00 *net.IPNet
var ff00 *net.IPNet
var localhost net.IP
var zero net.IP

func init() {
	_, fc00, _ = net.ParseCIDR("fc00::/7")
	_, ff00, _ = net.ParseCIDR("ff00::/8")
	localhost = net.ParseIP("::1")
	zero = net.ParseIP("::")
}

func IsGlobalIPV6(ipNet *net.IPNet) bool {
	return ipNet.IP.To4() == nil && ipNet.IP.IsGlobalUnicast() && !fc00.Contains(ipNet.IP) && !ff00.Contains(ipNet.IP) && !localhost.Equal(ipNet.IP) && !zero.Equal(ipNet.IP)
}

func GetAllAvailableIpNet() ([]net.IPNet, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var ipNets []net.IPNet
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if IsGlobalIPV6(ipNet) {
				ipNets = append(ipNets, *ipNet)
			}
		}
	}
	return ipNets, nil
}

func GetLinkLocalIPByInterfaceName(name string) (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Name == name {
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}

			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				// Check if it's a link-local IP address
				if ok && ipNet.IP.IsLinkLocalUnicast() {
					return ipNet.IP, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("interface not found: %s", name)
}

func GetMacByInterfaceName(name string) (net.HardwareAddr, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	return iface.HardwareAddr, nil
}

func IpNetToString(ipNet *net.IPNet) string {
	ones, _ := ipNet.Mask.Size()
	return ipNet.IP.String() + "/" + strconv.Itoa(ones)
}

func SGetMacByInterfaceName(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		log.Fatal(err)
	}
	return iface.HardwareAddr.String()
}

func GetMyIpWithSuffixIp(i int) {
	_, ipNet, err := net.ParseCIDR(C.Prefix)
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个自定义的net.Dialer
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(ipNet.IP.String() + fmt.Sprintf("%04d", i)), // 指定本地IP地址
		},
	}
	// 创建一个自定义的Transport，并设置DialContext字段为自定义的net.Dialer
	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

	// 创建一个自定义的http.Client，并设置Transport字段为自定义的Transport
	client := &http.Client{
		Transport: transport,
	}

	// 创建HTTP请求
	req, err := http.NewRequest("GET", "http://myip6.ipip.net", nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send request:", err)
		return
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response:", err)
		return
	}
	println(string(all))
}
