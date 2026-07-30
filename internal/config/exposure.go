package config

import (
	"fmt"
	"net"
	"strings"
)

// CheckExposure refuses a public address with no credential — a process that will not
// start beats a warning that scrolls past. Re-derived from kern-orch's and kern-ui's own
// version of this rule, not shared: no brick depends on another's internals.
func CheckExposure(addr, token string) error {
	if !isPublicAddr(addr) || token != "" {
		return nil
	}
	return fmt.Errorf(
		"refusing to listen on %s with no credential: set %s.\n"+
			"Anyone who can reach this address could read and resolve every document.\n"+
			"To run locally instead, leave %s at 127.0.0.1:7080",
		addr, EnvToken, EnvAddr)
}

// isPublicAddr reports whether addr can be reached from another machine. An empty host is
// the trap: `:7080` reads as innocent and binds every interface.
func isPublicAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.Trim(host, "[]")

	switch host {
	case "":
		return true
	case "localhost":
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsLoopback()
	}
	return true
}
