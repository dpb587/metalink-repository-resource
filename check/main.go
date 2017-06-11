package main

import (
	"encoding/json"
	"os"

	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink-repository-resource/factory"
	filter_and "github.com/dpb587/metalink/repository/filter/and"
	"github.com/dpb587/metalink/repository/sorter"
	sorter_fileversion "github.com/dpb587/metalink/repository/sorter/fileversion"
)

func main() {
	var request Request

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal("check: bad stdin: parse error", err)
	}

	andFilter := filter_and.NewFilter()

	err = request.ApplyFilter(&andFilter)
	if err != nil {
		api.Fatal("check: bad stdin: filter error", err)
	}

	repository, err := factory.GetSource(request.Source.URI, request.Source.Options)
	if err != nil {
		api.Fatal("check: bad stdin: source: uri", err)
	}

	err = repository.Load()
	if err != nil {
		api.Fatal("check: bad repository: load", err)
	}

	metalinks, err := repository.Filter(andFilter)
	if err != nil {
		api.Fatal("check: filtering metalinks", err)
	}

	sorter.Sort(metalinks, sorter_fileversion.Sorter{})

	response := Response{}

	for _, meta4 := range metalinks {
		response = append(
			response,
			api.Version{
				Version: meta4.Metalink.Files[0].Version,
			},
		)

		if request.Version == nil {
			break
		}
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		api.Fatal("check: bad stdout: json", err)
	}
}
