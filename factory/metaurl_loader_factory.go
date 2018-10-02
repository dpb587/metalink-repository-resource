package factory

import (
	"github.com/dpb587/metalink"
	"github.com/dpb587/metalink/file"
	"github.com/dpb587/metalink/file/metaurl"
)

func GetMetaURLLoaderFactory() metaurl.Loader {
	return metaurl.NewMultiLoader()
}

func GetOriginURL(ref metalink.MetaURL) (file.Reference, error) {
	return GetMetaURLLoaderFactory().LoadMetaURL(ref)
}
