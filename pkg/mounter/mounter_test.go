package mounter

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForMountLoop_ImmediateMount(t *testing.T) {
	// Mount point exists immediately
	check := func(path string) (bool, error) {
		return false, nil // false = IS a mount point
	}
	err := waitForMountLoop("/mnt/test", 1*time.Second, check, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWaitForMountLoop_MountAppearsAfterDelay(t *testing.T) {
	var calls int32
	check := func(path string) (bool, error) {
		n := atomic.AddInt32(&calls, 1)
		if n >= 3 {
			return false, nil // mounted after 3 checks
		}
		return true, nil // not mounted yet
	}
	err := waitForMountLoop("/mnt/test", 1*time.Second, check, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if atomic.LoadInt32(&calls) < 3 {
		t.Errorf("expected at least 3 mount checks, got %d", calls)
	}
}

func TestWaitForMountLoop_Timeout(t *testing.T) {
	check := func(path string) (bool, error) {
		return true, nil // never mounts
	}
	start := time.Now()
	err := waitForMountLoop("/mnt/test", 50*time.Millisecond, check, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "Timeout waiting for mount") {
		t.Errorf("expected timeout message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "/mnt/test") {
		t.Errorf("expected path in error message, got: %v", err)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("returned too quickly: %v", elapsed)
	}
}

func TestWaitForMountLoop_UnitFailed_FailsFast(t *testing.T) {
	check := func(path string) (bool, error) {
		return true, nil // never mounts
	}
	getUnitState := func() (string, string, error) {
		return "failed", "exit-code", nil
	}
	start := time.Now()
	err := waitForMountLoop("/mnt/test", 5*time.Second, check, getUnitState)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error when unit fails")
	}
	if !strings.Contains(err.Error(), `entered state "failed"`) {
		t.Errorf("expected failed state in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "exit-code") {
		t.Errorf("expected result in error, got: %v", err)
	}
	// Should fail fast, not wait for the full 5s timeout
	if elapsed > 1*time.Second {
		t.Errorf("should have failed fast, but took %v", elapsed)
	}
}

func TestWaitForMountLoop_UnitInactive_FailsFast(t *testing.T) {
	check := func(path string) (bool, error) {
		return true, nil
	}
	getUnitState := func() (string, string, error) {
		return "inactive", "resources", nil
	}
	err := waitForMountLoop("/mnt/test", 5*time.Second, check, getUnitState)
	if err == nil {
		t.Fatal("expected error when unit is inactive")
	}
	if !strings.Contains(err.Error(), `entered state "inactive"`) {
		t.Errorf("expected inactive state in error, got: %v", err)
	}
}

func TestWaitForMountLoop_UnitFailsAfterDelay(t *testing.T) {
	// Simulates: unit starts as "activating" then transitions to "failed"
	var checks int32
	check := func(path string) (bool, error) {
		atomic.AddInt32(&checks, 1)
		return true, nil
	}
	getUnitState := func() (string, string, error) {
		n := atomic.LoadInt32(&checks)
		if n >= 3 {
			return "failed", "signal", nil
		}
		return "activating", "", nil
	}
	err := waitForMountLoop("/mnt/test", 5*time.Second, check, getUnitState)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `entered state "failed"`) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWaitForMountLoop_UnitActiveAndMounts(t *testing.T) {
	// Unit stays active, mount eventually appears — success path
	var calls int32
	check := func(path string) (bool, error) {
		n := atomic.AddInt32(&calls, 1)
		if n >= 3 {
			return false, nil
		}
		return true, nil
	}
	getUnitState := func() (string, string, error) {
		return "active", "", nil
	}
	err := waitForMountLoop("/mnt/test", 5*time.Second, check, getUnitState)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWaitForMountLoop_UnitStateErrorIgnored(t *testing.T) {
	// If we can't query unit state, we should fall through to timeout
	// (not crash or fail early)
	check := func(path string) (bool, error) {
		return true, nil
	}
	getUnitState := func() (string, string, error) {
		return "", "", fmt.Errorf("dbus connection lost")
	}
	err := waitForMountLoop("/mnt/test", 50*time.Millisecond, check, getUnitState)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "Timeout") {
		t.Errorf("expected timeout, got: %v", err)
	}
}

func TestWaitForMountLoop_NilUnitState(t *testing.T) {
	// Without a unit state checker, behaves like original waitForMount
	check := func(path string) (bool, error) {
		return true, nil
	}
	err := waitForMountLoop("/mnt/test", 50*time.Millisecond, check, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "Timeout") {
		t.Errorf("expected timeout, got: %v", err)
	}
}

func TestWaitForMountLoop_MountCheckError(t *testing.T) {
	check := func(path string) (bool, error) {
		return false, fmt.Errorf("stat /mnt/test: no such file or directory")
	}
	err := waitForMountLoop("/mnt/test", 1*time.Second, check, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("expected stat error, got: %v", err)
	}
}
