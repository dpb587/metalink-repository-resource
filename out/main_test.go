package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/dpb587/metalink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	runCLI := func(stdin string) map[string]interface{} {
		command := exec.Command(cli, os.TempDir())
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

	var versionfile, metalinkfile, mirrorDir string

	BeforeEach(func() {
		version, err := ioutil.TempFile("", "metalink-repository-resource-version-file")
		Expect(err).NotTo(HaveOccurred())

		versionfile = version.Name()
		_, err = version.WriteString("2.1.0")
		Expect(err).NotTo(HaveOccurred())

		metalink, err := ioutil.TempFile("", "metalink-repository-resource-metalink-file")
		Expect(err).NotTo(HaveOccurred())

		metalinkfile = metalink.Name()
		_, err = metalink.WriteString(`{"files":[{"name":"fake-file1","version":"2.1.0"}]}`)
		Expect(err).NotTo(HaveOccurred())

		mirrorDir, err = ioutil.TempDir("", "metalink-repository-resource-mirror-dir")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if versionfile != "" {
			Expect(os.RemoveAll(versionfile)).NotTo(HaveOccurred())
		}

		if metalinkfile != "" {
			Expect(os.RemoveAll(metalinkfile)).NotTo(HaveOccurred())
		}

		if mirrorDir != "" {
			Expect(os.RemoveAll(mirrorDir)).NotTo(HaveOccurred())
		}
	})

	Describe("an existing metalink file", func() {
		It("adds the file to the repository", func() {
			result := runCLI(fmt.Sprintf(`{
		"source": {
			"uri": "file://%s/component",
			"branch": "master"
		},
		"params": {
			"metalink": "%s"
		}
	}`, repositorydir, metalinkfile))
			Expect(result["version"].(map[string]interface{})["version"]).To(Equal("2.1.0"))
			Expect(result["metadata"]).To(BeNil())

			By("committing the metalink", func() {
				meta4Bytes, err := ioutil.ReadFile(path.Join(repositorydir, "component/v2.1.0.meta4"))
				Expect(err).NotTo(HaveOccurred())

				var meta4 metalink.Metalink

				Expect(metalink.Unmarshal(meta4Bytes, &meta4)).NotTo(HaveOccurred())

				Expect(meta4.Files).To(HaveLen(1))
				Expect(meta4.Files[0].Version).To(Equal("2.1.0"))
			})
		})
	})

	Describe("generating metalinks", func() {
		var importFile1, importFile2 string

		BeforeEach(func() {
			version, err := ioutil.TempFile("", "metalink-repository-resource-import-file1")
			Expect(err).NotTo(HaveOccurred())

			importFile1 = version.Name()
			_, err = version.WriteString("a first file")
			Expect(err).NotTo(HaveOccurred())

			metalink, err := ioutil.TempFile("", "metalink-repository-resource-import-file2")
			Expect(err).NotTo(HaveOccurred())

			importFile2 = metalink.Name()
			_, err = metalink.WriteString("a second file")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if importFile1 != "" {
				Expect(os.RemoveAll(importFile1)).NotTo(HaveOccurred())
			}

			if importFile2 != "" {
				Expect(os.RemoveAll(importFile2)).NotTo(HaveOccurred())
			}
		})

		It("adds files", func() {
			result := runCLI(fmt.Sprintf(`{
		"source": {
			"uri": "file://%s/component",
			"branch": "master",
      "mirror_files": [
        {
          "destination": "file:///%s/{{.SHA1}}"
        }
      ]
		},
		"params": {
			"version": "%s",
			"files": [
        "%s",
        "%s"
      ]
		}
	}`, repositorydir, mirrorDir, versionfile, importFile1, importFile2))
			Expect(result["version"].(map[string]interface{})["version"]).To(Equal("2.1.0"))
			Expect(result["metadata"]).To(BeNil())

			By("mirroring files", func() {
				file1Bytes, err := ioutil.ReadFile(path.Join(mirrorDir, "70310a0bdf6e066479b091c0e5ad7e272d80fc8b"))
				Expect(err).NotTo(HaveOccurred())
				Expect(file1Bytes).To(Equal([]byte("a first file")))

				file2Bytes, err := ioutil.ReadFile(path.Join(mirrorDir, "0d18159152f12bc935b6293b6323f992e67140cc"))
				Expect(err).NotTo(HaveOccurred())
				Expect(file2Bytes).To(Equal([]byte("a second file")))
			})

			By("generating a metalink", func() {
				meta4Bytes, err := ioutil.ReadFile(path.Join(repositorydir, "component/v2.1.0.meta4"))
				Expect(err).NotTo(HaveOccurred())

				var meta4 metalink.Metalink

				Expect(metalink.Unmarshal(meta4Bytes, &meta4)).NotTo(HaveOccurred())

				Expect(meta4.Files).To(HaveLen(2))
				Expect(meta4.Files[0].Name).To(Equal(path.Base(importFile1)))
				Expect(meta4.Files[0].Hashes).To(HaveLen(4))
				Expect(meta4.Files[0].Hashes[0].Type).To(Equal(metalink.HashTypeSHA512))
				Expect(meta4.Files[0].Hashes[0].Hash).To(Equal("b97213406d0d6848f87d20770cffa2405cb85468939efea99b5f2e7154b15381add67cc62fa2d2871c352ce4ef381c75424cd2ff1e27d4a02fc7910ad29e5b00"))
				Expect(meta4.Files[0].URLs).To(HaveLen(1))
				Expect(meta4.Files[0].Size).To(Equal(uint64(12)))
				Expect(meta4.Files[0].Version).To(Equal("2.1.0"))
				Expect(meta4.Files[1].Name).To(Equal(path.Base(importFile2)))
				Expect(meta4.Files[1].Hashes).To(HaveLen(4))
				Expect(meta4.Files[1].Hashes[0].Type).To(Equal(metalink.HashTypeSHA512))
				Expect(meta4.Files[1].Hashes[0].Hash).To(Equal("5d30fb44a9bfaf535153e494387876bf48dcc9a62594c07abf122310a3045f7275f5856c091bcf62cf0cc7a1c9653689a090b55d99f83829c70e4550ef04ae11"))
				Expect(meta4.Files[1].Size).To(Equal(uint64(13)))
				Expect(meta4.Files[1].Version).To(Equal("2.1.0"))
				Expect(meta4.Files[1].URLs).To(HaveLen(1))
			})
		})
	})
})
