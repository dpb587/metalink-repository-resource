package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink-repository-resource/factory"
	"github.com/dpb587/metalink/verification"
	"github.com/dpb587/metalink/verification/hash"
	"github.com/pkg/errors"
)

func main() {
	err := os.Chdir(os.Args[1])
	if err != nil {
		api.Fatal("out: bad args: source dir", err)
	}

	var request Request

	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal("out: bad stdin: parse error", err)
	}

	var metalinkPath string

	if len(request.Params.Files) > 0 {
		metalinkPath, err = createMetalink(request)
		if err != nil {
			api.Fatal("out: create metalink", err)
		}

		defer os.Remove(metalinkPath)
	} else {
		metalinkPaths, err := filepath.Glob(filepath.Join(os.Args[1], request.Params.Metalink))
		if err != nil {
			api.Fatal("out: bad metalink: globbing path", err)
		} else if len(metalinkPaths) == 0 {
			api.Fatal("out: bad metalink", errors.New("path not found"))
		} else if len(metalinkPaths) > 1 {
			api.Fatal("out: bad metalink", errors.New("multiple paths found when one is expected"))
		}

		metalinkPath = metalinkPaths[0]
	}

	metalinkBytes, err := ioutil.ReadFile(metalinkPath)
	if err != nil {
		api.Fatal("out: bad metalink: read error", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", metalinkBytes)

	meta4 := metalink.Metalink{}

	err = metalink.Unmarshal(metalinkBytes, &meta4)
	if err != nil {
		api.Fatal("out: bad metalink: parse error", err)
	}

	if len(meta4.Files) == 0 {
		api.Fatal("out: bad metalink: content error", errors.New("missing file node"))
	} else if meta4.Files[0].Version == "" {
		api.Fatal("out: bad metalink: content error", errors.New("missing file version node"))
	}

	metalinkFile, err := os.OpenFile(metalinkPath, os.O_RDONLY, 0700)
	if err != nil {
		api.Fatal("out: version file: create", err)
	}

	defer metalinkFile.Close()

	var metalinkName string

	if request.Params.Rename != "" {
		metalinkName = request.Params.Rename
	} else {
		metalinkName = fmt.Sprintf("v%s.meta4", meta4.Files[0].Version)
	}

	options := request.Source.Options

	for k, v := range request.Params.Options {
		options[k] = v
	}

	repository, err := factory.GetSource(request.Source.URI, options)
	if err != nil {
		api.Fatal("out: bad stdin: source: uri", err)
	}

	err = repository.Put(metalinkName, metalinkFile)
	if err != nil {
		api.Fatal("out: storing metalink", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(Response{Version: api.Version{Version: meta4.Files[0].Version}})
	if err != nil {
		api.Fatal("out: bad stdout: json", err)
	}
}

func createMetalink(request Request) (string, error) {
	now := time.Now()
	meta4 := metalink.Metalink{
		Generator: "metalink-repository-resource/0.0.0",
		Published: &now,
	}

	versionBytes, err := ioutil.ReadFile(filepath.Join(os.Args[1], request.Params.Version))
	if err != nil {
		return "", errors.Wrap(err, "reading version")
	}

	version := strings.TrimSpace(string(versionBytes))

	for _, paramFile := range request.Params.Files {
		filePaths, err := filepath.Glob(filepath.Join(os.Args[1], paramFile))
		if err != nil {
			return "", errors.Wrap(err, "globbing path")
		}

		for _, filePath := range filePaths {
			file := metalink.File{
				Name:    path.Base(filePath),
				Version: version,
				Hashes:  []metalink.Hash{},
			}

			fileTmplVars := struct {
				Name, Version, SHA1, SHA256, SHA512, MD5 string
			}{
				Name:    file.Name,
				Version: file.Version,
			}

			local, err := factory.GetOrigin(metalink.URL{URL: fmt.Sprintf("file://%s", filePath)})
			if err != nil {
				return "", errors.Wrap(err, "loading local file")
			}

			file.Size, err = local.Size()
			if err != nil {
				return "", errors.Wrap(err, "getting size")
			}

			hashmap := []verification.Signer{
				hash.SHA512Verification,
				hash.SHA256Verification,
				hash.SHA1Verification,
				hash.MD5Verification,
			}

			for _, hasher := range hashmap {
				verification, err := hasher.Sign(local)
				if err != nil {
					return "", errors.Wrap(err, "building hash")
				}

				err = verification.Apply(&file)
				if err != nil {
					return "", errors.Wrap(err, "adding hash")
				}
			}

			for _, hash := range file.Hashes {
				switch hash.Type {
				case "sha-512":
					fileTmplVars.SHA512 = hash.Hash
				case "sha-256":
					fileTmplVars.SHA256 = hash.Hash
				case "sha-1":
					fileTmplVars.SHA1 = hash.Hash
				case "md5":
					fileTmplVars.MD5 = hash.Hash
				}
			}

			for _, uploadParams := range request.Source.MirrorFiles {
				remoteURLTmpl, err := template.New("remote").Parse(uploadParams.Destination)
				if err != nil {
					return "", errors.Wrap(err, "parsing upload destination")
				}

				remoteURLBytes := &bytes.Buffer{}

				err = remoteURLTmpl.Execute(remoteURLBytes, fileTmplVars)
				if err != nil {
					return "", errors.Wrap(err, "generating upload destination")
				}

				for k, v := range uploadParams.Env {
					// TODO unset/revert after?
					os.Setenv(k, v)
				}

				fmt.Fprintf(os.Stderr, "uploading to %s\n", remoteURLBytes.String())

				remote, err := factory.GetOrigin(metalink.URL{URL: remoteURLBytes.String()})
				if err != nil {
					return "", errors.Wrap(err, "loading upload destination")
				}

				uri := remote.ReaderURI()

				progress := pb.New64(int64(file.Size)).Set(pb.Bytes, true).SetRefreshRate(time.Second).SetWidth(80)
				progress.Start()

				err = remote.WriteFrom(local, progress)
				if err != nil {
					return "", errors.Wrap(err, "uploading")
				}

				progress.Finish()

				file.URLs = append(
					file.URLs,
					metalink.URL{
						Location: uploadParams.Location,
						Priority: uploadParams.Priority,
						URL:      uri,
					},
				)
			}

			meta4.Files = append(meta4.Files, file)
		}
	}

	meta4Bytes, err := metalink.MarshalXML(meta4)
	if err != nil {
		return "", errors.Wrap(err, "marshaling metalink")
	}

	tmpfile, err := ioutil.TempFile("", "metalink-repository")
	if err != nil {
		return "", errors.Wrap(err, "creating temp file")
	}

	_, err = tmpfile.Write(meta4Bytes)
	if err != nil {
		os.Remove(tmpfile.Name())

		return "", errors.Wrap(err, "writing metalink")
	}

	return tmpfile.Name(), nil
}
