package driver

import "testing"

func TestVolumeBucketPrefix(t *testing.T) {
	tests := []struct {
		name          string
		volumeID      string
		volumeContext map[string]string
		wantBucket    string
		wantPrefix    string
	}{
		{"static PV, bucket in volumeHandle", "bucket1", nil, "bucket1", ""},
		{"static PV, bucket and prefix in volumeHandle", "bucket1/dir", nil, "bucket1", "dir"},
		{"static PV, bucket attribute", "pv-1", map[string]string{"bucket": "bucket1"}, "bucket1", ""},
		{"static PV, bucket and prefix attributes", "pv-1", map[string]string{"bucket": "bucket1", "prefix": "dir"}, "bucket1", "dir"},
		{"static PV, bucket attribute matches volumeHandle", "bucket1", map[string]string{"bucket": "bucket1"}, "bucket1", ""},
		{"dynamic PV in a StorageClass with bucket", "bucket1/pvc-123", map[string]string{"bucket": "bucket1"}, "bucket1", "pvc-123"},
		{"empty bucket attribute is ignored", "bucket1/dir", map[string]string{"bucket": ""}, "bucket1", "dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, prefix := volumeBucketPrefix(tt.volumeID, tt.volumeContext)
			if bucket != tt.wantBucket || prefix != tt.wantPrefix {
				t.Errorf("volumeBucketPrefix(%q, %v) = (%q, %q), want (%q, %q)",
					tt.volumeID, tt.volumeContext, bucket, prefix, tt.wantBucket, tt.wantPrefix)
			}
		})
	}
}
