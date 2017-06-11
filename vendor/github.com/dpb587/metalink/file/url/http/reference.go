package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/cheggaaa/pb"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/dpb587/metalink/file"
)

type Reference struct {
	client *http.Client
	url    string
}

var _ file.Reference = Reference{}

func NewReference(client *http.Client, url string) Reference {
	return Reference{
		client: client,
		url:    url,
	}
}

func (o Reference) Name() (string, error) {
	parsed, err := url.Parse(o.url)
	if err != nil {
		return "", bosherr.WrapError(err, "Parsing URL")
	}

	return filepath.Base(parsed.Path), nil
}

func (o Reference) Size() (uint64, error) {
	// @todo
	return 0, errors.New("Unsupported")
}

func (o Reference) Reader() (io.ReadCloser, error) {
	response, err := o.client.Get(o.url)
	if err != nil {
		return nil, bosherr.WrapError(err, "Loading URL")
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected response code: %d", response.StatusCode)
	}

	return response.Body, nil
}

func (o Reference) ReaderURI() string {
	return o.url
}

func (o Reference) WriteFrom(_ file.Reference, _ *pb.ProgressBar) error {
	// @todo
	return errors.New("Unsupported")
}