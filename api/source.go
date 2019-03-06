package api

import (
	"github.com/dpb587/metalink/repository/filter/and"
	"github.com/dpb587/metalink/repository/filter/fileversion"
)

type Source struct {
	URI          string                 `json:"uri"`
	Options      map[string]interface{} `json:"options,omitempty"`
	MetalinkGlob string                 `json:"metalink_glob,omitempty"`

	SkipHashVerification      bool   `json:"skip_hash_verification,omitempty"`
	SkipSignatureVerification bool   `json:"skip_signature_verification,omitempty"`
	SignatureTrustStore       string `json:"signature_trust_store,omitempty"`

	URLHandlers []HandlerSource `json:"url_handlers,omitempty"`

	MirrorFiles []MirrorFileParams `json:"mirror_files,omitempty"`

	IncludeFiles []string `json:"include_files,omitempty"`
	ExcludeFiles []string `json:"exclude_files,omitempty"`

	Version string `json:"version,omitempty"`
}

type MirrorFileParams struct {
	Destination string            `json:"destination"`
	Location    string            `json:"location,omitempty"`
	Priority    *uint             `json:"priority,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

type HandlerSource struct {
	Type    string                 `json:"type"`
	Include RegexpList             `json:"include,omitempty"`
	Exclude RegexpList             `json:"exclude,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

func (s Source) ApplyFilter(filter *and.Filter) error {
	if s.Version == "" {
		return nil
	}

	addFilter, err := fileversion.CreateFilter(s.Version)
	if err != nil {
		return err
	}

	filter.Add(addFilter)

	return nil
}
