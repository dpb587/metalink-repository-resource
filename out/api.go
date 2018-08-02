package main

import (
	"github.com/dpb587/metalink-repository-resource/api"
)

type Request struct {
	Source api.Source `json:"source"`
	Params Params     `json:"params"`
}

type Params struct {
	Metalink       string                 `json:"metalink"`
	Files          []string               `json:"files"`
	Version        string                 `json:"version"`
	Rename         string                 `json:"rename,omitempty"`
	RenameFromFile string                 `json:"rename_from_file,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

type Response struct {
	Version  api.Version    `json:"version"`
	Metadata []api.Metadata `json:"metadata,omitempty"`
}
