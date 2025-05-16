package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	// Load .env if present (no error if missing)
	_ = godotenv.Load()

	go func() {
		// Defaults (for Docker Compose)
		hostName := "primordia"
		port := 14000

		// Check environment variables
		if v := os.Getenv("GAME_HOST"); v != "" {
			hostName = v
		}
		if v := os.Getenv("GAME_PORT"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				port = parsed
			} else {
				fmt.Printf("⚠️ GAME_PORT '%s' is not a valid integer, using default %d\n", v, port)
			}
		}

		timeout := 2 * time.Second
		waitDelay := 2 * time.Second
		maxAttempts := 10

		fmt.Printf("🔎 Will attempt to connect to game at %s:%d\n", hostName, port)
		if err := TryConnectWithRetries(hostName, port, timeout, waitDelay, maxAttempts); err != nil {
			fmt.Println("🛑", err)
		} else {
			fmt.Println("🎉 Connected to Primordia TCP server! Ready to proceed with AI logic!")
		}
	}()
}
