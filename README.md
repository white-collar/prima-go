# PRIMA — spheroidal geodesy (Go port)

[![Test](https://github.com/white-collar/prima-go/actions/workflows/test.yml/badge.svg)](https://github.com/white-collar/prima-go/actions/workflows/test.yml)

**English** | [Українська](README.uk.md)

A Go rewrite of `PRIMA.EXE`, the DOS program *«Решение задач сфероидической геодезии»*
(Solving problems of spheroidal geodesy). The original is a Borland Pascal 7 program (1992)
using the **Krasovsky 1940** ellipsoid. All formulas were recovered from the executable's
machine code, and the results were checked against the original running in DOSBox-X
(see [Verification](#verification-against-the-original)).

## Original authors

The original program was written by geodesy students as a study project:

| Role | Author |
|---|---|
| Author of the interface and most of the routines | **Бартенева А.В.** (A.V. Barteneva), group ПГ-90 |
| Author of individual routines | **Воронова В.И.** (V.I. Voronova), group ПГ-88 |
| Author of individual routines | **Котолуп Т.Ф.** (T.F. Kotolup), group ПГ-88 |
| Project supervisor | **Гавриленко Ю.Н.** (Yu.N. Gavrilenko), Doctor of Science, professor of the Department of Geoinformatics and Geodesy, [Donetsk National Technical University (DonNTU)](https://donntu.edu.ua) |

The algorithms and the structure of the program are theirs; this repository is a port to Go
made for learning. The original executable is not included here.

## Run

```bash
go run ./cmd/prima
```

```bash
go test ./geodesy -v
```

Angles are typed as `D M S` (`55 45 20.5`); a single number means decimal degrees.

## Layout

| File | What it contains | Original menu item |
|---|---|---|
| `geodesy/ellipsoid.go` | constants, V, W, N, M, R, r, R_A | Радиусы кривизны |
| `geodesy/arcs.go` | meridian arc (short formula and full series), parallel arc | Длины дуг |
| `geodesy/trapezoid.go` | trapezoid area and map-sheet frame | Площадь / рамки трапеций |
| `geodesy/problems.go` | direct problem (Schreiber, Runge–Kutta–Merson), inverse problem (mid-latitude formulas) | Главные геодезические задачи |
| `geodesy/gauss.go` | Gauss–Krüger forward/inverse, meridian convergence, ellipsoid→plane corrections | Перевычисление координат |
| `cmd/prima/main.go` | interactive menu, same structure as the original | — |
| `tools/re/` | the scripts used to extract the formulas from `PRIMA.EXE` | — |
| `tools/dosbox/` | harness that runs `PRIMA.EXE` in DOSBox-X and compares it with the Go port | — |

Each function's comment names the original procedure, e.g. `orig 2821:0573`.

## Differences from the original

The calculations match the original. These are the deliberate differences:

- **No data files or table editor.** The original kept input tables in `*.LB1`…`*.LBB` files. This version is a simple question-and-answer console program.
- **Plane → geodetic keeps the sign of y.** The original passed `|Y − Y0|`, so `l` always came out positive. This port also prints `L = L0 + l`.
- **Runge–Kutta–Merson can take several steps.** The original takes a single step over the whole line. `DirectRKM(..., steps)` lets you split the line; the menu uses 1 step, like the original.
- **Units in the library.** `ReductionToPlane` takes metres; the menu screen takes kilometres, like the original.
- **Division by zero is guarded.** If y_m = 0 the corrections are set to 0; the original would stop with a runtime error.

These quirks of the original are kept so the numbers stay the same, with a comment at each spot in the code:

- the Runge–Kutta routine uses a truncated e'² = 0.00673853;
- in the inverse problem, the Q coefficient is written `(1 − 9e'²) + 8η²`;
- the convergence routine first computes a series, then overwrites it with a closed-form formula. The textbook series is provided separately as `ConvergenceSeries`.

## How the code was extracted (`tools/re`)

1. **MZ header.** `PRIMA.EXE` is a 16-bit DOS executable. `mz.py` parses its header and its segments.
2. **Floating-point emulation.** Turbo Pascal compiles floating-point code as `INT 34h…3Dh` calls to its 8087 emulator. `fix.py` turns those back into real x87 instructions.
3. **Constants.** They are stored in the code as 10-byte `Extended` values, e.g. `6399698.9018`, `0.0067385254`, `6367558.4969`.
4. **Formulas.** `sym2.py` steps through a procedure's x87 stack operations symbolically and prints formulas such as `v28 := (1 - ((va * va) / 6)) * v14`.

```bash
pip install capstone
```

```bash
cd tools/re && python3 sym2.py 2821 0573
```

## Verification against the original

`tools/dosbox/compare.py` runs `PRIMA.EXE` in DOSBox-X and types the test rows from `cases.py` into every screen. It reads each result table back and compares it with the Go port.

The original is driven by `feeder.asm`, a small resident DOS program. It answers the original's keyboard requests (BIOS `INT 16h`) from a key script and saves the text screen to a file at marked points.

**Result.** 47 test rows on all 13 screens match within the original's display rounding: at most 0.0005 in the last printed digit (0.001″, 0.001 m), and 0.005 km² on the area screen, which prints two decimals.

Requirements: put your copy of `PRIMA.EXE` in the repository root (it is not distributed
here), and install DOSBox-X and nasm. The extraction scripts in `tools/re` need it too.

```bash
brew install dosbox-x nasm
```

```bash
python3 tools/dosbox/compare.py
```
