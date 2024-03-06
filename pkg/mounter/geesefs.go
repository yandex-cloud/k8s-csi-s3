package mounter

import (
	"fmt"
	"os"
	"strings"
	"time"

	systemd "github.com/coreos/go-systemd/v22/dbus"
	dbus "github.com/godbus/dbus/v5"
	"github.com/golang/glog"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

const (
	geesefsCmd    = "geesefs"
)

// Implements Mounter
type geesefsMounter struct {
	meta            *s3.FSMeta
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGeeseFSMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &geesefsMounter{
		meta:            meta,
		endpoint:        cfg.Endpoint,
		region:          cfg.Region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (geesefs *geesefsMounter) CopyBinary(from, to string) error {
	st, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("Failed to stat %s: %v", from, err)
	}
	st2, err := os.Stat(to)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Failed to stat %s: %v", to, err)
	}
	if err != nil || st2.Size() != st.Size() || st2.ModTime() != st.ModTime() {
		if err == nil {
			// remove the file first to not hit "text file busy" errors
			err = os.Remove(to)
			if err != nil {
				return fmt.Errorf("Error removing %s to update it: %v", to, err)
			}
		}
		bin, err := os.ReadFile(from)
		if err != nil {
			return fmt.Errorf("Error copying %s to %s: %v", from, to, err)
		}
		err = os.WriteFile(to, bin, 0755)
		if err != nil {
			return fmt.Errorf("Error copying %s to %s: %v", from, to, err)
		}
		err = os.Chtimes(to, st.ModTime(), st.ModTime())
		if err != nil {
			return fmt.Errorf("Error copying %s to %s: %v", from, to, err)
		}
	}
	return nil
}

func (geesefs *geesefsMounter) MountDirect(target string, args []string) error {
	args = append([]string{
		"--endpoint", geesefs.endpoint,
		"-o", "allow_other",
		"--log-file", "/dev/stderr",
	}, args...)
	envs := []string{
		"AWS_ACCESS_KEY_ID=" + geesefs.accessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + geesefs.secretAccessKey,
	}
	return fuseMount(target, geesefsCmd, args, envs)
}

type execCmd struct {
	Path             string
	Args             []string
	UncleanIsFailure bool
}

func (geesefs *geesefsMounter) Mount(target, volumeID string) error {
	fullPath := fmt.Sprintf("%s:%s", geesefs.meta.BucketName, geesefs.meta.Prefix)
	var args []string
	if geesefs.region != "" {
		args = append(args, "--region", geesefs.region)
	}
	args = append(
		args,
		"--setuid", "65534", // nobody. drop root privileges
		"--setgid", "65534", // nogroup
	)
	useSystemd := true
	for i := 0; i < len(geesefs.meta.MountOptions); i++ {
		opt := geesefs.meta.MountOptions[i]
		if opt == "--no-systemd" {
			useSystemd = false
		} else if len(opt) > 0 && opt[0] == '-' {
			// Remove unsafe options
			s := 1
			if len(opt) > 1 && opt[1] == '-' {
				s++
			}
			key := opt[s:]
			e := strings.Index(opt, "=")
			if e >= 0 {
				key = opt[s:e]
			}
			if key == "log-file" || key == "shared-config" || key == "cache" {
				// Skip options accessing local FS
				if e < 0 {
					i++
				}
			} else if key != "" {
				args = append(args, opt)
			}
		} else if len(opt) > 0 {
			args = append(args, opt)
		}
	}
	args = append(args, fullPath, target)
	// Try to start geesefs using systemd so it doesn't get killed when the container exits
	if !useSystemd {
		return geesefs.MountDirect(target, args)
	}
	conn, err := systemd.New()
	if err != nil {
		glog.Errorf("Failed to connect to systemd dbus service: %v, starting geesefs directly", err)
		return geesefs.MountDirect(target, args)
	}
	defer conn.Close()
	// systemd is present
	if err = geesefs.CopyBinary("/usr/bin/geesefs", "/csi/geesefs"); err != nil {
		return err
	}
	pluginDir := os.Getenv("PLUGIN_DIR")
	if pluginDir == "" {
		pluginDir = "/var/lib/kubelet/plugins/ru.yandex.s3.csi"
	}
	args = append([]string{pluginDir+"/geesefs", "-f", "-o", "allow_other", "--endpoint", geesefs.endpoint}, args...)
	glog.Info("Starting geesefs using systemd: "+strings.Join(args, " "))
	unitName := "geesefs-"+systemd.PathBusEscape(volumeID)+".service"
	newProps := []systemd.Property{
		systemd.Property{
			Name: "Description",
			Value: dbus.MakeVariant("GeeseFS mount for Kubernetes volume "+volumeID),
		},
		systemd.PropExecStart(args, false),
		systemd.Property{
			Name: "Environment",
			Value: dbus.MakeVariant([]string{ "AWS_ACCESS_KEY_ID="+geesefs.accessKeyID, "AWS_SECRET_ACCESS_KEY="+geesefs.secretAccessKey }),
		},
		systemd.Property{
			Name: "CollectMode",
			Value: dbus.MakeVariant("inactive-or-failed"),
		},
	}
	unitProps, err := conn.GetAllProperties(unitName)
	if err == nil {
		// Unit already exists
		if s, ok := unitProps["ActiveState"].(string); ok && (s == "active" || s == "activating" || s == "reloading") {
			// Unit is already active
			curPath := ""
			prevExec, ok := unitProps["ExecStart"].([][]interface{})
			if ok && len(prevExec) > 0 && len(prevExec[0]) >= 2 {
				execArgs, ok := prevExec[0][1].([]string)
				if ok && len(execArgs) >= 2 {
					curPath = execArgs[len(execArgs)-1]
				}
			}
			if curPath != target {
				// FIXME This may mean that the same bucket&path are used for multiple PVs. Support it somehow
				return fmt.Errorf(
					"GeeseFS for volume %v is already mounted on host, but"+
					" in a different directory. We want %v, but it's in %v",
					volumeID, target, curPath,
				)
			}
			// Already mounted at right location
			return nil
		} else {
			// Stop and garbage collect the unit if automatic collection didn't work for some reason
			conn.StopUnit(unitName, "replace", nil)
			conn.ResetFailedUnit(unitName)
		}
	}
	err = os.MkdirAll("/run/systemd/system/" + unitName + ".d", 0755)
	if err == nil {
		// force & lazy unmount to cleanup possibly dead mountpoints
		err = os.WriteFile(
			"/run/systemd/system/" + unitName + ".d/50-ExecStopPost.conf",
			[]byte("[Service]\nExecStopPost=/bin/umount -f -l "+target+"\n"),
			0600,
		)
		if err == nil {
			_, err = conn.StartTransientUnit(unitName, "replace", newProps[0:len(newProps)-1], nil)
		}
	}
	if err != nil {
		return fmt.Errorf("Error starting systemd unit %s on host: %v", unitName, err)
	}
	return waitForMount(target, 10*time.Second)
}
