package main

import (
	"io/ioutil"
	"os"

	"github.com/devil-dwj/go-wms/cmd/generator"
	_ "github.com/devil-dwj/go-wms/cmd/plugin/routers"

	"google.golang.org/protobuf/proto"
)

func main() {

	g := generator.New("routers")

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	// 命令行参数
	// g.CommandLineParameters(g.Request.GetParameter())

	g.WrapTypes()
	g.SetPackageNames()
	g.BuildTypeNameMap()

	g.GenerateAllFiles("routers")

	data, err = proto.Marshal(g.Response)
	if err != nil {
		g.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		g.Error(err, "failed to write output proto")
	}
}
