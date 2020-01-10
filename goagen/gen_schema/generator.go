package genschema

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/shogo82148/goa-v1/design"
	"github.com/shogo82148/goa-v1/goagen/codegen"
	"github.com/shogo82148/goa-v1/goagen/utils"
)

//NewGenerator returns an initialized instance of a JavaScript Client Generator
func NewGenerator(options ...Option) *Generator {
	g := &Generator{}

	for _, option := range options {
		option(g)
	}

	return g
}

// Generator is the application code generator.
type Generator struct {
	API      *design.APIDefinition // The API definition
	OutDir   string                // Path to output directory
	genfiles []string              // Generated files
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	var outDir, ver string
	set := flag.NewFlagSet("app", flag.PanicOnError)
	set.StringVar(&outDir, "out", "", "")
	set.StringVar(&ver, "version", "", "")
	set.String("design", "", "")
	set.Parse(os.Args[1:])

	if err := codegen.CheckVersion(ver); err != nil {
		return nil, err
	}

	g := &Generator{OutDir: outDir, API: design.Design}

	return g.Generate()
}

// Generate produces the skeleton main.
func (g *Generator) Generate() (_ []string, err error) {
	if g.API == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	s := APISchema(g.API)
	js, err := s.JSON()
	if err != nil {
		return
	}

	g.OutDir = filepath.Join(g.OutDir, "schema")
	os.RemoveAll(g.OutDir)
	os.MkdirAll(g.OutDir, 0755)
	g.genfiles = append(g.genfiles, g.OutDir)
	schemaFile := filepath.Join(g.OutDir, "schema.json")
	if err = ioutil.WriteFile(schemaFile, js, 0644); err != nil {
		return
	}
	g.genfiles = append(g.genfiles, schemaFile)

	return g.genfiles, nil
}

// Cleanup removes all the files generated by this generator during the last invokation of Generate.
func (g *Generator) Cleanup() {
	for _, f := range g.genfiles {
		os.Remove(f)
	}
	g.genfiles = nil
}
