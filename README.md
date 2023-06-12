# metalink-repository-resource

A [Concourse](https://concourse.ci) resource for managing versions/files in a [Metalink repository](https://github.com/dpb587/metalink/tree/master/repository#metalink-repository).


## Source Configuration

 * **`uri`** - location of the repository
 * `signature_trust_store` - identities and keys used for signature verification
 * `skip_hash_verification` - skip hash verification of files
 * `skip_signature_verification` - skip signature verification of files
 * `version` - a [supported](https://github.com/Masterminds/semver#basic-comparisons) version constraint (e.g. `^4.1`)
 * `filters` - a list of [supported](#filters) filters to limit the discovered metalinks
 * `options` - a hash of supported options, depending on the repository type
    * for git repositories
       * `private_key` - a SSH private key for `git+ssh` URIs
       * `rebase` - number of rebase attempts when pushing (default `3`)
    * for s3 repositories
       * `access_key` - access key for private S3 endpoints
       * `secret_key` - secret key for private S3 endpoints
       * `role_arn` - role arn for private S3 endpoints when using AssumeRole
 * `include_files` - a list of file globs to match when downloading a version's files (used by `in`)
 * `exclude_files` - a list of file globs to skip when downloading a version's files (used by `in`)
 * `url_handlers` - a list of URL handlers for custom download/upload configurations
    * **`type`** - handler type (i.e. `s3`)
    * `include` - a list of URIs that should use this handler (regex'd)
    * `exclude` - a list of URIs that should avoid this handler (regex'd)
    * `options` - a hash of supported options, depending on `type`
       * for `s3`:
          * `access_key` - access key for private S3 endpoints
          * `secret_key` - secret key for private S3 endpoints
          * `role_arn` - role arn for private S3 endpoints when using AssumeRole
 * `mirror_files` - a list of mirror configurations for mirroring files (used by `out`)
    * **`destination`** - the mirror URI for uploading files (templated; `Name`, `Version`, `SHA1`, `SHA256`, `SHA512`, `MD5`)
    * `location` - the ISO3166-1 alpha-2 country code for the geographical location (embedded in the metalink)
    * `priority` - a priority for the file (embedded in the metalink)


## Operations

### `check`

Check for new versions in the repository.

Metadata:

 * `version` - semantic version (e.g. `4.1.2`)


### `in`

Download and verify the referenced file(s).

 * `.resource/metalink.meta4` - metalink data used when downloading the file
 * `.resource/version` - version downloaded (e.g. `4.1.2`)
 * `*` - the downloaded file(s) from the metalink

Parameters:

 * `include_files` - a list of file globs to match when downloading files (intersects with `include_files` from source configuration, when present)
 * `skip_download` - do not download blobs (only `metalink.meta4` and `version` will be available)


### `out`

Publish a metalink file to the repository.

Parameters:

 * `metalink` - path to the metalink file (one of `metalink` or `files` must be configured)
 * `files` - a list of glob paths for files to create a metalink from (one of `metalink` or `files` must be configured; requires `version`)
 * `version` - path to a file with the version number (only effective with `files`)
 * `rename` - publish the metalink file with a different file name (templated; `Version`)
 * `rename_from_file` - path to a file whose content is the metalink file name (alternative to `rename`)
 * `options` - a hash of supported options, depending on the repository type
    * for git repositories
       * `author_name`, `author_email` - the commit author
       * `message` - the commit message


## Usage


To use this resource type, you should configure it in the [`resource_types`](https://concourse-ci.org/resource-types.html) section of your pipeline.

    - name: metalink-repository
      type: docker-image
      source:
        repository: dpb587/metalink-repository-resource


### URL Credentials

When working with authenticated URLs (for either upload or download), configure the `url_handlers` option of the resource:

    url_handlers:
    - type: s3
      options:
        access_key: AKIAA1B2C3...
        secret_key: a1b2c3d4e5...

When using multiple URLs which require different configurations, use the `include` or `exclude` options to restrict usage:

    url_handlers:
    - type: s3
      include:
      - s3://[^/]+/org1-bucket-name/
      options:
        access_key: AKIAA1B2C3...
        secret_key: a1b2c3d4e5...
    - type: s3
      include:
      - s3://[^/]+/org2-bucket-name/
      options:
        access_key: AKIAB2C3D4...
        secret_key: b2c3d4e5f6...
		mirror_files:
    - destination: s3://s3-external-1.amazonaws.com/org1-bucket-name/my-private-blobs/{{.Version}}/{{.Name}}
    - destination: s3://s3-external-1.amazonaws.com/org2-bucket-name/my-private-blobs/{{.Version}}/{{.Name}}


### Filters

The `fileversion` and `repositorypath` filters are supported.

    filters:
    - repositorypath: prefix-*.meta4
    - fileversion: 27.x              # equivalent to using source version


## License

[MIT License](LICENSE)
