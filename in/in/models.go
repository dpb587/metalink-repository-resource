package in

import "github.com/dpb587/metalink-repository-resource/models"

type Request struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
	Params  Params         `json:"params"`
}

type Params struct {
	SkipDownload bool `json:"skip_download"`
}

type Response struct {
	Version  models.Version    `json:"version"`
	Metadata []models.Metadata `json:"metadata,omitempty"`
}
