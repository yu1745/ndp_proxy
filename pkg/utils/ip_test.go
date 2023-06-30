package utils

import (
	"net"
	"testing"
)

func TestIpNetToString(t *testing.T) {
	ip, ipNet, _ := net.ParseCIDR("fd80::1/64")
	ipNet.IP = ip
	type args struct {
		ipNet *net.IPNet
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"",
			args{ipNet},
			"fd80::1/64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IpNetToString(tt.args.ipNet); got != tt.want {
				t.Errorf("IpNetToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
