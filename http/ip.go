
package http

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var cidrs []*net.IPNet

func init() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, err := net.ParseCIDR(maxCidrBlock)
		if err != nil {
			panic(fmt.Sprintf("failed to parse CIDR block %q: %v", maxCidrBlock, err))
		}
		cidrs[i] = cidr
	}
}

// IPFromRequest return client's real public IP address from http request headers.
func IPFromRequest(r *http.Request) string {
	// If we have it, return this first.
	//
	// https://developers.cloudflare.com/fundamentals/get-started/reference/http-request-headers/#cf-connecting-ip
	if ip := r.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip
	}

	// If we have it, try to return the first global address in X-Forwarded-For
	for _, ip := range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		isPrivate, err := isPrivateAddress(ip)
		if !isPrivate && err == nil {
			return ip
		}
	}

	// Check X-Real-Ip header next
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}

	// If all else fails, return the remote address
	//
	// If there are colon in remote address, remove the port number
	// otherwise, return remote address as is
	var ip string
	if strings.ContainsRune(r.RemoteAddr, ':') {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr) //nolint:errcheck
	} else {
		ip = r.RemoteAddr
	}
	return ip
}

// isPrivateAddress works by checking if the address is under private CIDR blocks.
// List of private CIDR blocks can be seen on :
//
// https://en.wikipedia.org/wiki/Private_network
//
// https://en.wikipedia.org/wiki/Link-local_address
func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}
