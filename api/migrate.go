package api

func MigrateSource(source *Source) {
	// previously upload settings were specified via mirror_files.env
	// previously only s3 was supported
	// assumes a single, global s3 handler
	for _, mirrorFiles := range source.MirrorFiles {
		var optsFound bool
		optsMap := map[string]interface{}{}

		for mk, mv := range mirrorFiles.Env {
			optsFound = true

			switch mk {
			case "AWS_ACCESS_KEY_ID":
				optsMap["access_key"] = mv
			case "AWS_SECRET_ACCESS_KEY":
				optsMap["secret_key"] = mv
			}
		}

		if optsFound {
			source.URLHandlers = append(source.URLHandlers, HandlerSource{
				Type: "s3",
				Options: optsMap,
			})
		}
	}
}
