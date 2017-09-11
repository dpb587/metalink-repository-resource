package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink-repository-resource/factory"
)

func main() {
	var request Request

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal("out: bad stdin: parse error", err)
	}

	metalinkPaths, err := filepath.Glob(request.Params.Metalink)
	if err != nil {
		api.Fatal("out: bad metalink: globbing path", err)
	} else if len(metalinkPaths) == 0 {
		api.Fatal("out: bad metalink: path not found", err)
	} else if len(metalinkPaths) > 1 {
		api.Fatal("out: bad metalink: multiple paths found when one is expected", err)
	}

	metalinkPath := metalinkPaths[0]

	metalinkBytes, err := ioutil.ReadFile(metalinkPath)
	if err != nil {
		api.Fatal("out: bad metalink: read error", err)
	}

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
		metalinkName = path.Base(metalinkFile.Name())
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
