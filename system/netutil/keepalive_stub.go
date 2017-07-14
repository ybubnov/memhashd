// +build !linux

package netutil

import (
	"net"
)

func setKeepalive(conn *net.TCPConn, count, idle, timeout int) error {
	return nil
}
