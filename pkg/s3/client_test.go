package s3

import (
	"os"
	"testing"
)

func TestNewClientFromSecret_StaticCreds(t *testing.T) {
	secret := map[string]string{
		"accessKeyID":     "AKID",
		"secretAccessKey": "SECRET",
		"region":          "us-east-1",
		"endpoint":        "https://s3.amazonaws.com",
	}
	client, err := NewClientFromSecret(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Config.AccessKeyID != "AKID" {
		t.Errorf("expected AccessKeyID=AKID, got %s", client.Config.AccessKeyID)
	}
	if client.Config.SecretAccessKey != "SECRET" {
		t.Errorf("expected SecretAccessKey=SECRET, got %s", client.Config.SecretAccessKey)
	}
	if client.Config.UseIRSA {
		t.Error("expected UseIRSA=false")
	}
}

func TestNewClientFromSecret_IRSA(t *testing.T) {
	// Set env vars that the IAM credential provider will look for
	os.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789012:role/test")
	os.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", "/var/run/secrets/token")
	defer os.Unsetenv("AWS_ROLE_ARN")
	defer os.Unsetenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	secret := map[string]string{
		"region":   "us-west-2",
		"endpoint": "https://s3.us-west-2.amazonaws.com",
		"useIRSA":  "true",
	}
	client, err := NewClientFromSecret(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.Config.UseIRSA {
		t.Error("expected UseIRSA=true")
	}
	if client.Config.AccessKeyID != "" {
		t.Errorf("expected empty AccessKeyID, got %s", client.Config.AccessKeyID)
	}
}

func TestNewClientFromSecret_IRSAWithExplicitRole(t *testing.T) {
	secret := map[string]string{
		"region":               "eu-west-1",
		"endpoint":             "https://s3.eu-west-1.amazonaws.com",
		"useIRSA":              "true",
		"roleArn":              "arn:aws:iam::999999999999:role/custom",
		"webIdentityTokenFile": "/custom/token/path",
	}
	client, err := NewClientFromSecret(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.Config.UseIRSA {
		t.Error("expected UseIRSA=true")
	}
	if client.Config.RoleArn != "arn:aws:iam::999999999999:role/custom" {
		t.Errorf("unexpected RoleArn: %s", client.Config.RoleArn)
	}
	if client.Config.WebIdentityTokenFile != "/custom/token/path" {
		t.Errorf("unexpected WebIdentityTokenFile: %s", client.Config.WebIdentityTokenFile)
	}
}

func TestNewClientFromSecret_Insecure(t *testing.T) {
	secret := map[string]string{
		"accessKeyID":     "AKID",
		"secretAccessKey": "SECRET",
		"endpoint":        "https://minio.local:9000",
		"insecure":        "true",
	}
	client, err := NewClientFromSecret(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.Config.Insecure {
		t.Error("expected Insecure=true")
	}
}

func TestNewClientFromSecret_UseIRSAFalseByDefault(t *testing.T) {
	secret := map[string]string{
		"accessKeyID":     "AKID",
		"secretAccessKey": "SECRET",
		"endpoint":        "https://s3.amazonaws.com",
	}
	client, err := NewClientFromSecret(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Config.UseIRSA {
		t.Error("UseIRSA should default to false when not specified")
	}
}
