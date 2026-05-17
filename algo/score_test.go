package algo

import (
	"math"
	"testing"
)

func assertNear(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBaselineBinaryKnownValues(t *testing.T) {
	assertNear(t, BaselineBinary(0.5, true), 0)
	assertNear(t, BaselineBinary(1.0, true), 100)
}

func TestBrierBinaryKnownValue(t *testing.T) {
	assertNear(t, BrierBinary(0.5, true), 0.25)
}

func TestGeometricMeanKnownValue(t *testing.T) {
	assertNear(t, GeometricMean([]float64{2, 8}), 4)
}

func TestCDFProbAtUniformDistribution(t *testing.T) {
	cdf := linearCDF()
	assertNear(t, CDFProbAt(cdf, 0), 0.005)
	assertNear(t, CDFProbAt(cdf, 100), 0.005)
	assertNear(t, CDFProbAt(cdf, 200), 0.005)
}

func TestAccumulateBrier(t *testing.T) {
	assertNear(t, AccumulateBrier([]float64{0.1, 0.2, 0.3}), 0.2)
	assertNear(t, AccumulateBrier(nil), 0)
}
