package factory

import (
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink/file"
	"github.com/dpb587/metalink/file/url"
	"github.com/dpb587/metalink/file/url/defaultloader"
)

func GetURLLoaderFactory() url.Loader {
	return defaultloader.New()
}

func GetOrigin(ref metalink.URL) (file.Reference, error) {
	return GetURLLoaderFactory().Load(ref)
}
