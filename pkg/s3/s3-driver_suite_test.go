package s3_test

import (
	"io/ioutil"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ctrox/csi-s3/pkg/s3"
	"github.com/kubernetes-csi/csi-test/pkg/sanity"
)

var _ = Describe("S3Driver", func() {
	mntDir, _ := ioutil.TempDir("", "mnt")
	stagingDir, _ := ioutil.TempDir("", "staging")

	AfterSuite(func() {
		os.RemoveAll(mntDir)
		os.RemoveAll(stagingDir)
	})

	Context("goofys", func() {
		socket := "/tmp/csi-goofys.sock"
		csiEndpoint := "unix://" + socket
		cfg := &s3.Config{
			AccessKeyID:     "FJDSJ",
			SecretAccessKey: "DSG643HGDS",
			Endpoint:        "http://127.0.0.1:9000",
			Mounter:         "goofys",
		}
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint, cfg)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  mntDir,
				StagingPath: stagingDir,
				Address:     csiEndpoint,
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3fs", func() {
		socket := "/tmp/csi-s3fs.sock"
		csiEndpoint := "unix://" + socket
		cfg := &s3.Config{
			AccessKeyID:     "FJDSJ",
			SecretAccessKey: "DSG643HGDS",
			Endpoint:        "http://127.0.0.1:9000",
			Mounter:         "s3fs",
		}
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint, cfg)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  mntDir,
				StagingPath: stagingDir,
				Address:     csiEndpoint,
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3ql", func() {
		socket := "/tmp/csi-s3ql.sock"
		csiEndpoint := "unix://" + socket

		cfg := &s3.Config{
			AccessKeyID:     "FJDSJ",
			SecretAccessKey: "DSG643HGDS",
			Endpoint:        "http://127.0.0.1:9000",
			Mounter:         "s3ql",
		}
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint, cfg)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		defer os.RemoveAll(mntDir)

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  mntDir,
				StagingPath: stagingDir,
				Address:     csiEndpoint,
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("s3backer", func() {
		socket := "/tmp/csi-s3backer.sock"
		csiEndpoint := "unix://" + socket

		cfg := &s3.Config{
			AccessKeyID:     "FJDSJ",
			SecretAccessKey: "DSG643HGDS",
			Endpoint:        "http://127.0.0.1:9000",
			Mounter:         "s3backer",
		}
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		// Clear loop device so we cover the creation of it
		os.Remove(s3.S3backerLoopDevice)
		driver, err := s3.NewS3("test-node", csiEndpoint, cfg)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  mntDir,
				StagingPath: stagingDir,
				Address:     csiEndpoint,
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})

	Context("rclone", func() {
		socket := "/tmp/csi-rclone.sock"
		csiEndpoint := "unix://" + socket

		cfg := &s3.Config{
			AccessKeyID:     "FJDSJ",
			SecretAccessKey: "DSG643HGDS",
			Endpoint:        "http://127.0.0.1:9000",
			Mounter:         "rclone",
		}
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		driver, err := s3.NewS3("test-node", csiEndpoint, cfg)
		if err != nil {
			log.Fatal(err)
		}
		go driver.Run()

		Describe("CSI sanity", func() {
			sanityCfg := &sanity.Config{
				TargetPath:  mntDir,
				StagingPath: stagingDir,
				Address:     csiEndpoint,
			}
			sanity.GinkgoTest(sanityCfg)
		})
	})
})
