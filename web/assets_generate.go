// +build ignore

package main

import (
	"log"

	"github.com/hsmade/OSM-ARDF/web"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(web.Assets, vfsgen.Options{
		PackageName:  "web",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
