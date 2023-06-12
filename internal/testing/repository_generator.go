package testing

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

func GenerateRepository() (string, error) {
	repositorydir, err := ioutil.TempDir("", "metalink-repository-resource-fake-repository")
	if err != nil {
		return "", err
	}

	err = RunCommands(
		repositorydir,
		[]string{
			"git init .",
			"git config receive.denyCurrentBranch updateInstead",
			"git config user.email testing@localhost",
			"git config user.name testing",
			"mkdir component",
			`echo '{"files":[{"name":"test","version":"1.0.0"}]}' > component/v1.0.0.meta4`,
			`echo '{"files":[{"name":"test","version":"2.0.0"}]}' > component/v2.0.0.meta4`,
			`echo '{"files":[{"name":"test","version":"1.1.0"}]}' > component/v1.1.0.meta4`,
			"git add . && git commit -m 'init'",
		},
	)

	if err != nil {
		os.RemoveAll(repositorydir)

		return "", errors.Wrap(err, "generating repository")
	}

	return repositorydir, nil
}
