package axiom

import "github.com/dpb587/metalink/repository/filter"

type Factory struct{}

var _ filter.FilterFactory = Factory{}

func (Factory) Create(_ string) (filter.Filter, error) {
	return Filter{}, nil
}
