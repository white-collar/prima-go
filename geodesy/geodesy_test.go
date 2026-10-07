package geodesy

import (
	"math"
	"testing"
)

const sec = math.Pi / 180 / 3600 // one arc second in radians

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.6f, want %.6f (±%g), diff %.3g", name, got, want, tol, got-want)
	}
}

func TestDMS(t *testing.T) {
	r := DMS(55, 45, 20.5)
	d, m, s := ToDMS(r)
	if d != 55 || m != 45 || math.Abs(s-20.5) > 1e-7 {
		t.Errorf("round trip gave %d %d %f", d, m, s)
	}
	near(t, "DMS(-0 30 0)", DMS(0, -30, 0), -0.5*math.Pi/180, 1e-15)
	if got := FormatDMS(DMS(-12, 3, 4.5), 2); got != `-12°03'04.50"` {
		t.Errorf("FormatDMS = %s", got)
	}
}

func TestRadii(t *testing.T) {
	near(t, "N(0)", N(0), A, 1e-3)
	near(t, "M(0)", M(0), A*(1-E2), 1e-2)
	near(t, "N(90)", N(math.Pi/2), C, 1e-6)
	near(t, "M(90)", M(math.Pi/2), C, 1e-6)
	B := DMS(55, 0, 0)
	near(t, "RA(A=0)", RA(B, 0), M(B), 1e-6)
	near(t, "RA(A=90)", RA(B, math.Pi/2), N(B), 1e-6)
	near(t, "r", ParallelRadius(B), N(B)*math.Cos(B), 1e-3)
}

func TestMeridianArc(t *testing.T) {
	// Length of the quarter meridian of the Krasovsky ellipsoid.
	near(t, "X(90°)", MeridianArc(math.Pi/2), 10002137.5, 0.1)
	// The short-arc formula agrees with the series for 1° arcs.
	B1, B2 := DMS(54, 30, 0), DMS(55, 30, 0)
	near(t, "Sm(1°)", MeridianArcProgram(B1, B2), MeridianArc(B2)-MeridianArc(B1), 0.05)
}

func TestTrapezoidArea(t *testing.T) {
	// Area of the whole ellipsoid ≈ 510 083 000 km².
	whole := 2 * TrapezoidArea(0, math.Pi/2, 0, 2*math.Pi)
	near(t, "ellipsoid area, km²", whole, 510083000, 1000)
}

func TestGaussKrugerRoundTrip(t *testing.T) {
	for _, Bd := range []float64{0.5, 30, 50, 55.75, 70} {
		for _, ld := range []float64{-3, -1, 0.5, 2, 3} {
			B, l := Bd*math.Pi/180, ld*math.Pi/180
			x, y := GeoToPlane(B, l)
			B2, l2 := PlaneToGeo(x, y)
			near(t, "B", B2/sec, B/sec, 1e-4)
			near(t, "l", l2/sec, l/sec, 1e-4)
		}
	}
	// On the central meridian x is the meridian arc.
	x, y := GeoToPlane(DMS(50, 0, 0), 0)
	near(t, "x(l=0)", x, MeridianArc(DMS(50, 0, 0)), 1e-6)
	near(t, "y(l=0)", y, 0, 1e-9)
}

func TestConvergence(t *testing.T) {
	for _, Bd := range []float64{10, 45, 60} {
		for _, ld := range []float64{-3, 1, 3} {
			B, l := Bd*math.Pi/180, ld*math.Pi/180
			near(t, "γ", Convergence(B, l)/sec, ConvergenceSeries(B, l)/sec, 0.05)
		}
	}
}

func TestDirectAndInverse(t *testing.T) {
	B1, L1 := DMS(55, 45, 0), DMS(37, 37, 0)
	for _, S := range []float64{1000, 30000, 100000} {
		for _, Ad := range []float64{0, 37, 135, 250, 330} {
			A12 := Ad * math.Pi / 180

			sch := DirectSchreiber(B1, L1, A12, S)
			rkm := DirectRKM(B1, L1, A12, S, 64)
			near(t, "Schreiber vs RKM, B2 ″", sch.B2/sec, rkm.B2/sec, 1e-3)
			near(t, "Schreiber vs RKM, L2 ″", sch.L2/sec, rkm.L2/sec, 1e-3)
			near(t, "Schreiber vs RKM, A21 ″", sch.A21/sec, rkm.A21/sec, 1e-2)

			inv := Inverse(B1, L1, rkm.B2, rkm.L2)
			near(t, "inverse S", inv.S, S, 2e-3*S/1000+1e-3)
			if S > 1000 { // azimuths of very short lines are noisy in arc seconds
				near(t, "inverse A12 ″", angDiff(inv.A12, A12)/sec, 0, 0.02)
				near(t, "inverse A21 ″", angDiff(inv.A21, rkm.A21)/sec, 0, 0.02)
			}
		}
	}
}

func TestRKMSingleStep(t *testing.T) {
	// The original takes one Merson step; for 30 km it is already sub-mm.
	B1, L1, A12 := DMS(50, 0, 0), DMS(30, 0, 0), DMS(45, 0, 0)
	one := DirectRKM(B1, L1, A12, 30000, 1)
	many := DirectRKM(B1, L1, A12, 30000, 100)
	near(t, "B2 ″", one.B2/sec, many.B2/sec, 1e-4)
}

func TestReductionToPlane(t *testing.T) {
	// Line near the central meridian: corrections are small.
	c := ReductionToPlane(6000000, 10000, 6010000, 12000)
	if math.Abs(c.Delta12) > 1 || math.Abs(c.DS) > 0.02 {
		t.Errorf("unexpectedly large corrections near the axial meridian: %+v", c)
	}
	// 10 km line 200 km from the axial meridian: ds ≈ S·ym²/(2R²) ≈ 4.9 m.
	c = ReductionToPlane(6000000, 200000, 6010000, 200000)
	near(t, "ds", c.DS, 4.93, 0.05)
	// δ12 ≈ −ρ·Δx·ym/(2R²) ≈ −5.07″ and δ21 ≈ −δ12.
	near(t, "δ12", c.Delta12, -5.07, 0.05)
	near(t, "δ21", c.Delta21, -c.Delta12, 1e-6)
}

func angDiff(a, b float64) float64 {
	d := math.Mod(a-b+3*math.Pi, 2*math.Pi) - math.Pi
	return d
}
