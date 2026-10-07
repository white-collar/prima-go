package geodesy

import "math"

// GeoToPlane converts geodetic latitude B and longitude difference
// l = L − L0 from the central meridian into Gauss–Krüger plane coordinates
// x (north) and y (east, without the 500 km false easting or zone number)
// (orig 2821:233f).
func GeoToPlane(B, l float64) (x, y float64) {
	n := N(B)
	e2 := Eta2(B)
	s, c := math.Sin(B), math.Cos(B)
	t := s / c
	t2, t4, t6 := t*t, t*t*t*t, math.Pow(t, 6)
	c3, c5, c7 := c*c*c, math.Pow(c, 5), math.Pow(c, 7)
	l2 := l * l

	a2 := n * s * c / 2
	a4 := n * s * c3 / 24 * (5 - t2 + 9*e2 + 4*e2*e2)
	a6 := n * s * c5 / 720 * (61 - 58*t2 + t4 + 270*e2 - 330*e2*t2)
	a8 := n * s * c7 / 40320 * (1385 - 3111*t2 + 543*t4 - t6)

	b1 := n * c
	b3 := n * c3 / 6 * (1 - t2 + e2)
	b5 := n * c5 / 120 * (5 - 18*t2 + t4 + 14*e2 - 58*e2*t2)
	b7 := n * c7 / 5040 * (61 - 479*t2 + 179*t4 - t6)

	x = MeridianArc(B) + l2*(a2+l2*(a4+l2*(a6+l2*a8)))
	y = l * (b1 + l2*(b3+l2*(b5+l2*b7)))
	return
}

// FootpointLatitude returns the latitude Bx of the point on the central
// meridian whose arc length from the equator equals x (orig 2821:18f1).
// It uses the closed working formula for the Krasovsky ellipsoid.
func FootpointLatitude(x float64) float64 {
	beta := x / Xm0
	s, c := math.Sin(beta), math.Cos(beta)
	s2 := s * s
	return beta + s*c*(50517738-(298373-2382*s2)*s2)*1e-10
}

// PlaneToGeo converts Gauss–Krüger plane coordinates (x, y), with y measured
// from the central meridian, into latitude B and longitude difference
// l = L − L0 (orig 2821:1a55).
//
// Note: the original screen passes |Y − Y0| here, so it always returns a
// positive l. This port keeps the sign of y instead.
func PlaneToGeo(x, y float64) (B, l float64) {
	bx := FootpointLatitude(x)
	e2 := Eta2(bx)
	v2 := 1 + e2
	n := N(bx)
	n2, n4 := n*n, n*n*n*n
	t := math.Tan(bx)
	t2, t4, t6 := t*t, t*t*t*t, math.Pow(t, 6)

	b2 := -v2 * t / (2 * n2)
	b4 := -b2 / (12 * n2) * (5 + 3*t2 + e2 - 9*e2*t2 - 4*e2*e2)
	b6 := b2 / (360 * n4) * (61 + 90*t2 + 45*t4 + 46*e2 - 252*e2*t2 - 90*e2*t4)
	b8 := -b2 / (20160 * n4 * n2) * (1385 + 3633*t2 + 4095*t4 + 1575*t6)

	a1 := 1 / (n * math.Cos(bx))
	a3 := -a1 / (6 * n2) * (1 + 2*t2 + e2)
	a5 := a1 / (120 * n4) * (5 + 28*t2 + 24*t4 + 6*e2 + 8*e2*t2)
	a7 := -a1 / (5040 * n4 * n2) * (61 + 662*t2 + 1320*t4 + 720*t6)

	y2 := y * y
	B = bx + y2*(b2+y2*(b4+y2*(b6+y2*b8)))
	l = y * (a1 + y2*(a3+y2*(a5+y2*a7)))
	return
}

// Convergence returns the Gaussian meridian convergence γ at latitude B and
// longitude difference l = L − L0 (orig 2821:124a). The sign of γ follows
// the sign of l.
//
// The original routine first evaluates a series in powers of l and then
// overwrites it with the closed-form expression below, so only this one
// affects the output.
func Convergence(B, l float64) float64 {
	sign := 1.0
	if l < 0 {
		sign, l = -1, -l
	}
	cb := math.Cos(B)
	c2 := cb * cb
	e2 := EP2 * c2
	sb := math.Sin(B)
	k := 1 + 2.0/3*e2 + c2*l*l
	return sign * math.Atan(sb*math.Tan(l)+sb*e2*c2*l*l*l*k)
}

// ConvergenceSeries is the textbook series for the meridian convergence,
//
//	γ = l·sinB·[1 + l²cos²B/3·(1 + 3η² + 2η⁴) + l⁴cos⁴B/15·(2 − tg²B)],
//
// provided for comparison with Convergence.
func ConvergenceSeries(B, l float64) float64 {
	c2 := math.Cos(B) * math.Cos(B)
	e2 := EP2 * c2
	t2 := math.Tan(B) * math.Tan(B)
	l2 := l * l
	return l * math.Sin(B) * (1 + l2*c2/3*(1+3*e2+2*e2*e2) + l2*l2*c2*c2/15*(2-t2))
}

// PlaneCorrections holds the corrections for reducing a line from the
// ellipsoid onto the Gauss–Krüger plane.
type PlaneCorrections struct {
	Delta12 float64 // direction correction at point 1, arc seconds
	Delta21 float64 // direction correction at point 2, arc seconds
	DS      float64 // distance correction s_plane − S_ellipsoid, m
}

// ReductionToPlane computes the corrections for going from the ellipsoid to
// the plane for the line between plane points (x1, y1) and (x2, y2), in
// metres with y measured from the central meridian (orig 2821:1488).
//
// The geodesic length S is obtained by converting both points to B, l and
// solving the inverse problem, exactly as the original does.
func ReductionToPlane(x1, y1, x2, y2 float64) PlaneCorrections {
	const r0 = C / 1000 // km

	B1, l1 := PlaneToGeo(x1, y1)
	B2, l2 := PlaneToGeo(x2, y2)
	xm, ym := (x1+x2)/2, (y1+y2)/2
	Bm, _ := PlaneToGeo(xm, ym)
	S := Inverse(B1, l1, B2, l2).S

	// The direction corrections are evaluated in kilometres.
	dx := (x2 - x1) / 1000
	dy := (y2 - y1) / 1000
	ym /= 1000

	v2 := 1 + Eta2(Bm)
	k := v2 * v2 * Rho / (12 * r0) * ym / r0 * dx // = ρ·Δx·ym / (12 R²)
	q := Rho * EP2 / (2 * r0) * math.Sin(2*Bm) * (ym / r0) * (ym / r0) * dy
	u := (ym / r0) * (ym / r0)
	var d12, d21 float64
	if ym != 0 {
		d12 = -k*(6-2*u-dy/ym) - q
		d21 = k*(6-2*u+dy/ym) + q
	}

	// Scale of the projection along the line (series for cosh(ym/R)).
	rm := R(Bm) / 1000
	w := ym * ym / (rm * rm)
	m := 1 + w/2 + dy*dy/(24*rm*rm) + w*w/24 + w*w*w/720
	return PlaneCorrections{Delta12: d12, Delta21: d21, DS: S*m - S}
}
