// Command prima is an interactive console version of PRIMA.EXE
// ("Решение задач сфероидической геодезии"). The menu structure follows the
// original program; the calculations live in package prima/geodesy.
//
// Angles are entered as degrees, minutes and seconds separated by spaces
// (e.g. "55 45 20.5"); a single number is taken as decimal degrees.
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	g "github.com/white-collar/prima-go/geodesy"
)

var in = bufio.NewScanner(os.Stdin)

type item struct {
	title string
	run   func()
}

func main() {
	fmt.Println("РЕШЕНИЕ ЗАДАЧ СФЕРОИДИЧЕСКОЙ ГЕОДЕЗИИ — Solving problems of spheroidal geodesy")
	fmt.Println("Ellipsoid: Krasovsky 1940. Go port of PRIMA.EXE.")
	menu("Main menu", []item{
		{"Ellipsoid elements", func() {
			menu("Ellipsoid elements", []item{
				{"Radii of curvature", func() {
					menu("Radii of curvature", []item{
						{"N, M, R, rB at a point", radii},
						{"RA of a normal section with azimuth A", radiusRA},
					})
				}},
				{"Arc lengths of meridians and parallels", func() {
					menu("Arc lengths", []item{
						{"Meridian arc", meridianArc},
						{"Parallel arc", parallelArc},
					})
				}},
				{"Frames of a survey trapezoid", frames},
				{"Area of a survey trapezoid", area},
			})
		}},
		{"Main geodetic problems", func() {
			menu("Main geodetic problems", []item{
				{"Direct problem (Schreiber's method)", func() { direct(false) }},
				{"Direct problem (Runge–Kutta–Merson)", func() { direct(true) }},
				{"Inverse problem (mid-latitude formulas)", inverse},
			})
		}},
		{"Coordinate conversion", func() {
			menu("Coordinate conversion", []item{
				{"Plane (x, y) → geodetic (B, L)", planeToGeo},
				{"Geodetic (B, L) → plane (x, y)", geoToPlane},
				{"Gaussian meridian convergence", convergence},
				{"Corrections for ellipsoid → plane", corrections},
			})
		}},
	})
}

// menu shows numbered items until the user picks 0 or input ends.
func menu(title string, items []item) {
	for {
		fmt.Printf("\n== %s ==\n", title)
		for i, it := range items {
			fmt.Printf("  %d. %s\n", i+1, it.title)
		}
		fmt.Println("  0. Back / exit")
		s, ok := prompt("Choice: ")
		if !ok || s == "0" {
			return
		}
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > len(items) {
			fmt.Println("Unknown item.")
			continue
		}
		items[n-1].run()
	}
}

func prompt(label string) (string, bool) {
	fmt.Print(label)
	if !in.Scan() {
		fmt.Println()
		return "", false
	}
	return strings.TrimSpace(in.Text()), true
}

func readNumbers(label string) []float64 {
	for {
		s, ok := prompt(label)
		if !ok {
			os.Exit(0)
		}
		s = strings.NewReplacer("°", " ", "'", " ", "\"", " ", ":", " ", ",", ".").Replace(s)
		var out []float64
		bad := false
		for _, f := range strings.Fields(s) {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil {
				bad = true
				break
			}
			out = append(out, v)
		}
		if !bad && len(out) > 0 {
			return out
		}
		fmt.Println("  Please enter a number.")
	}
}

func readFloat(label string) float64 { return readNumbers(label)[0] }

func readAngle(label string) float64 {
	for {
		v := readNumbers(label + " (D M S): ")
		switch len(v) {
		case 1:
			return v[0] * math.Pi / 180
		case 2:
			return g.DMS(v[0], v[1], 0)
		case 3:
			return g.DMS(v[0], v[1], v[2])
		}
		fmt.Println("  Enter degrees, minutes and seconds.")
	}
}

func dms(a float64) string { return g.FormatDMS(a, 4) }

func radii() {
	B := readAngle("B")
	fmt.Printf("  N  = %.3f m\n  M  = %.3f m\n  R  = %.3f m\n  rB = %.3f m\n",
		g.N(B), g.M(B), g.R(B), g.ParallelRadius(B))
}

func radiusRA() {
	B := readAngle("B")
	A := readAngle("A")
	fmt.Printf("  RA = %.3f m\n", g.RA(B, A))
}

func meridianArc() {
	B1 := readAngle("B1")
	B2 := readAngle("B2")
	fmt.Printf("  Sm = %.3f m   (original formula M(Bm)·ΔB)\n", g.MeridianArcProgram(B1, B2))
	fmt.Printf("  Sm = %.3f m   (full series X(B2) − X(B1))\n", math.Abs(g.MeridianArc(B2)-g.MeridianArc(B1)))
}

func parallelArc() {
	B := readAngle("B")
	L1 := readAngle("L1")
	L2 := readAngle("L2")
	fmt.Printf("  Sп = %.3f m\n", g.ParallelArc(B, L1, L2))
}

func trapezoid() (B1, B2, L1, L2 float64) {
	return readAngle("B1"), readAngle("B2"), readAngle("L1"), readAngle("L2")
}

func frames() {
	B1, B2, L1, L2 := trapezoid()
	scale := readFloat("Scale denominator (e.g. 10000): ")
	f := g.TrapezoidFrame(B1, B2, L1, L2, scale)
	fmt.Printf("  West/east side = %.2f cm\n  North side     = %.2f cm\n  South side     = %.2f cm\n",
		f.Side, f.North, f.South)
}

func area() {
	B1, B2, L1, L2 := trapezoid()
	fmt.Printf("  P = %.6f km²\n", g.TrapezoidArea(B1, B2, L1, L2))
}

func direct(rkm bool) {
	B1 := readAngle("B1")
	L1 := readAngle("L1")
	A12 := readAngle("A12")
	S := readFloat("S12, m: ")
	var r g.DirectResult
	if rkm {
		r = g.DirectRKM(B1, L1, A12, S, 1) // one step, like the original
	} else {
		r = g.DirectSchreiber(B1, L1, A12, S)
	}
	fmt.Printf("  B2  = %s\n  L2  = %s\n  A21 = %s\n", dms(r.B2), dms(r.L2), dms(r.A21))
}

func inverse() {
	B1 := readAngle("B1")
	L1 := readAngle("L1")
	B2 := readAngle("B2")
	L2 := readAngle("L2")
	r := g.Inverse(B1, L1, B2, L2)
	fmt.Printf("  S12 = %.3f m\n  A12 = %s\n  A21 = %s\n", r.S, dms(r.A12), dms(r.A21))
}

func planeToGeo() {
	x := readFloat("X, m: ")
	y := readFloat("Y, m: ")
	y0 := readFloat("Y0 (false easting, e.g. 500000), m: ")
	L0 := readAngle("L0 (central meridian)")
	B, l := g.PlaneToGeo(x, y-y0)
	fmt.Printf("  B = %s\n  l = %s\n  L = %s\n", dms(B), dms(l), dms(L0+l))
}

func geoToPlane() {
	B := readAngle("B")
	L := readAngle("L")
	L0 := readAngle("L0 (central meridian)")
	x, y := g.GeoToPlane(B, L-L0)
	fmt.Printf("  X = %.3f m\n  Y = %.3f m  (from the central meridian, no false easting)\n", x, y)
}

func convergence() {
	B := readAngle("B")
	L := readAngle("L")
	L0 := readAngle("L0")
	fmt.Printf("  γ = %s\n", dms(g.Convergence(B, L-L0)))
}

func corrections() {
	fmt.Println("  Plane coordinates in kilometres, y measured from the central meridian.")
	x1 := readFloat("x1, km: ")
	y1 := readFloat("y1, km: ")
	x2 := readFloat("x2, km: ")
	y2 := readFloat("y2, km: ")
	c := g.ReductionToPlane(x1*1000, y1*1000, x2*1000, y2*1000)
	fmt.Printf("  δ12 = %.3f\"\n  δ21 = %.3f\"\n  Δs  = %.4f m\n", c.Delta12, c.Delta21, c.DS)
}
