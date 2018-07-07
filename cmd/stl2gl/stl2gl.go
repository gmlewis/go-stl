// stl2gl converts an STL file to Go code using goxjs/gl.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gmlewis/go-stl/stl"
)

var (
	pkg = flag.String("pkg", "main", "The package of the generated file")
)

func main() {
	flag.Parse()

	fmt.Printf("package %v\n\n", *pkg)
	fmt.Println(`import (
	"encoding/binary"

	"github.com/goxjs/gl"
	"golang.org/x/mobile/exp/f32"
)`)

	for _, arg := range flag.Args() {
		tris, err := stl.Read(arg)
		if err != nil {
			log.Fatalf("stl.Read(%q): %v", arg, err)
		}

		fmt.Println()
		name := strings.TrimSuffix(arg, ".stl")
		fmt.Printf("var %v gl.Buffer\n", name)
		fmt.Printf("var %vNumVerts int\n", name)
		fmt.Printf(`
func init() {
	%[1]v = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, %[1]v)
	%[1]vNumVerts = %[2]v
	vertices := []float32{
`, name, len(tris)*3)
		for _, tri := range tris {
			fmt.Printf("%v, %v, %v, %v, %v, %v, %v, %v, %v,\n", tri.V1x, tri.V1y, tri.V1z, tri.V2x, tri.V2y, tri.V2z, tri.V3x, tri.V3y, tri.V3z)
		}
		fmt.Printf(`	}
	gl.BufferData(gl.ARRAY_BUFFER, f32.Bytes(binary.LittleEndian, vertices...), gl.STATIC_DRAW)
}`)
	}
}
