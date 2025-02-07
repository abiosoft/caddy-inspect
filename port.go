package inspect

import (
	"fmt"
	"net"
)

const httpServerListenPort = 2020

// findAvailablePort returns an available port on the host machine.
// it attempts port 2020 up till 2029
func findAvailablePort() (int, error) {
	allocatePort := func(port int) error {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return fmt.Errorf("error picking an available port: %w", err)
		}

		if err := listener.Close(); err != nil {
			return fmt.Errorf("error closing temporary port listener: %w", err)
		}

		return nil
	}

	var err error
	for i := 0; i < 10; i++ {
		port := httpServerListenPort + i
		err = allocatePort(port)
		if err == nil {
			return port, nil
		}
	}

	return 0, err
}
