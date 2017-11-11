package utils

import (
	"errors"
	"net"
)

// LocalAddrs returns the local address
// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func LocalAddrs() ([]net.Addr, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback
		}
		return iface.Addrs()
	}

	return nil, errors.New("not found")
}
