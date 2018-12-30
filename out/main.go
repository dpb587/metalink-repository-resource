package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	metalinktemplate "github.com/dpb587/metalink/template"
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
	var localCache = map[string]string{}

	if len(request.Params.Files) > 0 {
		metalinkPath, localCache, err = createMetalink(request)
		if err != nil {
			api.Fatal("out: create metalink", err)
		}

		defer os.Remove(metalinkPath)
	} else {
		metalinkPaths, err := filepath.Glob(request.Params.Metalink)
		if err != nil {
			api.Fatal("out: bad metalink: globbing path", err)
		} else if len(metalinkPaths) == 0 {
			api.Fatal("out: bad metalink", errors.New("path not found"))
		} else if len(metalinkPaths) > 1 {
			api.Fatal("out: bad metalink", errors.New("multiple paths found when one is expected"))
		}

		metalinkPath = metalinkPaths[0]
	}

	meta4Bytes, err := ioutil.ReadFile(metalinkPath)
	if err != nil {
		api.Fatal("out: bad metalink: read error", err)
	}

	meta4 := metalink.Metalink{}

	err = metalink.Unmarshal(meta4Bytes, &meta4)
	if err != nil {
		api.Fatal("out: bad metalink: parse error", err)
	}

	if len(meta4.Files) == 0 {
		api.Fatal("out: bad metalink: content error", errors.New("missing file node"))
	} else if meta4.Files[0].Version == "" {
		api.Fatal("out: bad metalink: content error", errors.New("missing file version node"))
	}

	var metalinkFile io.Reader

	if len(request.Source.MirrorFiles) > 0 {
		meta4, err = mirrorMetalink(request, meta4, localCache)
		if err != nil {
			api.Fatal("out: mirroring", err)
		}

		meta4Bytes, err := metalink.MarshalXML(meta4)
		if err != nil {
			api.Fatal("out: bad metalink: marshal error", err)
		}

		metalinkFile = bytes.NewBuffer(meta4Bytes)
	} else {
		// preserve original raw file if we're not modifying it at all
		metalinkFileReal, err := os.OpenFile(metalinkPath, os.O_RDONLY, 0700)
		if err != nil {
			api.Fatal("out: version file: create", err)
		}

		metalinkFile = metalinkFileReal

		defer metalinkFileReal.Close()
	}

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

func createMetalink(request Request) (string, map[string]string, error) {
	now := time.Now()
	meta4 := metalink.Metalink{
		Generator: "metalink-repository-resource/0.0.0",
		Published: &now,
	}
	localCache := map[string]string{}

	versionBytes, err := ioutil.ReadFile(request.Params.Version)
	if err != nil {
		return "", nil, errors.Wrap(err, "reading version")
	}

	version := strings.TrimSpace(string(versionBytes))

	urlLoader := factory.GetURLLoader(request.Source.URLHandlers)

	for _, paramFile := range request.Params.Files {
		filePaths, err := filepath.Glob(paramFile)
		if err != nil {
			return "", nil, errors.Wrap(err, "globbing path")
		}

		for _, filePath := range filePaths {
			file := metalink.File{
				Name:    path.Base(filePath),
				Version: version,
				Hashes:  []metalink.Hash{},
			}

			localCache[file.Name] = fmt.Sprintf("file://%s", filePath)

			local, err := urlLoader.LoadURL(metalink.URL{URL: localCache[file.Name]})
			if err != nil {
				return "", nil, errors.Wrap(err, "loading local file")
			}

			file.Size, err = local.Size()
			if err != nil {
				return "", nil, errors.Wrap(err, "getting size")
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
					return "", nil, errors.Wrap(err, "building hash")
				}

				err = verification.Apply(&file)
				if err != nil {
					return "", nil, errors.Wrap(err, "adding hash")
				}
			}

			meta4.Files = append(meta4.Files, file)
		}
	}

	meta4Bytes, err := metalink.MarshalXML(meta4)
	if err != nil {
		return "", nil, errors.Wrap(err, "marshaling metalink")
	}

	tmpfile, err := ioutil.TempFile("", "metalink-repository")
	if err != nil {
		return "", nil, errors.Wrap(err, "creating temp file")
	}

	_, err = tmpfile.Write(meta4Bytes)
	if err != nil {
		os.Remove(tmpfile.Name())

		return "", nil, errors.Wrap(err, "writing metalink")
	}

	return tmpfile.Name(), localCache, nil
}

func mirrorMetalink(request Request, meta4 metalink.Metalink, localCache map[string]string) (metalink.Metalink, error) {
	urlLoader := factory.GetURLLoader(request.Source.URLHandlers)

	for fileIdx, file := range meta4.Files {
		// TODO support multiple URLs
		localURI, isLocal := localCache[file.Name]
		if !isLocal {
			if len(file.URLs) < 1 {
				return meta4, errors.New("file is missing url")
			}

			localURI = file.URLs[0].URL
		}

		local, err := urlLoader.LoadURL(metalink.URL{URL: localURI})
		if err != nil {
			return meta4, errors.Wrap(err, "loading local file")
		}

		for _, uploadParams := range request.Source.MirrorFiles {
			remoteURLTmpl, err := metalinktemplate.New(uploadParams.Destination)
			if err != nil {
				return meta4, errors.Wrap(err, "parsing upload destination")
			}

			remoteURL, err := remoteURLTmpl.ExecuteString(file)
			if err != nil {
				return meta4, errors.Wrap(err, "generating upload destination")
			}

			var uri string
			var uploadError error

			for retry := 1; retry <= 3; retry++ {
				uploadError = nil

				if retry > 1 {
					fmt.Fprintf(os.Stderr, "\nretrying (attempt #%d)...\n", retry)
				}

				fmt.Fprintf(os.Stderr, "uploading to %s\n", remoteURL)

				remote, err := urlLoader.LoadURL(metalink.URL{URL: remoteURL})
				if err != nil {
					return meta4, errors.Wrap(err, "loading upload destination")
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
				return meta4, errors.Wrap(uploadError, "uploading")
			}

			meta4.Files[fileIdx].URLs = append(
				meta4.Files[fileIdx].URLs,
				metalink.URL{
					Location: uploadParams.Location,
					Priority: uploadParams.Priority,
					URL:      uri,
				},
			)
		}
	}

	return meta4, nil
}
