package main

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestNextSubnet(t *testing.T) {
	tests := []struct {
		name    string
		cidr    net.IPNet
		want    *net.IPNet
		wantErr bool
	}{
		{
			cidr: *MustParseCidr("172.18.0.0/16"),
			want: MustParseCidr("172.18.128.0/24"),
		},
		{
			cidr: *MustParseCidr("172.18.0.0/30"),
			want: MustParseCidr("172.18.0.2/31"),
		},
		{
			cidr: net.IPNet{
				IP:   net.IPv4(10, 16, 0, 0),
				Mask: net.CIDRMask(12, 32),
			},
			want: MustParseCidr("10.24.0.0/22"),
		},
		{
			cidr: net.IPNet{
				IP:   net.IPv4(192, 168, 24, 0).To4(),
				Mask: net.CIDRMask(24, 32),
			},
			want: MustParseCidr("192.168.24.128/28"),
		},
		{
			cidr: *MustParseCidr("fc00:f853:ccd:e793::/64"),
			want: MustParseCidr("fc00:f853:ccd:e793:8000::/96"),
		},
		{
			cidr:    *MustParseCidr("127.0.0.1/32"),
			wantErr: true,
		},
		{
			cidr:    *MustParseCidr("fc00:f853:ccd:e793::/128"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("[%s]", &tt.cidr), func(t *testing.T) {
			got, err := NextSubnet(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextSubnet(%v) error = %v, wantErr %v", &tt.cidr, err, tt.wantErr)
				return
			}
			if !IPNetEqual(got, tt.want) {
				t.Errorf("NextSubnet(%v) = %v, want %v", &tt.cidr, got, tt.want)
			}
		})
	}
}

func IPNetEqual(a, b *net.IPNet) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil /* implies b == nil */ {
		return true
	}

	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func MustParseCidr(s string) *net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return cidr
}
