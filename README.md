# metalink-repository-resource

A [Concourse](https://concourse.ci) resource for pulling versions/files from a [Metalink repository](https://github.com/dpb587/metalink/tree/master/repository#metalink-repository).


## Source Configuration

 * **`uri`** - location of the repository
 * `version` - a [supported](https://github.com/Masterminds/semver#basic-comparisons) version constraint (e.g. `^4.1`)


### `git` Repository

 * `git_private_key` - a SSH private key for `git+ssh` URIs (@todo)


### `s3` Repository

 * `s3_access_key` - access key for non-public S3 endpoints (@todo)
 * `s3_secret_key` - secret key for non-public S3 endpoints (@todo)


## `check`

Check for new versions in the repository.

Metadata:

 * `version` - semantic version (e.g. `4.1.2`)


## `in`

Download and verify the referenced file.

 * `.metalink/metalink.meta4` - metalink file used when downloading the file
 * `.metalink/version` - version downloaded (e.g. `4.1.2`)
 * `*` - the downloaded file(s) from the metalink


## `out`

Not supported.


## License

[MIT License](LICENSE)
