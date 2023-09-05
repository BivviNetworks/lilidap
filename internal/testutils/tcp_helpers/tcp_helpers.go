package tcp_helpers

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// Find a free port on the local host to use for starting a server
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// wait for a port to open on a host
func WaitForPort(t *testing.T, serverAddress string, port int) {
	t.Log("waitForPort begins")
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverAddress, port))
		if err == nil {
			t.Log("waitForPort success")
			conn.Close()
			break
		}
		t.Logf("waitForPort must keep waiting: %v", err)
		time.Sleep(100 * time.Millisecond)
	}
}
