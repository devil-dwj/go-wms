package main

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {

	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !strings.HasSuffix(f.Proto.GetName(), "_pr.proto") {
				f.Generate = false
			}
			gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}
