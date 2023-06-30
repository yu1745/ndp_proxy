package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/yu1745/ndp_proxy/pkg/C"
	"github.com/yu1745/ndp_proxy/pkg/ndp"
	"github.com/yu1745/ndp_proxy/pkg/utils"
)

func init() {
	flag.StringVar(&C.Prefix, "p", "", "ipv6 prefix")
	flag.StringVar(&C.Interface, "i", "", "interface")
	flag.Parse()
	if C.Prefix == "" || C.Interface == "" {
		flag.Usage()
		os.Exit(0)
	}

	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	ip, ipNet, err := net.ParseCIDR(C.Prefix)
	if err != nil {
		panic(err)
	}
	ipNet.IP = ip
	cmd := fmt.Sprintf("ip r add local %s dev lo", ipNet.String())
	err = utils.ExecCmdWithSpaces(cmd)
	if err != nil {
		if err.Error() == "exit status 2" {
			// 条目已存在，忽略错误
		} else {
			panic(err)
		}
	}
	ndp.Listen(*ipNet, C.Interface)
}
