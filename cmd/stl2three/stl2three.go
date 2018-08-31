// stl2three converts an STL file to Go code using go-threejs.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gmlewis/go-stl/v2/stl"
)

var (
	pkg = flag.String("pkg", "main", "The package of the generated file")
)

func main() {
	flag.Parse()

	fmt.Printf("package %v\n\n", *pkg)
	fmt.Println(`import (
	"github.com/gmlewis/go-threejs/three"
	"github.com/gopherjs/gopherjs/js"
)`)

	for _, arg := range flag.Args() {
		tris, err := stl.Read(arg)
		if err != nil {
			log.Fatalf("stl.Read(%q): %v", arg, err)
		}

		fmt.Println()
		name := strings.TrimSuffix(arg, ".stl")
		fmt.Printf("var %v *js.Object\n", name)

		fmt.Printf(`
func init() {
	t := three.New()
	%[1]v = t.Geometry()
`, name)
		for i, tri := range tris {
			fmt.Printf(`	%v.Get("vertices").Call("push",
		t.Vector3(%v, %v, %v),
		t.Vector3(%v, %v, %v),
		t.Vector3(%v, %v, %v))`+"\n",
				name, tri.V1x, tri.V1y, tri.V1z, tri.V2x, tri.V2y, tri.V2z, tri.V3x, tri.V3y, tri.V3z)
			fmt.Printf(`	%v.Get("faces").Call("push", t.Face3(%v, %v, %v, t.Vector3(%v, %v, %v)))`+"\n",
				name, i*3, i*3+1, i*3+2, tri.Nx, tri.Ny, tri.Nz)
		}
		fmt.Printf(`	%v.Call("computeBoundingSphere")`+"\n", name)
		fmt.Println("}")
	}
}
