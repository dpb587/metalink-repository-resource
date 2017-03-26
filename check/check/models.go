package check

import "github.com/dpb587/metalink-repository-resource/models"

type Request struct {
	Source  models.Source   `json:"source"`
	Version *models.Version `json:"version"`
}

type Response []models.Version
