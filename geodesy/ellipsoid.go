// Package geodesy is a Go port of PRIMA.EXE, "Решение задач сфероидической
// геодезии" (Solving problems of spheroidal geodesy), a Borland Pascal 7
// teaching program written by students of groups ПГ-88 and ПГ-90
// (supervisor Yu. N. Gavrilenko).
//
// The formulas were recovered by disassembling the original executable.
// Comments of the form "orig 2821:0442" give the segment:offset of the
// procedure in PRIMA.EXE that each function reproduces.
//
// Conventions used throughout the package:
//   - angles are in radians (use DMS / ToDMS to convert);
//   - linear quantities are in metres unless stated otherwise;
//   - B is geodetic latitude, L is longitude, A is a geodetic azimuth.
package geodesy

import "math"

// Krasovsky 1940 ellipsoid, exactly as the constants are stored in PRIMA.EXE.
const (
	A    = 6378245.0    // semi-major axis a, m
	Bax  = 6356863.0188 // semi-minor axis b, m
	C    = 6399698.9018 // polar radius of curvature c = a²/b, m
	E2   = 0.0066934216 // first eccentricity squared e²
	EP2  = 0.0067385254 // second eccentricity squared e'²
	Rho  = 206265.0     // seconds per radian, as used in the correction formulas
	Xm0  = 6367558.4969 // mean radius used for the footpoint latitude (orig 2821:18f1)
	TwoP = 2 * math.Pi
)

// Eta2 returns η² = e'²·cos²B (orig 2821:19a9).
func Eta2(B float64) float64 {
	c := math.Cos(B)
	return EP2 * c * c
}

// V returns V = √(1 + e'²·cos²B) (orig 2821:03a8).
func V(B float64) float64 { return math.Sqrt(1 + Eta2(B)) }

// W returns W = √(1 − e²·sin²B) (orig 2821:03f7).
func W(B float64) float64 {
	s := math.Sin(B)
	return math.Sqrt(1 - E2*s*s)
}

// N returns the radius of curvature of the prime vertical, N = c/V (orig 2821:0442).
func N(B float64) float64 { return C / V(B) }

// M returns the radius of curvature of the meridian, M = c/V³ (orig 2821:0486).
func M(B float64) float64 {
	v := V(B)
	return C / (v * v * v)
}

// R returns the mean (Gaussian) radius of curvature, R = √(M·N) (orig 257a:03a4).
func R(B float64) float64 { return math.Sqrt(M(B) * N(B)) }

// ParallelRadius returns r = a·cosB / W = N·cosB (orig 2821:04f0).
func ParallelRadius(B float64) float64 { return A * math.Cos(B) / W(B) }

// RA returns the radius of curvature of a normal section with azimuth Az
// (Euler's formula): R_A = M·N / (N·cos²A + M·sin²A) (orig 235d:038c).
func RA(B, Az float64) float64 {
	m, n := M(B), N(B)
	ca, sa := math.Cos(Az), math.Sin(Az)
	return m * n / (n*ca*ca + m*sa*sa)
}
