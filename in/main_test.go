package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var tmpDir, repositoryDir, storageDir, inDir string

	runCLI := func(stdin string) map[string]interface{} {
		command := exec.Command(cli, inDir)
		command.Stdin = bytes.NewBufferString(stdin)

		stdout := &bytes.Buffer{}

		session, err := gexec.Start(command, stdout, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		session.Wait(time.Minute)

		var result map[string]interface{}

		err = json.Unmarshal(stdout.Bytes(), &result)
		Expect(err).NotTo(HaveOccurred())

		return result
	}

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "metalink-repository-resource-in")
		Expect(err).NotTo(HaveOccurred())

		repositoryDir = path.Join(tmpDir, "repository")
		err = os.MkdirAll(repositoryDir, 0700)
		Expect(err).NotTo(HaveOccurred())

		storageDir = path.Join(tmpDir, "storage")
		err = os.MkdirAll(storageDir, 0700)
		Expect(err).NotTo(HaveOccurred())

		inDir = path.Join(tmpDir, "in")
		err = os.MkdirAll(inDir, 0700)
		Expect(err).NotTo(HaveOccurred())

		for _, dir := range []string{"repository", "storage", "in"} {
			os.MkdirAll(path.Join(tmpDir, dir), 0700)
		}

		var stubFiles = map[string]string{
			"repository/v0.1.0.meta4": fmt.Sprintf(`<metalink xmlns="urn:ietf:params:xml:ns:metalink">
  <file name="a-first.txt">
    <hash type="sha-512">b97213406d0d6848f87d20770cffa2405cb85468939efea99b5f2e7154b15381add67cc62fa2d2871c352ce4ef381c75424cd2ff1e27d4a02fc7910ad29e5b00</hash>
    <size>12</size>
    <url>file://%s/storage/a-first.txt</url>
    <version>0.1.0</version>
  </file>
	<file name="a-second.txt">
    <hash type="sha-512">5d30fb44a9bfaf535153e494387876bf48dcc9a62594c07abf122310a3045f7275f5856c091bcf62cf0cc7a1c9653689a090b55d99f83829c70e4550ef04ae11</hash>
    <size>13</size>
    <url>file://%s/storage/a-second.txt</url>
    <version>0.1.0</version>
  </file>
	<file name="a-third.txt">
    <hash type="sha-512">778c89626df1145ffaeae041363efd4fd0fedf5c76e5f01dfc902ff80a6ff9341f539c3a5ba0460d6f8830c3c8736ba60b56dd5f6cfd5be3a04270a73c82d276</hash>
    <size>12</size>
    <url>file://%s/storage/a-third.txt</url>
    <version>0.1.0</version>
  </file>
</metalink>`, tmpDir, tmpDir, tmpDir),
			"storage/a-first.txt":  "a first file",
			"storage/a-second.txt": "a second file",
			"storage/a-third.txt":  "a third file",
		}

		for stubPath, stubData := range stubFiles {
			err = ioutil.WriteFile(filepath.Join(tmpDir, stubPath), []byte(stubData), 0700)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		if tmpDir != "" {
			Expect(os.RemoveAll(tmpDir)).NotTo(HaveOccurred())
		}
	})

	It("downloads all files", func() {
		result := runCLI(fmt.Sprintf(`{
	"source": {
		"uri": "file://%s"
	},
	"version": {
		"version": "0.1.0"
	}
}`, repositoryDir))
		Expect(result["version"].(map[string]interface{})["version"]).To(Equal("0.1.0"))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "files")))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "bytes")))

		knownFiles, err := filepath.Glob(filepath.Join(inDir, "*"))
		Expect(err).NotTo(HaveOccurred())
		Expect(knownFiles).To(ConsistOf(
			filepath.Join(inDir, ".resource"),
			filepath.Join(inDir, "a-first.txt"),
			filepath.Join(inDir, "a-second.txt"),
			filepath.Join(inDir, "a-third.txt"),
		))

		storageBytes, err := ioutil.ReadFile(filepath.Join(inDir, "a-first.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a first file")))

		storageBytes, err = ioutil.ReadFile(filepath.Join(inDir, "a-second.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a second file")))

		storageBytes, err = ioutil.ReadFile(filepath.Join(inDir, "a-third.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a third file")))
	})

	It("respects source.include_files", func() {
		result := runCLI(fmt.Sprintf(`{
	"source": {
		"uri": "file://%s",
		"include_files": [
			"*d.txt"
		]
	},
	"version": {
		"version": "0.1.0"
	}
}`, repositoryDir))
		Expect(result["version"].(map[string]interface{})["version"]).To(Equal("0.1.0"))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "files")))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "bytes")))

		knownFiles, err := filepath.Glob(filepath.Join(inDir, "*"))
		Expect(err).NotTo(HaveOccurred())
		Expect(knownFiles).To(ConsistOf(
			filepath.Join(inDir, ".resource"),
			filepath.Join(inDir, "a-second.txt"),
			filepath.Join(inDir, "a-third.txt"),
		))

		storageBytes, err := ioutil.ReadFile(filepath.Join(inDir, "a-second.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a second file")))

		storageBytes, err = ioutil.ReadFile(filepath.Join(inDir, "a-third.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a third file")))
	})

	It("respects source.include_files + params.include_files", func() {
		result := runCLI(fmt.Sprintf(`{
	"source": {
		"uri": "file://%s",
		"include_files": [
			"*d.txt"
		]
	},
	"params": {
		"include_files": [
			"*s*"
		]
	},
	"version": {
		"version": "0.1.0"
	}
}`, repositoryDir))
		Expect(result["version"].(map[string]interface{})["version"]).To(Equal("0.1.0"))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "files")))
		Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "bytes")))

		knownFiles, err := filepath.Glob(filepath.Join(inDir, "*"))
		Expect(err).NotTo(HaveOccurred())
		Expect(knownFiles).To(ConsistOf(
			filepath.Join(inDir, ".resource"),
			filepath.Join(inDir, "a-second.txt"),
		))

		storageBytes, err := ioutil.ReadFile(filepath.Join(inDir, "a-second.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(storageBytes).To(Equal([]byte("a second file")))
	})
})
