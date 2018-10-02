package api

import (
  "regexp"
	"encoding/json"

  "github.com/pkg/errors"
)

type Regexp struct {
  *regexp.Regexp
}

func (r *Regexp) UnmarshalJSON(bytes []byte) error {
  var s string

  err := json.Unmarshal(bytes, &s)
  if err != nil {
    return err
  }

  parsed, err := regexp.Compile(s)
  if err != nil {
    return errors.Wrap(err, "parsing regexp")
  }

  r.Regexp = parsed

  return nil
}

type RegexpList []*Regexp

func (rl RegexpList) AsRegexp() []*regexp.Regexp {
  var as []*regexp.Regexp

  for _, r := range rl {
    as = append(as, r.Regexp)
  }

  return as
}
