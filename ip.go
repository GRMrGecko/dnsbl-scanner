package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

type IPAddr struct {
	ip        net.IP
	ipNet     *net.IPNet
	broadcast net.IP
}

// Calculate broadcast address for an provided IP network.
func getBroadcast(n *net.IPNet) net.IP {
	// Make new IP.
	l := len(n.IP)
	out := make(net.IP, l)

	// For each octet, mask with inverse of mask to get the broadcast.
	for i := 0; i < l; i++ {
		out[i] = n.IP[i] | ^n.Mask[i]
	}
	return out
}

// Parse an IP address with or without CIDR.
func ParseIPAddr(ip string) (*IPAddr, error) {
	var err error
	out := new(IPAddr)

	// If doesn't contain CIDR, do a normal IPv4 address parse.
	if !strings.Contains(ip, "/") {
		out.ip = net.ParseIP(ip)
		// If IP is nil, there was an parse error.
		if out.ip == nil {
			return nil, fmt.Errorf("ip address parse error")
		}
		// Convert to IPv4 space as we do not care about keeping IPv4 in IPv6.
		i4 := out.ip.To4()
		if i4 != nil {
			out.ip = i4
		}
		return out, nil
	}

	// Parse CIDR, and return error if error parsing.
	out.ip, out.ipNet, err = net.ParseCIDR(ip)
	if err != nil {
		return nil, err
	}

	// Convert to IPv4 space as we do not care about keeping IPv4 in IPv6.
	// This also makes broadcast calculation work better.
	i4 := out.ip.To4()
	if i4 != nil {
		out.ip = i4
	}
	i4 = out.ipNet.IP.To4()
	if i4 != nil {
		// Convert IPv6 mask to IPv4 mask as this is an IPv4 address.
		if len(out.ipNet.Mask) == net.IPv6len {
			out.ipNet.Mask = out.ipNet.Mask[12:]
		}
		out.ipNet.IP = i4
	}

	// Calculate broadcast.
	out.broadcast = getBroadcast(out.ipNet)
	return out, nil
}

// Return string represtation of IP address/CIDR.
func (n *IPAddr) String() string {
	if n.ipNet == nil {
		return n.ip.String()
	}
	return n.ipNet.String()
}

// Compare IP addreseses to see if the networks contain or is equal the IP address provided.
func (n *IPAddr) Contains(ip *IPAddr) bool {
	// If both are just IP addresses, do an equal operation.
	if n.ipNet == nil && ip.ipNet == nil {
		return n.ip.Equal(ip.ip)
	}

	// If this is an network, but provided is not, use ip network contains.
	if n.ipNet != nil && ip.ipNet == nil {
		return n.ipNet.Contains(ip.ip)
	}

	// If this is not an IP net, but provided address is... Return false.
	if n.ipNet == nil && ip.ipNet != nil {
		return false
	}

	// If provided network addr is within IP range of this IP network, return true.
	if bytes.Compare(ip.ipNet.IP, n.ipNet.IP) >= 0 && bytes.Compare(ip.ipNet.IP, n.broadcast) <= 0 {
		return true
	}

	// If provided broadcast addr is within IP range of this IP network, return true.
	if bytes.Compare(ip.broadcast, n.ipNet.IP) >= 0 && bytes.Compare(ip.broadcast, n.broadcast) <= 0 {
		return true
	}

	return false
}

// If IP addresses intercept.
func (n *IPAddr) Intercepts(ip *IPAddr) bool {
	// If this IP contains provided, it intercepts.
	if n.Contains(ip) {
		return true
	}
	// If provided IP contains this IP, it intercepts.
	return ip.Contains(n)
}
