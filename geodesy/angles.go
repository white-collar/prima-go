package geodesy

import (
	"fmt"
	"math"
)

// DMS converts degrees, minutes and seconds to radians (orig 2821:0282).
// The sign of the result is taken from the first non-zero component, so
// -30°15' can be written as DMS(-30, 15, 0) or DMS(0, -15, 0).
func DMS(deg, min, sec float64) float64 {
	sign := 1.0
	if deg < 0 || (deg == 0 && min < 0) || (deg == 0 && min == 0 && sec < 0) {
		sign = -1
	}
	d := math.Abs(deg) + (math.Abs(min)+math.Abs(sec)/60)/60
	return sign * d * math.Pi / 180
}

// ToDMS splits an angle in radians into degrees, minutes and seconds
// (orig 2821:0309). For negative angles all three parts are negative or zero.
func ToDMS(rad float64) (deg, min int, sec float64) {
	sign := 1
	if rad < 0 {
		sign, rad = -1, -rad
	}
	d := rad * 180 / math.Pi
	deg = int(d)
	m := (d - float64(deg)) * 60
	min = int(m)
	sec = (m - float64(min)) * 60
	// Guard against 59.99999… rounding up to 60.0 when printed.
	if sec >= 59.9999995 {
		sec, min = 0, min+1
		if min == 60 {
			min, deg = 0, deg+1
		}
	}
	return sign * deg, sign * min, float64(sign) * sec
}

// FormatDMS prints an angle as D°MM'SS.ssss" with the given number of
// decimals in the seconds.
func FormatDMS(rad float64, decimals int) string {
	d, m, s := ToDMS(math.Abs(rad))
	sign := ""
	if rad < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%d°%02d'%0*.*f\"", sign, d, m, decimals+3, decimals, s)
}

// normAz brings an azimuth into [0, 2π).
func normAz(a float64) float64 {
	a = math.Mod(a, TwoP)
	if a < 0 {
		a += TwoP
	}
	return a
}

// azimuth returns the direction of the vector (dx north, dy east) in [0, 2π).
// It is the quadrant-resolving arctangent used by the inverse problem
// (orig 2821:002c).
func azimuth(dx, dy float64) float64 { return normAz(math.Atan2(dy, dx)) }
