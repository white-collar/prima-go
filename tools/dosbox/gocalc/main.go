// Command gocalc reads cases JSON on stdin, prints Go-port results as JSON lists of numbers
// (angles in arc seconds, lengths in metres).
package main

import (
	"encoding/json"
	"math"
	"os"

	g "github.com/white-collar/prima-go/geodesy"
)

const sec = math.Pi / 180 / 3600

func ang(v any) float64 {
	a := v.([]any)
	return g.DMS(a[0].(float64), a[1].(float64), a[2].(float64))
}

func main() {
	var cases map[string][][]any
	json.NewDecoder(os.Stdin).Decode(&cases)
	out := map[string][][]float64{}
	for name, rows := range cases {
		for _, r := range rows {
			var res []float64
			switch name {
			case "radii":
				B := ang(r[0])
				res = []float64{g.N(B), g.M(B), g.R(B), g.ParallelRadius(B)}
			case "RA":
				res = []float64{g.RA(ang(r[0]), ang(r[1]))}
			case "merid":
				res = []float64{g.MeridianArcProgram(ang(r[0]), ang(r[1]))}
			case "paral":
				res = []float64{g.ParallelArc(ang(r[0]), 0, ang(r[1]))}
			case "frames":
				f := g.TrapezoidFrame(ang(r[0]), ang(r[3]), ang(r[1]), ang(r[4]), r[2].(float64))
				res = []float64{f.Side, f.North, f.South}
			case "area":
				res = []float64{g.TrapezoidArea(ang(r[0]), ang(r[1]), ang(r[2]), ang(r[3]))}
			case "schreib", "rkm":
				var d g.DirectResult
				if name == "rkm" {
					d = g.DirectRKM(ang(r[0]), ang(r[1]), ang(r[2]), r[3].(float64), 1)
				} else {
					d = g.DirectSchreiber(ang(r[0]), ang(r[1]), ang(r[2]), r[3].(float64))
				}
				res = []float64{d.B2 / sec, d.L2 / sec, d.A21 / sec}
			case "inverse":
				i := g.Inverse(ang(r[0]), ang(r[2]), ang(r[1]), ang(r[3]))
				res = []float64{i.A12 / sec, i.A21 / sec, i.Am / sec, i.S}
			case "xy2bl":
				// Reproduce the original screen: it passes |Y - Y0|.
				B, l := g.PlaneToGeo(r[0].(float64), math.Abs(r[1].(float64)-r[2].(float64)))
				res = []float64{B / sec, l / sec}
			case "bl2xy":
				x, y := g.GeoToPlane(ang(r[0]), ang(r[1])-ang(r[2]))
				res = []float64{x, y}
			case "gamma":
				res = []float64{g.Convergence(ang(r[0]), ang(r[1])-ang(r[2])) / sec}
			case "corr":
				c := g.ReductionToPlane(r[0].(float64)*1000, r[1].(float64)*1000, r[2].(float64)*1000, r[3].(float64)*1000)
				res = []float64{c.Delta12, c.Delta21, c.DS}
			}
			out[name] = append(out[name], res)
		}
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
