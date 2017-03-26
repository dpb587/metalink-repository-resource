package models

type Source struct {
	URI     string `json:"uri"`
	Version string `json:"version,omitempty"`

	GitPrivateKey string `json:"git_private_key"`

	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`
}
