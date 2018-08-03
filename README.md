# metalink-repository-resource

A [Concourse](https://concourse.ci) resource for managing versions/files in a [Metalink repository](https://github.com/dpb587/metalink/tree/master/repository#metalink-repository).


## Source Configuration

 * **`uri`** - location of the repository
 * `signature_trust_store` - identities and keys used for signature verification
 * `skip_hash_verification` - skip hash verification of files
 * `skip_signature_verification` - skip signature verification of files
 * `version` - a [supported](https://github.com/Masterminds/semver#basic-comparisons) version constraint (e.g. `^4.1`)
 * `options` - a hash of supported options, depending on the repository type
    * for git repositories
       * `private_key` - a SSH private key for `git+ssh` URIs
    * for s3 repositories
       * `access_key` - access key for private S3 endpoints
       * `secret_key` - secret key for private S3 endpoints
 * `include_files` - a list of file globs to match when downloading a version's files (used by `in`)
 * `exclude_files` - a list of file globs to skip when downloading a version's files (used by `in`)
 * `mirror_files` - a list of mirror configurations for mirroring files (used by `out`)
    * **`destination`** - the mirror URI for uploading files (templated; `Name`, `Version`, `SHA1`, `SHA256`, `SHA512`, `MD5`)
    * `location` - the ISO3166-1 alpha-2 country code for the geographical location (embedded in the metalink)
    * `priority` - a priority for the file (embedded in the metalink)
    * `env` - environment variables for credentials used when uploading the file


## `check`

Check for new versions in the repository.

Metadata:

 * `version` - semantic version (e.g. `4.1.2`)


## `in`

Download and verify the referenced file(s).

 * `.resource/metalink.meta4` - metalink data used when downloading the file
 * `.resource/version` - version downloaded (e.g. `4.1.2`)
 * `*` - the downloaded file(s) from the metalink

Parameters:

 * `skip_download` - do not download blobs (only `metalink.meta4` and `version` will be available)


## `out`

Publish a metalink file to the repository.

Parameters:

 * `metalink` - path to the metalink file (one of `metalink` or `files` must be configured)
 * `files` - a list of glob paths for files to create a metalink from (one of `metalink` or `files` must be configured; requires `version`)
 * `version` - path to a file with the version number (only effective with `files`)
 * `rename` - publish the metalink file with a different file name
 * `rename_from_file` - path to a file whose content is the metalink file name (alternative to `rename`)
 * `options` - a hash of supported options, depending on the repository type
    * for git repositories
       * `author_name`, `author_email` - the commit author
       * `message` - the commit message


## License

[MIT License](LICENSE)
