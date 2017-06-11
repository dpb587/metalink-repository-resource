package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink-repository-resource/factory"
	filter_and "github.com/dpb587/metalink/repository/filter/and"
	"github.com/dpb587/metalink/transfer"
)

func main() {
	if len(os.Args) < 2 {
		api.Fatal("in: bad invocation", fmt.Errorf("%s DESTINATION-DIR", os.Args[0]))
	}

	destination := os.Args[1]

	err := os.MkdirAll(destination, 0755)
	if err != nil {
		api.Fatal("in: bad destination", err)
	}

	var request Request

	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal("in: bad stdin: parse error", err)
	}

	andFilter := filter_and.NewFilter()

	err = request.ApplyFilter(&andFilter)
	if err != nil {
		api.Fatal("in: bad stdin: filter error", err)
	}

	repository, err := factory.GetSource(request.Source.URI, request.Source.Options)
	if err != nil {
		api.Fatal("in: bad stdin: source: uri", err)
	}

	err = repository.Load()
	if err != nil {
		api.Fatal("in: bad repository: load", err)
	}

	metalinks, err := repository.Filter(andFilter)
	if err != nil {
		api.Fatal("in: bad filter", err)
	}

	if len(metalinks) == 0 {
		api.Fatal("in: nothing to do", errors.New("match not found"))
	} else if len(metalinks) > 1 {
		api.Fatal("in: too much to do", errors.New("multiple matches found"))
	}

	var fileCount int
	var byteCount uint64

	for _, file := range metalinks[0].Metalink.Files {
		if !request.Params.SkipDownload {
			local, err := factory.GetOrigin(metalink.URL{URL: filepath.Join(destination, file.Name)})
			if err != nil {
				api.Fatal(fmt.Sprintf("in: bad file: %s", file.Name), err)
			}

			progress := pb.New64(int64(file.Size)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Second).SetWidth(80)
			progress.ShowPercent = false
			progress.Output = os.Stderr

			verifier, err := factory.DynamicVerification.GetVerifier(file, request.Source.SkipHashVerification, request.Source.SkipSignatureVerification, request.Source.SignatureTrustStore)
			if err != nil {
				api.Fatal(fmt.Sprintf("in: bad file verifier: %s", file.Name), err)
			}

			err = transfer.NewVerifiedTransfer(factory.GetMetaURLLoaderFactory(), factory.GetURLLoaderFactory(), verifier).TransferFile(file, local, progress)
			if err != nil {
				api.Fatal(fmt.Sprintf("in: bad file transfer: %s", file.Name), err)
			}
		}

		byteCount = byteCount + file.Size
	}

	fileCount = len(metalinks[0].Metalink.Files)

	err = os.MkdirAll(filepath.Join(destination, ".resource"), 0700)
	if err != nil {
		api.Fatal("in: fs metadata: mkdir", err)
	}

	meta4bytes, err := metalink.Marshal(metalinks[0].Metalink)
	if err != nil {
		api.Fatal("in: fs metadata: marshal metalink", err)
	}

	err = ioutil.WriteFile(filepath.Join(destination, ".resource", "metalink.meta4"), meta4bytes, 0744)
	if err != nil {
		api.Fatal("in: fs metadata: marshal metalink", err)
	}

	err = ioutil.WriteFile(filepath.Join(destination, ".resource", "version"), []byte(request.Version.Version), 0744)
	if err != nil {
		api.Fatal("in: fs metadata: version", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(Response{
		Version: request.Version,
		Metadata: []api.Metadata{
			{
				Name:  "files",
				Value: fmt.Sprintf("%d", fileCount),
			},
			{
				Name:  "bytes",
				Value: fmt.Sprintf("%d", byteCount),
			},
		},
	})
	if err != nil {
		api.Fatal("in: bad stdout: json", err)
	}
}