package s3

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"

	"github.com/golang/glog"
)

func waitForProcess(p *os.Process, backoff int) error {
	if backoff == 20 {
		return fmt.Errorf("Timeout waiting for PID %v to end", p.Pid)
	}
	cmdLine, err := getCmdLine(p.Pid)
	if err != nil {
		glog.Warningf("Error checking cmdline of PID %v, assuming it is dead: %s", p.Pid, err)
		return nil
	}
	if cmdLine == "" {
		// ignore defunct processes
		// TODO: debug why this happens in the first place
		// seems to only happen on k8s, not on local docker
		glog.Warning("Fuse process seems dead, returning")
		return nil
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		glog.Warningf("Fuse process does not seem active or we are unprivileged: %s", err)
		return nil
	}
	glog.Infof("Fuse process with PID %v still active, waiting...", p.Pid)
	time.Sleep(time.Duration(backoff*100) * time.Millisecond)
	return waitForProcess(p, backoff+1)
}

func findFuseMountProcess(path string, name string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		if strings.Contains(p.Executable(), name) {
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
	}
	return nil, nil
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
