package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink-repository-resource/factory"
	"github.com/dpb587/metalink-repository-resource/in/in"
	"github.com/dpb587/metalink/crypto"
	filter_and "github.com/dpb587/metalink/repository/filter/and"
	filter_fileversion "github.com/dpb587/metalink/repository/filter/fileversion"
	"github.com/dpb587/metalink/repository/sorter"
	sorter_fileversion "github.com/dpb587/metalink/repository/sorter/fileversion"
	sorter_reverse "github.com/dpb587/metalink/repository/sorter/reverse"
)

func main() {
	if len(os.Args) < 2 {
		fatal("bad invocation", fmt.Errorf("%s DESTINATION-DIR", os.Args[0]))
	}

	destination := os.Args[1]

	err := os.MkdirAll(destination, 0755)
	if err != nil {
		fatal("bad destination", err)
	}

	var request in.Request

	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("bad stdin: parse error", err)
	}

	andFilter := filter_and.NewFilter()

	if request.Source.Version != "" {
		addFilter, err := filter_fileversion.CreateFilter(request.Source.Version)
		if err != nil {
			fatal("bad stdin: source: version", err)
		}

		andFilter.Add(addFilter)
	}

	addFilter, err := filter_fileversion.CreateFilter(request.Version.Version)
	if err != nil {
		fatal("bad stdin: version", err)
	}

	andFilter.Add(addFilter)

	repository, err := factory.GetSource(request.Source.URI)
	if err != nil {
		fatal("bad stdin: source: uri", err)
	}

	err = repository.Load()
	if err != nil {
		fatal("bad repository: load", err)
	}

	metalinks, err := repository.Filter(andFilter)
	if err != nil {
		fatal("bad repository: filter", err)
	}

	if len(metalinks) == 0 {
		fatal("nothing to do", errors.New("version not found"))
	}

	sorter.Sort(metalinks, sorter_reverse.Sorter{Sorter: sorter_fileversion.Sorter{}})

	for _, file := range metalinks[0].Metalink.Files {
		local, err := factory.GetOrigin(filepath.Join(destination, file.Name))
		if err != nil {
			fatal("bad file: local", err)
		}

		prefix := fmt.Sprintf("%s\tget\t", file.Name)
		progress := pb.New64(int64(file.Size)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Second).SetWidth(80 + len(prefix))
		progress.Prefix(prefix)
		progress.Output = os.Stderr
		progress.ShowPercent = false

		for _, url := range file.URLs {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s\torigin\t%s", file.Name, url.URL))

			remote, err := factory.GetOrigin(url.URL)
			if err != nil {
				fatal("bad file: origin", err)
			}

			progress.Start()

			err = local.WriteFrom(remote, progress)
			if err != nil {
				fatal("bad file: transfer", err)
			}

			progress.Finish()

			knownHashes := []string{}

			for _, hash := range file.Hashes {
				knownHashes = append(knownHashes, hash.Type)
			}

			algorithm, err := crypto.GetStrongestAlgorithm(knownHashes)
			if err != nil {
				fatal("bad file: verify", errors.New("no hash found"))
			}

			var expectedHash *metalink.Hash

			for _, seekHash := range file.Hashes {
				if seekHash.Type != crypto.GetDigestType(algorithm.Name()) {
					continue
				}

				expectedHash = &seekHash

				break
			}

			if expectedHash == nil {
				fatal("bad file: verify", errors.New("missing hash"))
			}

			reader, err := local.Reader()
			if err != nil {
				fatal("bad file: verify: read", err)
			}

			actualHash, err := algorithm.CreateDigest(reader)
			if err != nil {
				fatal("bad file: verify: algorithm: digest", err)
			}

			if expectedHash.Hash != crypto.GetDigestHash(actualHash) {
				fatal("bad file: verify: mismatch", fmt.Errorf("expected %s, actual %s", expectedHash.Hash, crypto.GetDigestHash(actualHash)))
			}

			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s\t%s\t%s", file.Name, expectedHash.Type, expectedHash.Hash))
		}
	}

	versionFile, err := os.Create(filepath.Join(destination, "version"))
	if err != nil {
		fatal("version file: create", err)
	}

	defer versionFile.Close()

	_, err = versionFile.WriteString(request.Version.Version)
	if err != nil {
		fatal("version file: write", err)
	}

	json.NewEncoder(os.Stdout).Encode(in.Response{Version: request.Version})
}

func fatal(msg string, err error) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err))

	os.Exit(1)
}
