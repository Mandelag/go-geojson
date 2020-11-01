package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	// "strconv"
	// "strings"

	"github.com/golang/geo/s2"
)

var up = [3]byte{27, 91, 65}
var down = [3]byte{27, 91, 66}
var right = [3]byte{27, 91, 67}
var left = [3]byte{27, 91, 68}

func main() {
	var geo GeoJSON

	f, err := os.Open("jakarta_4326.geojson")
	if err != nil {
		log.Println(err)
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		return
	}

	f.Close()

	err = json.Unmarshal(b, &geo)
	if err != nil {
		log.Println(err)
		return
	}

	PrintMemUsage()
	query := TestTestGeometry(geo)
	PrintMemUsage()
	runtime.GC()
	PrintMemUsage()

	fmt.Println("Please enter <longitude>,<latitude>")
	lat := -6.3
	lon := 106.85

	var buf [3]byte
	// sc := bufio.NewScanner(os.Stdin)
	// sc.Split(bufio.)
	// Disable input buffering first
	// stty -F /dev/tty cbreak min 1
	// https://stackoverflow.com/questions/15159118/read-a-character-from-standard-input-in-go-without-pressing-enter
	for {
		os.Stdin.Read(buf[:])

		switch buf {
		case up:
			// fmt.Println("UP")
			lat += 0.005
		case down:
			// fmt.Println("DOWN")
			lat -= 0.005
		case right:
			lon += 0.005
			// fmt.Println("RIGHT")
		case left:
			lon -= 0.005
			// fmt.Println("LEFT")
		}

		point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
		sps := query.ContainingShapes(point)

		if len(sps) == 0 {
			fmt.Printf("You're at %.3f %.3f %+v\n", lon, lat, "-")
		}
		for _, sp := range sps {
			w, ok := sp.(PolygonWithAttributes)
			if !ok {
				log.Println("Its not oke!")
			}
			fmt.Printf("You're at %.3f %.3f %+v\n", lon, lat, w.Attributes["KEL_NAME"])
			// fmt.Printf("  %+v\n", w.Attributes["KEL_NAME"])
		}
	}

}

func TestTestGeometry(geo GeoJSON) *s2.ContainsPointQuery {
	index := s2.NewShapeIndex()
	attributes := make(map[int32]map[string]string)

	log.Println(len(geo.Features))
	// var invertcount int

	for _, f := range geo.Features {
		// iterate polygon in multi-polygon
		for _, polygon := range f.Geometry.Coordinates {
			loops := make([]*s2.Loop, 0, len(polygon))

			// iterate linear ring in the polygon
			//
			// According to https://tools.ietf.org/html/rfc7946#section-3.1.6 ,
			// the first is the exterior ring, while the rest of MUST BE the interior ring.
			//
			// The linear ring in GeoJSON must respect right hand rule:
			// ie. exterior ring CCW (Counter Clock Wise), then the interior ring CW (Clock Wise).
			for _, linearRing := range polygon {
				points := make([]s2.Point, 0, len(linearRing))

				// because golang geo cannot allow duplicate point but the RFC-7946 did not,
				// we need to de duplicate ourself
				pointExist := make(map[string]map[string]struct{})

				for _, coordinates := range linearRing {
					lon := float64(coordinates[0])
					lat := float64(coordinates[1])

					point := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))

					latTrunc := fmt.Sprintf("%.3f", lat)
					lonTrunc := fmt.Sprintf("%.3f", lon)

					if _, ok := pointExist[latTrunc]; !ok {
						pointExist[latTrunc] = make(map[string]struct{})
					}

					if _, ok := pointExist[latTrunc][lonTrunc]; !ok {
						points = append(points, point)
						pointExist[latTrunc][lonTrunc] = struct{}{}
					}
				}

				loop := s2.LoopFromPoints(points)

				areaKM := loop.Area() * math.Pow(6378.1370, 2)

				// filter out geometry that doesn't have area
				if areaKM < 0.1 {
					continue
				}

				// because my GeoJSON data uses CW for the first linear ring of the polygon, and CCW for the rest;
				// while the geo library `s2.PolygonFromOrientedLoops` function expect CCW for the first, and CW for the holes.
				loop.Invert()

				if err := loop.Validate(); err != nil {
					log.Println("NOT VALIID", err)
					continue
				}

				loops = append(loops, loop)
			}

			polygon := s2.PolygonFromOrientedLoops(loops)

			wrapper := PolygonWithAttributes{
				Polygon:    polygon,
				Attributes: f.Properties,
			}

			shapeID := index.Add(wrapper)
			attributes[shapeID] = f.Properties
		}
	}
	index.Build()

	cpq := s2.NewContainsPointQuery(index, s2.VertexModelClosed)
	// sps := cpq.ContainingShapes(s2.PointFromLatLng(s2.LatLngFromDegrees(-6.35, 106.83)))
	return cpq
}

func invert(points []s2.Point) {
	for i := 0; i < len(points)/2; i++ {
		points[i] = points[len(points)-1]
	}
}

type GeoJSON struct {
	Type     string
	Name     string
	CRS      CRS
	Features []MultiPolygonFeature
}

type CRS struct {
	Type       string
	Properties map[string]string
}

// specificaly only to multi polygon
type MultiPolygonFeature struct {
	Type       string
	Properties map[string]string
	Geometry   MultiPolygon
}

type MultiPolygon struct {
	Type        string
	Coordinates [][][][2]float32
}

type PolygonWithAttributes struct {
	*s2.Polygon
	Attributes map[string]string
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
