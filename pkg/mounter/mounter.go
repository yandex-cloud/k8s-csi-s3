package mounter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/ctrox/csi-s3/pkg/s3"
	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"k8s.io/kubernetes/pkg/util/mount"
)

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Stage(stagePath string) error
	Unstage(stagePath string) error
	Mount(source string, target string) error
}

const (
	s3fsMounterType     = "s3fs"
	geesefsMounterType  = "geesefs"
	rcloneMounterType   = "rclone"
	TypeKey             = "mounter"
	BucketKey           = "bucket"
	OptionsKey          = "options"
)

// New returns a new mounter depending on the mounterType parameter
func New(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	mounter := meta.Mounter
	// Fall back to mounterType in cfg
	if len(meta.Mounter) == 0 {
		mounter = cfg.Mounter
	}
	switch mounter {
	case geesefsMounterType:
		return newGeeseFSMounter(meta, cfg)

	case s3fsMounterType:
		return newS3fsMounter(meta, cfg)

	case rcloneMounterType:
		return newRcloneMounter(meta, cfg)

	default:
		// default to GeeseFS
		return newGeeseFSMounter(meta, cfg)
	}
}

func fuseMount(path string, command string, args []string) error {
	cmd := exec.Command(command, args...)
	glog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error fuseMount command: %s\nargs: %s\noutput: %s", command, args, out)
	}

	return waitForMount(path, 10*time.Second)
}

func FuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := findFuseMountProcess(path)
	if err != nil {
		glog.Errorf("Error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		glog.Warningf("Unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	glog.Infof("Found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return waitForProcess(process, 20)
}

func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.New("").IsNotMountPoint(path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("Timeout waiting for mount")
		}
	}
}

func findFuseMountProcess(path string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		cmdLine, err := getCmdLine(p.Pid())
		if err != nil {
			glog.Errorf("Unable to get cmdline of PID %v: %s", p.Pid(), err)
			continue
		}
		if strings.Contains(cmdLine, path) {
			glog.Infof("Found matching pid %v on path %s", p.Pid(), path)
			return os.FindProcess(p.Pid())
		}
	}
	return nil, nil
}

func waitForProcess(p *os.Process, limit int) error {
	for backoff := 0; backoff < limit; backoff++ {
		cmdLine, err := getCmdLine(p.Pid)
		if err != nil {
			glog.Warningf("Error checking cmdline of PID %v, assuming it is dead: %s", p.Pid, err)
			p.Wait()
			return nil
		}
		if cmdLine == "" {
			glog.Warning("Fuse process seems dead, returning")
			p.Wait()
			return nil
		}
		if err := p.Signal(syscall.Signal(0)); err != nil {
			glog.Warningf("Fuse process does not seem active or we are unprivileged: %s", err)
			p.Wait()
			return nil
		}
		glog.Infof("Fuse process with PID %v still active, waiting...", p.Pid)
		time.Sleep(time.Duration(math.Pow(1.5, float64(backoff))*100) * time.Millisecond)
	}
	p.Release()
	return fmt.Errorf("Timeout waiting for PID %v to end", p.Pid)
}

func getCmdLine(pid int) (string, error) {
	cmdLineFile := fmt.Sprintf("/proc/%v/cmdline", pid)
	cmdLine, err := ioutil.ReadFile(cmdLineFile)
	if err != nil {
		return "", err
	}
	return string(cmdLine), nil
}

func createLoopDevice(device string) error {
	if _, err := os.Stat(device); !os.IsNotExist(err) {
		return nil
	}
	args := []string{
		device,
		"b", "7", "0",
	}
	cmd := exec.Command("mknod", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error configuring loop device: %s", out)
	}
	return nil
}
