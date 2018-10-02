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
	metalinktemplate "github.com/dpb587/metalink/template"
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

	api.MigrateSource(&request.Source)

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
	} else if request.Params.RenameFromFile != "" {
		var metalinkNameBytes []byte
		metalinkNameBytes, err = ioutil.ReadFile(request.Params.RenameFromFile)
		if err != nil {
			api.Fatal("out: could not open rename file: read error", err)
		}
		metalinkName = string(metalinkNameBytes)
	} else {
		metalinkName = "v{{ .Version }}.meta4"
	}

	metalinkNameTmpl, err := template.New("metalink").Parse(metalinkName)
	if err != nil {
		api.Fatal("out: parsing metalink name", err)
	}

	metalinkNameBytes := &bytes.Buffer{}

	err = metalinkNameTmpl.Execute(metalinkNameBytes, map[string]string{
		"Version": meta4.Files[0].Version,
	})
	if err != nil {
		api.Fatal("out: executing metalink name template", err)
	}

	metalinkName = metalinkNameBytes.String()

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

	urlLoader := factory.GetURLLoader(request.Source.URLHandlers)

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

			local, err := urlLoader.LoadURL(metalink.URL{URL: fmt.Sprintf("file://%s", filePath)})
			if err != nil {
				return "", errors.Wrap(err, "loading local file")
			}

			file.Size, err = local.Size()
			if err != nil {
				return "", errors.Wrap(err, "getting size")
			}

			hashmap := []verification.Signer{
				hash.SHA512SignerVerifier,
				hash.SHA256SignerVerifier,
				hash.SHA1SignerVerifier,
				hash.MD5SignerVerifier,
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

			for _, uploadParams := range request.Source.MirrorFiles {
				remoteURLTmpl, err := metalinktemplate.New(uploadParams.Destination)
				if err != nil {
					return "", errors.Wrap(err, "parsing upload destination")
				}

				remoteURL, err := remoteURLTmpl.ExecuteString(file)
				if err != nil {
					return "", errors.Wrap(err, "generating upload destination")
				}

				var uri string
				var uploadError error

				for retry := 1; retry <= 3; retry ++ {
					uploadError = nil

					if retry > 1 {
						fmt.Fprintf(os.Stderr, "\nretrying (attempt #%d)...\n", retry)
					}

					fmt.Fprintf(os.Stderr, "uploading to %s\n", remoteURL)

					remote, err := urlLoader.LoadURL(metalink.URL{URL: remoteURL})
					if err != nil {
						return "", errors.Wrap(err, "loading upload destination")
					}

					uri = remote.ReaderURI()

					progress := pb.New64(int64(file.Size)).Set(pb.Bytes, true).SetRefreshRate(time.Second).SetWidth(80)
					progress.Start()

					err = remote.WriteFrom(local, progress)
					progress.Finish()

					if err != nil {
						fmt.Fprintf(os.Stderr, "uploading failed: %v\n", err)

						uploadError = err

						continue
					}

					break
				}

				if uploadError != nil {
					return "", errors.Wrap(uploadError, "uploading")
				}

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
