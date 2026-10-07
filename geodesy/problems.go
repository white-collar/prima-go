package geodesy

import "math"

// DirectResult is the solution of the direct (forward) geodetic problem.
type DirectResult struct {
	B2, L2 float64 // coordinates of the second point
	A21    float64 // reverse azimuth at the second point, [0, 2π)
}

// InverseResult is the solution of the inverse geodetic problem.
type InverseResult struct {
	S   float64 // geodesic distance, m
	A12 float64 // forward azimuth at point 1, [0, 2π)
	A21 float64 // reverse azimuth at point 2, [0, 2π)
	Am  float64 // azimuth at the mid point
}

// DirectSchreiber solves the direct problem by Schreiber's method
// ("способ Шрейбера", orig 2821:0573): the line is first solved on an
// auxiliary sphere of radius N1 and the result is then corrected for the
// ellipsoid. Suitable for distances up to a few hundred kilometres.
func DirectSchreiber(B1, L1, A12, S float64) DirectResult {
	n1 := N(B1)
	u := S * math.Cos(A12) / n1
	v := S * math.Sin(A12) / n1

	b := u * (1 + v*v/3) // spherical latitude increment
	w := v * (1 - u*u/6)
	phi := B1 + b

	t := math.Tan(phi) * w
	l := w / math.Cos(phi)

	dA := t * (1 - l*l/6 - t*t/6) // spherical azimuth increment
	dL := l * (1 - t*t/3)         // longitude difference
	k := w * t * (1 - l*l/12 - t*t/6) / 2
	db := b - k // spherical latitude difference

	v1 := V(B1)
	v2 := v1 * v1
	dB := v2 * db * (1 - 3*EP2*math.Sin(2*B1)*db/4 - EP2*math.Cos(2*B1)*db*db/2)
	corr := b * w / 2 / v2

	A21 := A12 + math.Pi + dA - corr
	if A21 >= TwoP {
		A21 -= TwoP
	}
	return DirectResult{B2: B1 + dB, L2: L1 + dL, A21: A21}
}

// EP2RKM is the value of e'² the original uses inside the Runge–Kutta–Merson
// routine. It is truncated compared with EP2 (0.0067385254), which is how
// PRIMA.EXE was written; it is kept here so results match the original.
const EP2RKM = 0.00673853

// geodesicODE returns dB/ds, dL/ds and dA/ds along a geodesic, scaled by h:
//
//	dB/ds = cosA·V³/c,  dL/ds = sinA·V/(c·cosB),  dA/ds = sinB·dL/ds
func geodesicODE(B, Az, h float64) (dB, dL, dA float64) {
	cb := math.Cos(B)
	v2 := 1 + EP2RKM*cb*cb
	v := math.Sqrt(v2)
	dB = h * math.Cos(Az) * v2 * v / C
	dL = h * math.Sin(Az) * v / (C * cb)
	dA = math.Sin(B) * dL
	return
}

// DirectRKM solves the direct problem by numerically integrating the geodesic
// equations with the Runge–Kutta–Merson method (orig 2821:0c23). Like the
// original, it takes a single Merson step over the whole distance S; pass
// steps > 1 to split the line into equal sub-steps for better accuracy.
func DirectRKM(B1, L1, A12, S float64, steps int) DirectResult {
	if steps < 1 {
		steps = 1
	}
	h := S / float64(steps)
	B, L, Az := B1, L1, A12
	for i := 0; i < steps; i++ {
		// Merson's scheme:
		//   k1 = h·f(y)
		//   k2 = h·f(y + k1/3)
		//   k3 = h·f(y + k1/6 + k2/6)
		//   k4 = h·f(y + k1/8 + 3k3/8)
		//   k5 = h·f(y + k1/2 − 3k3/2 + 2k4)
		//   y' = y + (k1 + 4k4 + k5)/6
		b1, l1, a1 := geodesicODE(B, Az, h)
		b2, _, a2 := geodesicODE(B+b1/3, Az+a1/3, h)
		b3, _, a3 := geodesicODE(B+b1/6+b2/6, Az+a1/6+a2/6, h)
		b4, l4, a4 := geodesicODE(B+b1/8+3*b3/8, Az+a1/8+3*a3/8, h)
		b5, l5, a5 := geodesicODE(B+b1/2-3*b3/2+2*b4, Az+a1/2-3*a3/2+2*a4, h)
		B += (b1 + 4*b4 + b5) / 6
		L += (l1 + 4*l4 + l5) / 6
		Az += (a1 + 4*a4 + a5) / 6
	}
	// Reverse azimuth = forward azimuth at point 2 ± 180°.
	if Az >= math.Pi {
		Az -= math.Pi
	} else {
		Az += math.Pi
	}
	return DirectResult{B2: B, L2: L, A21: Az}
}

// Inverse solves the inverse geodetic problem with Gauss's mid-latitude
// formulas ("формулы со средними аргументами", orig 2821:0899).
// Suitable for distances up to roughly 200 km.
func Inverse(B1, L1, B2, L2 float64) InverseResult {
	b := B2 - B1
	l := L2 - L1
	Bm := (B1 + B2) / 2

	eta2 := Eta2(Bm)
	n := C / math.Sqrt(1+eta2)
	mm := n / (1 + eta2) // = N/V²
	sm, cm := math.Sin(Bm), math.Cos(Bm)
	ls, lc := l*sm, l*cm
	b2 := b * b

	// P = S·cosAm, Q = S·sinAm, dA = A21 − A12 ∓ 180°.
	P := b * mm * (1 - (EP2-2*eta2)*b2/8 - (1+eta2)*lc*lc/12 - (1-2*eta2)*ls*ls/12 - ls*ls/24)
	dA := ls * (1 + (3+2*eta2)*b2/24 + (1+eta2)*lc*lc/12)
	// The original writes the b² coefficient as (1 − 9e'²) + 8η²
	// (stored as the constant 0.9393532714); kept unchanged.
	Q := lc * n * (1 + (8*eta2+(1-9*EP2))*b2/24 - ls*ls/24)

	S := math.Hypot(P, Q)
	Am := azimuth(P, Q)
	A12 := normAz(Am - dA/2)
	A21 := Am + dA/2
	if A21 >= math.Pi {
		A21 -= math.Pi
	} else {
		A21 += math.Pi
	}
	return InverseResult{S: S, A12: A12, A21: A21, Am: Am}
}
