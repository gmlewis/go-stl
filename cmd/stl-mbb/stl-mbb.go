// stl-mbb calculates the minimum bounding box (MBB) of an STL file.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gmlewis/go-stl/v2/stl"
)

func main() {
	flag.Parse()

	tris, err := stl.Read(flag.Arg(0))
	if err != nil {
		log.Fatalf("stl.Read(%q): %v", flag.Arg(0), err)
	}

	var minX, minY, minZ, maxX, maxY, maxZ float32
	for i, tri := range tris {
		if i == 0 || tri.V1x < minX {
			minX = tri.V1x
		}
		if i == 0 || tri.V2x < minX {
			minX = tri.V2x
		}
		if i == 0 || tri.V3x < minX {
			minX = tri.V3x
		}
		if i == 0 || tri.V1x > maxX {
			maxX = tri.V1x
		}
		if i == 0 || tri.V2x > maxX {
			maxX = tri.V2x
		}
		if i == 0 || tri.V3x > maxX {
			maxX = tri.V3x
		}

		if i == 0 || tri.V1y < minY {
			minY = tri.V1y
		}
		if i == 0 || tri.V2y < minY {
			minY = tri.V2y
		}
		if i == 0 || tri.V3y < minY {
			minY = tri.V3y
		}
		if i == 0 || tri.V1y > maxY {
			maxY = tri.V1y
		}
		if i == 0 || tri.V2y > maxY {
			maxY = tri.V2y
		}
		if i == 0 || tri.V3y > maxY {
			maxY = tri.V3y
		}

		if i == 0 || tri.V1z < minZ {
			minZ = tri.V1z
		}
		if i == 0 || tri.V2z < minZ {
			minZ = tri.V2z
		}
		if i == 0 || tri.V3z < minZ {
			minZ = tri.V3z
		}
		if i == 0 || tri.V1z > maxZ {
			maxZ = tri.V1z
		}
		if i == 0 || tri.V2z > maxZ {
			maxZ = tri.V2z
		}
		if i == 0 || tri.V3z > maxZ {
			maxZ = tri.V3z
		}
	}
	fmt.Printf("(%g,%g,%g)-(%g,%g,%g)\n", minX, minY, minZ, maxX, maxY, maxZ)
}
