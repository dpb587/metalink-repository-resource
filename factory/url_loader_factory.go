package factory

import (
	"fmt"

	"github.com/dpb587/metalink-repository-resource/api"
	"github.com/dpb587/metalink/file/url"
	fileurl "github.com/dpb587/metalink/file/url/file"
	ftpurl "github.com/dpb587/metalink/file/url/ftp"
	httpurl "github.com/dpb587/metalink/file/url/http"
	s3url "github.com/dpb587/metalink/file/url/s3"
	"github.com/dpb587/metalink/file/url/urlutil"
)

func GetURLLoader(handlers []api.HandlerSource) url.Loader {
	loader := url.NewMultiLoader()

	for _, handlerSource := range handlers {
		var handlerLoader url.Loader

		switch handlerSource.Type {
		case "s3":
			opts := s3url.Options{}

			if val, ok := handlerSource.Options["access_key"]; ok {
				valStr, ok := val.(string)
				if !ok {
					panic("unsupported handler option: s3: access_key: expected string")
				}

				opts.AccessKey = valStr
			}

			if val, ok := handlerSource.Options["secret_key"]; ok {
				valStr, ok := val.(string)
				if !ok {
					panic("unsupported handler option: s3: secret_key: expected string")
				}

				opts.SecretKey = valStr
			}

			if val, ok := handlerSource.Options["role_arn"]; ok {
				valStr, ok := val.(string)
				if !ok {
					panic("unsupported handler option: s3: role_arn: expected string")
				}

				opts.RoleARN = valStr
			}

			handlerLoader = s3url.NewLoader(opts)
		default:
			panic(fmt.Errorf("unsupported handler: %s", handlerSource.Type))
		}

		if len(handlerSource.Include) > 0 || len(handlerSource.Exclude) > 0 {
			handlerLoader = urlutil.NewFilteredLoader(
				handlerLoader,
				handlerSource.Include.AsRegexp(),
				handlerSource.Exclude.AsRegexp(),
			)
		}

		loader.Add(handlerLoader)
	}

	// defaults
	file := fileurl.NewLoader()
	loader.Add(file)
	loader.Add(ftpurl.Loader{})
	loader.Add(httpurl.Loader{})
	loader.Add(s3url.NewLoader(s3url.Options{}))
	loader.Add(urlutil.NewEmptySchemeLoader(file))

	return loader
}
