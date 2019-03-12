package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/workpool"
	"github.com/cheggaaa/pb"
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink-repository-resource/factory"
	filter_and "github.com/dpb587/metalink/repository/filter/and"
	"github.com/dpb587/metalink/transfer"
	"github.com/dpb587/metalink/verification"
)

func main() {
	if len(os.Args) < 2 {
		api.Fatal("in: bad invocation", fmt.Errorf("%s DESTINATION-DIR", os.Args[0]))
	}

	destination, err := filepath.Abs(os.Args[1])
	if err != nil {
		api.Fatal("in: bad destination", err)
	}

	err = os.MkdirAll(destination, 0755)
	if err != nil {
		api.Fatal("in: bad destination", err)
	}

	request := Request{}
	request.Source.Parallel = 2

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

	urlLoader := factory.GetURLLoader(request.Source.URLHandlers)

	var fileCount int
	var byteCount uint64

	var parallelize []func()
	var pbList []*pb.ProgressBar
	var pbPool *pb.Pool

	verifierResults := bytes.NewBuffer(nil)

	for _, file := range metalinks[0].Metalink.Files {
		var matched = true

		if len(request.Source.IncludeFiles) > 0 {
			matched = false

			for _, pattern := range request.Source.IncludeFiles {
				if match, _ := filepath.Match(pattern, file.Name); match {
					matched = true

					break
				}
			}
		}

		if matched && len(request.Params.IncludeFiles) > 0 {
			matched = false

			for _, pattern := range request.Params.IncludeFiles {
				if match, _ := filepath.Match(pattern, file.Name); match {
					matched = true

					break
				}
			}
		}

		if matched && len(request.Source.ExcludeFiles) > 0 {
			for _, pattern := range request.Source.ExcludeFiles {
				if match, _ := filepath.Match(pattern, file.Name); match {
					matched = false

					break
				}
			}
		}

		if !matched {
			continue
		}

		if !request.Params.SkipDownload {
			file := file // closure

			prefix := fmt.Sprintf("%.40s", file.Name)

			progress := pb.New64(int64(file.Size))
			progress.Prefix(prefix + strings.Repeat(" ", 40-len(prefix)))
			progress.Units = pb.U_BYTES
			progress.SetRefreshRate(time.Second)
			progress.SetWidth(120)

			pbList = append(pbList, progress)

			parallelize = append(parallelize, func() {
				defer progress.Finish()

				local, err := urlLoader.LoadURL(metalink.URL{URL: filepath.Join(destination, file.Name)})
				if err != nil {
					api.Fatal(fmt.Sprintf("in: bad file: %s", file.Name), err)
				}

				verifier, err := factory.DynamicVerification.GetVerifier(file, request.Source.SkipHashVerification, request.Source.SkipSignatureVerification, request.Source.SignatureTrustStore)
				if err != nil {
					api.Fatal(fmt.Sprintf("in: bad file verifier: %s", file.Name), err)
				}

				downloader := transfer.NewVerifiedTransfer(factory.GetMetaURLLoaderFactory(), urlLoader, verifier)

				err = downloader.TransferFile(file, local, progress, verification.NewSimpleVerificationResultReporter(verifierResults))
				if err != nil {
					api.Fatal(fmt.Sprintf("in: bad file transfer: %s", file.Name), err)
				}
			})
		}

		byteCount = byteCount + file.Size
		fileCount = fileCount + 1
	}

	if len(parallelize) > 0 {
		pbPool, _ = pb.StartPool(pbList...)
		pbPool.Output = os.Stderr

		pool, err := workpool.NewThrottler(request.Source.Parallel, parallelize)
		if err != nil {
			api.Fatal("in: parallelizing", err)
		}
		pool.Work()

		pbPool.Stop()
	}

	// defer verification output since it confuses progress bar
	os.Stderr.Write(verifierResults.Bytes())

	err = os.MkdirAll(filepath.Join(destination, ".resource"), 0700)
	if err != nil {
		api.Fatal("in: fs metadata: mkdir", err)
	}

	meta4bytes, err := metalink.MarshalXML(metalinks[0].Metalink)
	if err != nil {
		api.Fatal("in: fs metadata: marshal metalink", err)
	}

	err = ioutil.WriteFile(filepath.Join(destination, ".resource", "metalink.meta4"), meta4bytes, 0644)
	if err != nil {
		api.Fatal("in: fs metadata: metalink.meta4", err)
	}

	err = ioutil.WriteFile(filepath.Join(destination, ".resource", "version"), []byte(request.Version.Version), 0644)
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
