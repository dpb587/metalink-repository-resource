package api

import (
	"path"
	"path/filepath"

	"github.com/dpb587/metalink/repository"
	"github.com/dpb587/metalink/repository/filter"
)

type FilePathFilter struct {
	Glob string
}

var _ filter.Filter = FilePathFilter{}

func CreateFilePathFilter(glob string) (FilePathFilter, error) {
	_, err := filepath.Match(glob, "")
	if err != nil {
		return FilePathFilter{}, err
	}

	return FilePathFilter{
		Glob: glob,
	}, nil
}

func (f FilePathFilter) IsTrue(meta4 repository.RepositoryMetalink) (bool, error) {
	match, err := filepath.Match(f.Glob, path.Base(meta4.Reference.Path))
	if err != nil {
		return false, err
	}

	return match, nil
}
