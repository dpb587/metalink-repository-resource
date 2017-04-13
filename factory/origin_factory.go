package factory

import (
	"github.com/dpb587/metalink/origin"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

func getOriginFactory() origin.OriginFactory {
	logger := boshlog.NewLogger(boshlog.LevelError)
	fs := boshsys.NewOsFileSystem(logger)

	return origin.NewDefaultFactory(fs)
}

func GetOrigin(uri string) (origin.Origin, error) {
	return getOriginFactory().Create(uri)
}
