// stl2voxels dices an STL file into voxels within a small bounded subregion.
//
// All dimensions are in millimeters unless otherwise noted.
//
// Usage:
//   stl2voxels -mbb "(-100,-100,-3)-(100,100,1)" -dpi 600 -ndx 10 -ndy 10 -ndz 1 -x 0 -y 0 -z 0 ~/Downloads/"Tesla Transformer - Base.stl"
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/gmlewis/go-stl/stl"
)

const (
	inchesPerMM = 1.0 / 25.4
)

var (
	mbbFlag   = flag.String("mbb", "", "minimum bounding box of the entire STL design")
	dpi       = flag.Float64("dpi", 600, "dots per inch resolution of voxels")
	ndx       = flag.Int64("ndx", 10, "number of divisions in X direction")
	ndy       = flag.Int64("ndy", 10, "number of divisions in Y direction")
	ndz       = flag.Int64("ndz", 10, "number of divisions in Z direction")
	srx       = flag.Int64("x", 0, "index of voxel subregion in X direction")
	sry       = flag.Int64("y", 0, "index of voxel subregion in Y direction")
	srz       = flag.Int64("z", 0, "index of voxel subregion in Z direction")
	outPrefix = flag.String("out_prefix", "out-", "output filename prefix for pgm files")

	mbbRE = regexp.MustCompile(`\(([^,]+),([^,]+),([^,]+)\)\-\(([^,]+),([^,]+),([^,]+)\)`)
)

func main() {
	flag.Parse()
	if *ndx < 1 || *ndy < 1 || *ndz < 1 {
		log.Fatalf("Number of dimensions in all directions must be > 0, got (%v,%v,%v)", *ndx, *ndy, *ndz)
	}
	if *srx < 0 || *sry < 0 || *srz < 0 {
		log.Fatalf("Subregion index must be >= 0, got (%v,%v,%v)", *srx, *sry, *srz)
	}
	if *srx >= *ndx || *sry >= *ndy || *srz >= *ndz {
		log.Fatalf("Subregion index must be < (%v,%v,%v), got (%v,%v,%v)", *ndx, *ndy, *ndz, *srx, *sry, *srz)
	}
	if *dpi < 1 {
		log.Fatalf("dpi must be > 0, got %v", *dpi)
	}
	dotsPerMM := *dpi * inchesPerMM

	stlMBB, err := parseMBB()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("design MBB: (%v,%v,%v)-(%v,%v,%v) mm", stlMBB.minX, stlMBB.minY, stlMBB.minZ, stlMBB.maxX, stlMBB.maxY, stlMBB.maxZ)
	dotsX := int(math.Floor((stlMBB.maxX - stlMBB.minX) * dotsPerMM))
	dotsY := int(math.Floor((stlMBB.maxY - stlMBB.minY) * dotsPerMM))
	dotsZ := int(math.Floor((stlMBB.maxZ - stlMBB.minZ) * dotsPerMM))
	log.Printf("size of design: (%v,%v,%v) voxels, (%v,%v,%v) mm", dotsX, dotsY, dotsZ, (stlMBB.maxX - stlMBB.minX), (stlMBB.maxY - stlMBB.minY), (stlMBB.maxZ - stlMBB.minZ))

	nx := int(math.Floor((stlMBB.maxX - stlMBB.minX) * dotsPerMM / float64(*ndx)))
	ny := int(math.Floor((stlMBB.maxY - stlMBB.minY) * dotsPerMM / float64(*ndy)))
	nz := int(math.Floor((stlMBB.maxZ - stlMBB.minZ) * dotsPerMM / float64(*ndz)))
	if nx == 0 || ny == 0 || nz == 0 {
		log.Fatalf("error calculating voxel size: (%v,%v,%v)", nx, ny, nz)
	}

	vSizeX := (stlMBB.maxX - stlMBB.minX) / float64(*ndx)
	vSizeY := (stlMBB.maxY - stlMBB.minY) / float64(*ndy)
	vSizeZ := (stlMBB.maxZ - stlMBB.minZ) / float64(*ndz)
	log.Printf("size of voxel: (%v,%v,%v) voxels, (%v,%v,%v) mm", nx, ny, nz, vSizeX, vSizeY, vSizeZ)

	voxelMBB := &mbb{
		minX: float64(*srx)*vSizeX + stlMBB.minX,
		minY: float64(*sry)*vSizeY + stlMBB.minY,
		minZ: float64(*srz)*vSizeZ + stlMBB.minZ,
		maxX: float64(*srx+1)*vSizeX + stlMBB.minX,
		maxY: float64(*sry+1)*vSizeY + stlMBB.minY,
		maxZ: float64(*srz+1)*vSizeZ + stlMBB.minZ,
	}
	log.Printf("voxel MBB: (%v,%v,%v)-(%v,%v,%v)", voxelMBB.minX, voxelMBB.minY, voxelMBB.minZ, voxelMBB.maxX, voxelMBB.maxY, voxelMBB.maxZ)

	voxels := make([]float64, nx*ny*nz)
	idx := func(x, y, z int) int {
		return (z * nx * ny) + (y * nx) + x
	}
	scaleX := float64(dotsX) / (stlMBB.maxX - stlMBB.minX)
	scaleY := float64(dotsY) / (stlMBB.maxY - stlMBB.minY)
	scaleZ := float64(dotsZ) / (stlMBB.maxZ - stlMBB.minZ)
	voxel := func(x, y, z float32) (nx, ny, nz int) {
		nx = int(math.Floor((float64(x) - stlMBB.minX) * scaleX))
		ny = int(math.Floor((float64(y) - stlMBB.minY) * scaleY))
		nz = int(math.Floor((float64(z) - stlMBB.minZ) * scaleZ))
		return nx, ny, nz
	}

	tris, err := stl.Read(flag.Arg(0))
	if err != nil {
		log.Fatalf("stl.Read(%q): %v", flag.Arg(0), err)
	}

	for _, tri := range tris {
		// render triangle into voxels
		x1, y1, z1 := voxel(tri.V1x, tri.V1y, tri.V1z)
		x2, y2, z2 := voxel(tri.V2x, tri.V2y, tri.V2z)
		x3, y3, z3 := voxel(tri.V3x, tri.V3y, tri.V3z)
		log.Printf("x1,y1,z1=(%v,%v,%v)", x1, y1, z1)
		log.Printf("x2,y2,z2=(%v,%v,%v)", x2, y2, z2)
		log.Printf("x3,y3,z3=(%v,%v,%v)", x3, y3, z3)
	}

	// Write out pixel image files
	for z := 0; z < nz; z++ {
		fileBase := fmt.Sprintf("%v%v-%v-%v-%03d", *outPrefix, *srx, *sry, *srz, z)
		filename := fileBase + ".pgm"
		out, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Unable to create file %v: %v", filename, err)
		}
		fmt.Fprintf(out, "P2\n# %v\n%v %v\n255\n", filename, nx, ny)

		for y := 0; y < ny; y++ {
			for x := 0; x < ny; x++ {
				fmt.Fprintf(out, "%v\n", clamp(voxels[idx(x, y, z)], 255))
			}
		}

		if err := out.Close(); err != nil {
			log.Fatalf("Unable to close file %v: %v", filename, err)
		}
		fmt.Printf("autotrace -input-format PGM -output-file %[1]v.svg %[1]v.pgm\n", fileBase)
	}
	log.Println("Done.")
}

type mbb struct {
	minX, minY, minZ float64
	maxX, maxY, maxZ float64
}

func parseMBB() (*mbb, error) {
	if *mbbFlag == "" {
		return nil, errors.New("must specify MBB of design")
	}
	parts := mbbRE.FindStringSubmatch(*mbbFlag)
	if len(parts) != 7 {
		return nil, fmt.Errorf("Unable to parse -mbb %v", *mbbFlag)
	}

	var err error
	result := &mbb{}
	result.minX, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	result.minY, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, err
	}
	result.minZ, err = strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, err
	}
	result.maxX, err = strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return nil, err
	}
	result.maxY, err = strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return nil, err
	}
	result.maxZ, err = strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func clamp(v float64, maxVal int) int {
	switch {
	case v < 0:
		return 0
	case v >= 1:
		return maxVal
	default:
		return int(math.Floor(0.5 + v*float64(maxVal)))
	}
}
