package s3_test

import (
	"log"
	"os"

	"github.com/ctrox/csi-s3/pkg/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

var _ = Describe("S3Driver", func() {

	Context("goofys", func() {
		socket := "/tmp/csi-goofys.sock"
		csiEndpoint := "unix://" + socket
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/goofys-target",
				StagingPath: os.TempDir() + "/goofys-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "goofys",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3fs", func() {
		socket := "/tmp/csi-s3fs.sock"
		csiEndpoint := "unix://" + socket
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/s3fs-target",
				StagingPath: os.TempDir() + "/s3fs-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "s3fs",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3backer", func() {
		socket := "/tmp/csi-s3backer.sock"
		csiEndpoint := "unix://" + socket

		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		// Clear loop device so we cover the creation of it
		os.Remove(s3.S3backerLoopDevice)
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/s3backer-target",
				StagingPath: os.TempDir() + "/s3backer-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "s3backer",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("rclone", func() {
		socket := "/tmp/csi-rclone.sock"
		csiEndpoint := "unix://" + socket

		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  os.TempDir() + "/rclone-target",
				StagingPath: os.TempDir() + "/rclone-staging",
				Address:     csiEndpoint,
				SecretsFile: "../../test/secret.yaml",
				TestVolumeParameters: map[string]string{
					"mounter": "rclone",
				},
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})
})
