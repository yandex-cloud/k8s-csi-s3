package mounter

import (
	"testing"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

func TestGeeseFSAuthEnvs_StaticCreds(t *testing.T) {
	m := &geesefsMounter{
		accessKeyID:     "AKID",
		secretAccessKey: "SECRET",
		useIRSA:         false,
	}
	envs := m.authEnvs()
	if len(envs) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(envs))
	}
	if envs[0] != "AWS_ACCESS_KEY_ID=AKID" {
		t.Errorf("unexpected env[0]: %s", envs[0])
	}
	if envs[1] != "AWS_SECRET_ACCESS_KEY=SECRET" {
		t.Errorf("unexpected env[1]: %s", envs[1])
	}
}

func TestGeeseFSAuthEnvs_IRSA(t *testing.T) {
	m := &geesefsMounter{
		useIRSA:              true,
		roleArn:              "arn:aws:iam::123456789012:role/test",
		webIdentityTokenFile: "/var/run/secrets/token",
	}
	envs := m.authEnvs()
	if len(envs) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(envs))
	}
	if envs[0] != "AWS_ROLE_ARN=arn:aws:iam::123456789012:role/test" {
		t.Errorf("unexpected env[0]: %s", envs[0])
	}
	if envs[1] != "AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/token" {
		t.Errorf("unexpected env[1]: %s", envs[1])
	}
}

func TestGeeseFSAuthEnvs_IRSAFromEnv(t *testing.T) {
	// When roleArn/tokenFile are empty, IRSA relies on pod-injected env vars.
	// The mounter should pass no extra env vars (the process inherits them).
	m := &geesefsMounter{
		useIRSA: true,
	}
	envs := m.authEnvs()
	if len(envs) != 0 {
		t.Fatalf("expected 0 env vars when IRSA fields are empty, got %d: %v", len(envs), envs)
	}
}

func TestNewGeeseFSMounter_IRSA(t *testing.T) {
	meta := &s3.FSMeta{BucketName: "test-bucket"}
	cfg := &s3.Config{
		Endpoint:             "https://s3.amazonaws.com",
		Region:               "us-east-1",
		UseIRSA:              true,
		RoleArn:              "arn:aws:iam::123456789012:role/test",
		WebIdentityTokenFile: "/var/run/secrets/token",
	}
	m, err := newGeeseFSMounter(meta, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gm := m.(*geesefsMounter)
	if !gm.useIRSA {
		t.Error("expected useIRSA=true")
	}
	if gm.roleArn != cfg.RoleArn {
		t.Errorf("unexpected roleArn: %s", gm.roleArn)
	}
}

func TestNewRcloneMounter_IRSA(t *testing.T) {
	meta := &s3.FSMeta{BucketName: "test-bucket"}
	cfg := &s3.Config{
		Endpoint:             "https://s3.amazonaws.com",
		Region:               "us-east-1",
		UseIRSA:              true,
		RoleArn:              "arn:aws:iam::123456789012:role/test",
		WebIdentityTokenFile: "/var/run/secrets/token",
	}
	m, err := newRcloneMounter(meta, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rm := m.(*rcloneMounter)
	if !rm.useIRSA {
		t.Error("expected useIRSA=true")
	}
}

func TestNewS3fsMounter_IRSA(t *testing.T) {
	meta := &s3.FSMeta{BucketName: "test-bucket"}
	cfg := &s3.Config{
		Endpoint:             "https://s3.amazonaws.com",
		Region:               "us-east-1",
		UseIRSA:              true,
		RoleArn:              "arn:aws:iam::123456789012:role/test",
		WebIdentityTokenFile: "/var/run/secrets/token",
	}
	m, err := newS3fsMounter(meta, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sm := m.(*s3fsMounter)
	if !sm.useIRSA {
		t.Error("expected useIRSA=true")
	}
	if sm.pwFileContent != ":" {
		t.Errorf("expected empty pw content for IRSA, got %s", sm.pwFileContent)
	}
}
