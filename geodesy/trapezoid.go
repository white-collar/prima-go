package geodesy

import "math"

// TrapezoidArea returns the area of the ellipsoidal trapezoid bounded by the
// parallels B1, B2 and the meridians L1, L2, in km² (orig 1710:002c):
//
//	P = b²·Δl·[ sinB + ⅔e²·sin³B + ⅗e⁴·sin⁵B ] from B1 to B2
func TrapezoidArea(B1, B2, L1, L2 float64) float64 {
	if B1 > B2 {
		B1, B2 = B2, B1
	}
	s1, s2 := math.Sin(B1), math.Sin(B2)
	p3 := func(x float64) float64 { return x * x * x }
	p5 := func(x float64) float64 { return x * x * x * x * x }
	sum := (s2 - s1) + 2.0/3*E2*(p3(s2)-p3(s1)) + 0.6*E2*E2*(p5(s2)-p5(s1))
	return Bax * Bax * math.Abs(L2-L1) * sum / 1e6
}

// Frame holds the sizes of a map-sheet frame on paper, in centimetres.
type Frame struct {
	Side  float64 // western/eastern (meridian) side
	North float64 // northern (parallel) side
	South float64 // southern (parallel) side
}

// TrapezoidFrame computes the frame of a survey trapezoid drawn at scale
// 1:scale (orig 1a8c:07de). Arc lengths are approximated the same way the
// original does it: meridian side = M(Bm)·ΔB, parallel sides = N·cosB·Δl.
func TrapezoidFrame(B1, B2, L1, L2, scale float64) Frame {
	if B1 > B2 {
		B1, B2 = B2, B1
	}
	dl := math.Abs(L2 - L1)
	toCm := 100 / scale
	return Frame{
		Side:  MeridianArcProgram(B1, B2) * toCm,
		North: N(B2) * math.Cos(B2) * dl * toCm,
		South: N(B1) * math.Cos(B1) * dl * toCm,
	}
}
