package main

import "google.golang.org/protobuf/compiler/protogen"

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + "_router.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	return g
}
