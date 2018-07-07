// Package stl provides functions to read and write ASCII and binary STL files.
package stl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type STLTri struct {
	Nx, Ny, Nz    float32
	V1x, V1y, V1z float32
	V2x, V2y, V2z float32
	V3x, V3y, V3z float32
}

func Read(filename string) ([]STLTri, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var comment [5]byte
	if err := binary.Read(f, binary.LittleEndian, &comment); err != nil {
		return nil, fmt.Errorf("read comment: %v", err)
	}
	if strings.HasPrefix(string(comment[:]), "solid") {
		return readAscii(f)
	}
	return readBinary(f)
}

func nextLine(buf *bufio.Reader) (string, error) {
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading line: %v", err)
	}
	return strings.TrimSpace(line), nil
}

func threeFloats(p1, p2, p3 string) (float32, float32, float32, error) {
	p1 = strings.Replace(p1, ",", ".", -1)
	p2 = strings.Replace(p2, ",", ".", -1)
	p3 = strings.Replace(p3, ",", ".", -1)
	f1, err := strconv.ParseFloat(p1, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("first float: %v", err)
	}
	f2, err := strconv.ParseFloat(p2, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("second float: %v", err)
	}
	f3, err := strconv.ParseFloat(p3, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("third float: %v", err)
	}
	return float32(f1), float32(f2), float32(f3), nil
}

func getLine(buf *bufio.Reader, s string) error {
	line, err := nextLine(buf)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, s) {
		return fmt.Errorf("expected %q: %q", s, line)
	}
	return nil
}

func getVertex(buf *bufio.Reader) (float32, float32, float32, error) {
	line, err := nextLine(buf)
	if err != nil {
		return 0, 0, 0, err
	}
	if !strings.HasPrefix(line, "vertex ") {
		return 0, 0, 0, fmt.Errorf("expected vertex: %q", line)
	}
	parts := strings.Split(line, " ")
	return threeFloats(parts[1], parts[2], parts[3])
}

func readAscii(f io.Reader) ([]STLTri, error) {
	buf := bufio.NewReader(f)
	_, err := buf.ReadString('\n') // slurp rest of the first line
	if err != nil {
		return nil, fmt.Errorf("error reading first line: %v", err)
	}
	var tris []STLTri
	for {
		line, err := nextLine(buf)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(line, "endsolid") {
			break
		}
		tri := STLTri{}
		if !strings.HasPrefix(line, "facet normal ") {
			return nil, fmt.Errorf("expected facet normal, got %q", line)
		}
		parts := strings.Split(line, " ")
		if len(parts) < 5 {
			return nil, fmt.Errorf("expected 5 parts, got %v: %q", len(parts), line)
		}
		tri.Nx, tri.Ny, tri.Nz, err = threeFloats(parts[2], parts[3], parts[4])
		if err != nil {
			return nil, fmt.Errorf("nx, ny, nz parse: %v", err)
		}
		if err := getLine(buf, "outer loop"); err != nil {
			return nil, err
		}
		tri.V1x, tri.V1y, tri.V1z, err = getVertex(buf)
		if err != nil {
			return nil, fmt.Errorf("v1x, v1y, v1z parse: %v", err)
		}
		tri.V2x, tri.V2y, tri.V2z, err = getVertex(buf)
		if err != nil {
			return nil, fmt.Errorf("v2x, v2y, v2z parse: %v", err)
		}
		tri.V3x, tri.V3y, tri.V3z, err = getVertex(buf)
		if err != nil {
			return nil, fmt.Errorf("v3x, v3y, v3z parse: %v", err)
		}
		if err := getLine(buf, "endloop"); err != nil {
			return nil, err
		}
		if err := getLine(buf, "endfacet"); err != nil {
			return nil, err
		}
		tris = append(tris, tri)
	}
	return tris, nil
}

func readBinary(f io.Reader) ([]STLTri, error) {
	var comment [75]byte // slurp rest of the comment
	if err := binary.Read(f, binary.LittleEndian, &comment); err != nil {
		return nil, fmt.Errorf("read comment: %v", err)
	}
	var numTris uint32
	if err := binary.Read(f, binary.LittleEndian, &numTris); err != nil {
		return nil, fmt.Errorf("read num triangles: %v", err)
	}
	tris := make([]STLTri, numTris)
	for i := 0; i < int(numTris); i++ {
		if err := binary.Read(f, binary.LittleEndian, &tris[i]); err != nil {
			return nil, fmt.Errorf("read triangle #%v: %v", i, err)
		}
		var pad uint16
		if err := binary.Read(f, binary.LittleEndian, &pad); err != nil {
			return nil, fmt.Errorf("read pad #%v: %v", i, err)
		}
	}
	return tris, nil
}
