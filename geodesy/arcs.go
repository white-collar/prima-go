package geodesy

import "math"

// MeridianArcProgram is the meridian arc length between latitudes B1 and B2
// exactly as the "Вычисление длин дуг меридианов" screen computes it
// (orig 214c:038c): Sm = M(Bm)·|B2 − B1|, with Bm the mean latitude.
// This is accurate for short arcs only; see MeridianArc for the full series.
func MeridianArcProgram(B1, B2 float64) float64 {
	return math.Abs(B2-B1) * M((B1+B2)/2)
}

// MeridianArc returns the meridian arc length X from the equator to latitude B
// using the series expansion found in orig 2821:20d7 (used by the
// Gauss–Krüger conversion).
func MeridianArc(B float64) float64 {
	m0 := C * math.Sqrt(1-E2) * (1 - E2) // = a(1 − e²)
	m2 := 1.5 * E2 * m0
	m4 := 1.25 * E2 * m2
	m6 := 7.0 / 6 * E2 * m4
	m8 := 1.125 * E2 * m6

	a0 := m0 + m2/2 + 3.0/8*m4 + 5.0/16*m6 + 35.0/128*m8
	a2 := m2/2 + m4/2 + 15.0/32*m6 + 7.0/16*m8
	a4 := m4/8 + 3.0/16*m6 + 7.0/32*m8
	a6 := m6/32 + m8/16

	s, c := math.Sin(B), math.Cos(B)
	s2 := s * s
	return a0*B - s*c*((a2-a4+a6)+(2*a4-16.0/3*a6)*s2+16.0/3*a6*s2*s2)
}

// ParallelArc returns the length of a parallel arc at latitude B between
// longitudes L1 and L2: Sп = N·cosB·|L2 − L1| (orig 1f3c:038c).
func ParallelArc(B, L1, L2 float64) float64 {
	return N(B) * math.Cos(B) * math.Abs(L2-L1)
}
