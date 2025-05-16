package main

import (
	"fmt"
	"net"
	"time"
)

// TryConnect tests TCP connectivity to host:port with a timeout.
func TryConnect(host string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("connection to %s failed: %w", address, err)
	}
	defer conn.Close()
	return nil // success!
}

// TryConnectWithRetries attempts to connect to host:port up to maxAttempts, with waitDelay between tries.
func TryConnectWithRetries(host string, port int, timeout, waitDelay time.Duration, maxAttempts int) error {
	address := fmt.Sprintf("%s:%d", host, port)
	for i := 1; i <= maxAttempts; i++ {
		fmt.Printf("🔄 Attempt %d: Connecting to %s...\n", i, address)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err == nil {
			conn.Close()
			fmt.Println("✅ Connected to Primordia TCP server!")
			return nil
		}
		fmt.Printf("❌ Attempt %d failed: %v\n", i, err)
		time.Sleep(waitDelay)
	}
	return fmt.Errorf("failed to connect to %s after %d attempts", address, maxAttempts)
}

func TryToConnect() {
	// Start connection test in a goroutine (background)
	go func() {
		hostName := "primordia"
		port := 14000
		timeout := 2 * time.Second
		waitDelay := 2 * time.Second
		maxAttempts := 10

		if err := TryConnectWithRetries(hostName, port, timeout, waitDelay, maxAttempts); err != nil {
			fmt.Println("🛑", err)
		} else {
			fmt.Println("🎉 Connected to Primordia TCP server! Ready to proceed with AI logic!")
		}
	}()
}
