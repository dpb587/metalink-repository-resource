package main

import (
	"os"

	"github.com/dpb587/metalink/cli/cmd"
	"github.com/dpb587/metalink/cli/verification"
	// metaurldefaultloader "github.com/dpb587/metalink/file/metaurl/defaultloader"
	"github.com/dpb587/metalink/file/metaurl"
	urldefaultloader "github.com/dpb587/metalink/file/url/defaultloader"
	"github.com/dpb587/metalink/storage"
	flags "github.com/jessevdk/go-flags"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

func main() {
	logger := boshlog.NewLogger(boshlog.LevelError)
	fs := boshsys.NewOsFileSystem(logger)

	urlLoader := urldefaultloader.New(fs)
	// metaurlLoader := metaurldefaultloader.New()
	metaurlLoader := metaurl.NewLoaderFactory()
	storageFactory := storage.NewDefaultFactory(fs)

	meta4 := cmd.Meta4{
		Metalink:       "metalink.meta4",
		StorageFactory: storageFactory,
	}

	meta4file := cmd.Meta4File{
		Meta4: meta4,
	}

	verifier := verification.NewDynamicVerifierImpl(fs)

	c := struct {
		AddFile        cmd.AddFile        `command:"add-file" description:"Add a new file by name"`
		ImportFile     cmd.ImportFile     `command:"import-file" description:"Import a local file"`
		ImportMetalink cmd.ImportMetalink `command:"import-metalink" description:"Import files from an existing metalink"`
		RemoveFile     cmd.RemoveFile     `command:"remove-file" description:"Remove an existing file by name"`
		Files          cmd.Files          `command:"files" description:"List existing files by name"`

		Create cmd.Create `command:"create" description:"Create a new metalink file"`

		FileDownload   cmd.FileDownload   `command:"file-download" description:"Download file"`
		FileHash       cmd.FileHash       `command:"file-hash" description:"Show hash of a file"`
		FileHashes     cmd.FileHashes     `command:"file-hashes" description:"List hashes of a file"`
		FileRemoveURL  cmd.FileRemoveURL  `command:"file-remove-url" description:"Remove download URL of a file"`
		FileSetHash    cmd.FileSetHash    `command:"file-set-hash" description:"Set hash of a file"`
		FileSetSize    cmd.FileSetSize    `command:"file-set-size" description:"Set size of a file"`
		FileSetURL     cmd.FileSetURL     `command:"file-set-url" description:"Set download URL of a file"`
		FileSetVersion cmd.FileSetVersion `command:"file-set-version" description:"Set version of a file"`
		FileUpload     cmd.FileUpload     `command:"file-upload" description:"Upload file and add URL"`
		FileURLs       cmd.FileURLs       `command:"file-urls" description:"List existing URLs"`
		FileVerify     cmd.FileVerify     `command:"file-verify" description:"Verify integrity of a local file"`
		FileVersion    cmd.FileVersion    `command:"file-version" description:"Show version of a file"`

		SetOrigin    cmd.SetOrigin    `command:"set-origin" description:"Set origin URI for the metalink"`
		SetPublished cmd.SetPublished `command:"set-published" description:"Set published timestamp"`
		SetUpdated   cmd.SetUpdated   `command:"set-updated" description:"Set updated timestamp"`
	}{
		AddFile:        cmd.AddFile{Meta4: meta4},
		ImportFile:     cmd.ImportFile{Meta4File: meta4file, URLLoader: urlLoader},
		ImportMetalink: cmd.ImportMetalink{Meta4: meta4, StorageFactory: storageFactory},
		RemoveFile:     cmd.RemoveFile{Meta4: meta4},
		Files:          cmd.Files{Meta4: meta4},

		Create: cmd.Create{Meta4: meta4},

		FileDownload:   cmd.FileDownload{Meta4File: meta4file, URLLoader: urlLoader, MetaURLLoader: metaurlLoader, Verification: verifier},
		FileHash:       cmd.FileHash{Meta4File: meta4file},
		FileHashes:     cmd.FileHashes{Meta4File: meta4file},
		FileRemoveURL:  cmd.FileRemoveURL{Meta4File: meta4file},
		FileSetHash:    cmd.FileSetHash{Meta4File: meta4file},
		FileSetSize:    cmd.FileSetSize{Meta4File: meta4file},
		FileSetURL:     cmd.FileSetURL{Meta4File: meta4file},
		FileSetVersion: cmd.FileSetVersion{Meta4File: meta4file},
		FileUpload:     cmd.FileUpload{Meta4File: meta4file, URLLoader: urlLoader},
		FileVerify:     cmd.FileVerify{Meta4File: meta4file, Verification: verifier},
		FileURLs:       cmd.FileURLs{Meta4File: meta4file},
		FileVersion:    cmd.FileVersion{Meta4File: meta4file},

		SetOrigin:    cmd.SetOrigin{Meta4: meta4},
		SetPublished: cmd.SetPublished{Meta4: meta4},
		SetUpdated:   cmd.SetUpdated{Meta4: meta4},
	}

	var parser = flags.NewParser(&c, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
