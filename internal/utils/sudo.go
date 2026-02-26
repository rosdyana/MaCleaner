// Package utils provides sudo management functionality
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// SudoManager handles sudo authentication and keeps it alive
type SudoManager struct {
	mu       sync.Mutex
	HasSudo  bool
	LastAuth time.Time
	Timeout  time.Duration
}

// NewSudoManager creates a new SudoManager
func NewSudoManager() *SudoManager {
	return &SudoManager{
		Timeout: 5 * time.Minute, // Sudo timeout is typically 5 minutes
	}
}

// EnsureSudo ensures sudo credentials are valid, prompting for password if needed
func (s *SudoManager) EnsureSudo() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we have valid sudo
	if s.HasSudo && time.Since(s.LastAuth) < s.Timeout {
		// Refresh sudo timestamp to extend timeout
		cmd := exec.Command("sudo", "-n", "true")
		if err := cmd.Run(); err == nil {
			s.LastAuth = time.Now()
			return nil
		}
	}

	// Need to authenticate - this will prompt for password once
	fmt.Println("\n🔐 Some operations require administrator privileges.")
	fmt.Println("   Please enter your password (will be cached for 5 minutes):")

	cmd := exec.Command("sudo", "-v") // Validate credentials
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		s.HasSudo = false
		return fmt.Errorf("sudo authentication failed")
	}

	s.HasSudo = true
	s.LastAuth = time.Now()

	// Start background refresh to keep sudo alive
	go s.keepAlive()

	return nil
}

// keepAlive runs in background to refresh sudo timestamp
func (s *SudoManager) keepAlive() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if !s.HasSudo {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		cmd := exec.Command("sudo", "-n", "true")
		if err := cmd.Run(); err != nil {
			// Sudo expired
			s.mu.Lock()
			s.HasSudo = false
			s.mu.Unlock()
			return
		}

		s.mu.Lock()
		s.LastAuth = time.Now()
		s.mu.Unlock()
	}
}

// Run runs a command with sudo if needed
func (s *SudoManager) Run(args ...string) error {
	if err := s.EnsureSudo(); err != nil {
		return err
	}
	cmd := exec.Command("sudo", args...)
	return cmd.Run()
}

// RunWithOutput runs a command with sudo and returns the output
func (s *SudoManager) RunWithOutput(args ...string) ([]byte, error) {
	if err := s.EnsureSudo(); err != nil {
		return nil, err
	}
	cmd := exec.Command("sudo", args...)
	return cmd.Output()
}
