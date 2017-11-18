package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/tools/imports"
)

type Generator struct {
	buf         bytes.Buffer
	PackageName string
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) Generate(it *InterfaceType, fs []FuncType) error {
	g.PackageName = it.PackageName
	g.appendHeader(it)
	g.appendStructDefs(it)
	g.appendMethod(fs, "")
	return nil
}

func (g *Generator) Out(w io.Writer, filename string) error {
	dist, err := imports.Process(filename, g.buf.Bytes(), &imports.Options{Comments: true})
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, bytes.NewReader(dist)); err != nil {
		return err
	}
	return nil
}

func (g *Generator) appendHeader(it *InterfaceType) {
	g.Printf("// Code generated by \"dicon\"; DO NOT EDIT.\n")
	g.Printf("\n")
	g.Printf("package %s\n", it.PackageName)
	g.Printf("\n")
	g.Printf("import \"log\"\n")
}

func (g *Generator) appendStructDefs(it *InterfaceType) {
	g.Printf("type dicontainer struct {\n")
	g.Printf("store map[string]interface{}\n")
	g.Printf("}\n")
	g.Printf("func NewDIContainer() %s {\n", it.Name)
	g.Printf("return &dicontainer{\n")
	g.Printf("store: map[string]interface{}{},\n")
	g.Printf("}\n")
	g.Printf("}\n")
	g.Printf("\n")
}

func (g *Generator) appendMethod(funcs []FuncType, _ string) {
	for _, f := range funcs {
		g.Printf("func (d *dicontainer) %s()", f.Name)
		if len(f.ReturnTypes) != 1 {
			log.Fatalf("Must be 1 instance but %d", len(f.ReturnTypes))
		}

		returnType := f.ReturnTypes[0].Type
		g.Printf("%s%s {\n", g.relativePackageName(f.PackageName), returnType)

		g.Printf("if i, ok := d.store[\"%s\"]; ok {\n", f.Name)
		g.Printf("if instance, ok := i.(%s%s); ok {\n", g.relativePackageName(f.PackageName), f.Name)
		g.Printf("return instance\n")
		g.Printf("}\n")
		g.Printf("log.Fatal(\"cached instance is polluted\")")
		g.Printf("}\n")

		dep := make([]string, 0, len(f.ArgumentTypes))
		for i, a := range f.ArgumentTypes {
			g.Printf("dep%d := d.%s()\n", i, a.Type)
			dep = append(dep, fmt.Sprintf("dep%d", i))
		}

		g.Printf("instance := %sNew%s(%s)\n", g.relativePackageName(f.PackageName), f.Name, strings.Join(dep, ", "))
		g.Printf("d.store[\"%s\"] = instance\n", f.Name)
		g.Printf("return instance")
		g.Printf("}\n")
	}
}

func (g *Generator) relativePackageName(packageName string) string {
	if strings.Compare(g.PackageName, packageName) == 0 {
		return ""
	}

	return packageName + "."
}
