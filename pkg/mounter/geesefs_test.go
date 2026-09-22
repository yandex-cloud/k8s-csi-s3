package mounter

import (
	"testing"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

func TestValidateGeeseFSOptions(t *testing.T) {
	forbidden := []struct {
		name    string
		value   string
		boolean bool
	}{
		{"iam", "", true},
		{"iam-flavor", "imdsv1", false},
		{"iam-url", "http://127.0.0.1:8080", false},
		{"iam-header", "X-Test-Token", false},
		{"endpoint", "https://example.com", false},
		{"pprof", "127.0.0.1:6060", false},
		{"setuid", "0", false},
		{"setgid", "65534", false},
		{"log-file", "/csi/pwn.txt", false},
		{"shared-config", "/csi/credentials", false},
		{"cache", "/csi/cache", false},
		{"no-systemd", "", true},
		{"f", "", true},
	}

	checkRejected := func(t *testing.T, args []string) {
		t.Helper()
		if err := ValidateGeeseFSOptions(args); err == nil {
			t.Fatalf("expected rejection of %#v", args)
		}
	}

	for _, option := range forbidden {
		for _, prefix := range []string{"-", "--"} {
			flag := prefix + option.name
			t.Run(flag, func(t *testing.T) {
				if option.boolean {
					checkRejected(t, []string{flag})
					checkRejected(t, []string{flag + "=false"})
				} else {
					checkRejected(t, []string{flag, option.value})
					checkRejected(t, []string{flag + "=" + option.value})
				}
			})
		}
	}

	t.Run("allowed options are accepted", func(t *testing.T) {
		args := []string{
			"--no-checksum", "--memory-limit=4000",
			"--dir-mode", "0755", "--debug_fuse",
			"--cache-file-mode=0644",
			"--future-geesefs-option=anything",
			"--region=ru-central1",
			"--profile=default",
		}
		if err := ValidateGeeseFSOptions(args); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("following option is checked after equals form", func(t *testing.T) {
		checkRejected(t, []string{"--memory-limit=4000", "--setuid=0"})
	})
}

func TestGeeseFSMountRejectsUserOptionBeforeStarting(t *testing.T) {
	m := &geesefsMounter{
		meta: &s3.FSMeta{
			MountOptions: []string{"--memory-limit=4000", "--setuid=0"},
		},
	}

	err := m.Mount("/unused/mountpoint", "test-volume")
	want := `GeeseFS option "--setuid=0" is not allowed`
	if err == nil || err.Error() != want {
		t.Fatalf("Mount error = %v, want %q", err, want)
	}
}
