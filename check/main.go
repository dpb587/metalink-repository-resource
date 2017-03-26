package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dpb587/metalink-repository-resource/check/check"
	"github.com/dpb587/metalink-repository-resource/factory"
	"github.com/dpb587/metalink-repository-resource/models"
	filter_and "github.com/dpb587/metalink/repository/filter/and"
	filter_fileversion "github.com/dpb587/metalink/repository/filter/fileversion"
	"github.com/dpb587/metalink/repository/sorter"
	sorter_fileversion "github.com/dpb587/metalink/repository/sorter/fileversion"
	sorter_reverse "github.com/dpb587/metalink/repository/sorter/reverse"
)

func main() {
	var request check.Request

	err := json.NewDecoder(os.Stdin).Decode(&request)
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

	if request.Version != nil {
		addFilter, err := filter_fileversion.CreateFilter(fmt.Sprintf("> %s", request.Version.Version))
		if err != nil {
			fatal("bad stdin: version", err)
		}

		andFilter.Add(addFilter)
	}

	repository, err := factory.GetSource(request.Source.URI)
	if err != nil {
		fatal("bad stdin: source: uri", err)
	}

	err = repository.Reload()
	if err != nil {
		fatal("bad repository: load", err)
	}

	files, err := repository.FilterFiles(andFilter)
	if err != nil {
		fatal("bad repository: filter", err)
	}

	sorter.Sort(files, sorter_reverse.Sorter{Sorter: sorter_fileversion.Sorter{}})

	response := check.Response{}
	versionsSeen := map[string]bool{}

	for _, file := range files {
		if _, seen := versionsSeen[file.File.Version]; seen {
			continue
		}

		response = append(
			response,
			models.Version{
				Version: file.File.Version,
			},
		)

		versionsSeen[file.File.Version] = true

		if request.Version == nil {
			break
		}
	}

	json.NewEncoder(os.Stdout).Encode(response)
}

func fatal(msg string, err error) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err))

	os.Exit(1)
}
