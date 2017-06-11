package api

import (
	"github.com/dpb587/metalink/repository/filter/and"
	"github.com/dpb587/metalink/repository/filter/fileversion"
)

type Source struct {
	URI     string                 `json:"uri"`
	Options map[string]interface{} `json:"options,omitempty"`

	SkipHashVerification      bool   `json:"skip_hash_verification,omitempty"`
	SkipSignatureVerification bool   `json:"skip_signature_verification,omitempty"`
	SignatureTrustStore       string `json:"signature_trust_store,omitempty"`

	Version string `json:"version,omitempty"`
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
