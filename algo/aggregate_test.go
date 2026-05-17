package algo

import (
	"testing"
	"time"
)

func forecastAt(offset int, p float64) ActiveForecast {
	return ActiveForecast{
		UserID:         "u",
		StartTime:      time.Unix(1700000000+int64(offset), 0).UTC(),
		ProbabilityYes: p,
	}
}

func TestRecencyWeightsReturnsInputOrderWeights(t *testing.T) {
	fs := []ActiveForecast{
		forecastAt(30, 0.3),
		forecastAt(10, 0.1),
		forecastAt(20, 0.2),
	}
	got := RecencyWeights(fs)
	want := []float64{3, 1, 2}
	for i := range want {
		assertNear(t, got[i], want[i])
	}
}

func TestWeightedMedianBinaryKnownValue(t *testing.T) {
	fs := []ActiveForecast{
		forecastAt(10, 0.1),
		forecastAt(20, 0.5),
		forecastAt(30, 0.9),
	}
	got, ok := WeightedMedianBinary(fs)
	if !ok {
		t.Fatal("expected ok")
	}
	assertNear(t, got, 0.5)
}

func TestWeightedAvgCDFEmptyIsLinearPrior(t *testing.T) {
	got, err := WeightedAvgCDF(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != CDFLen {
		t.Fatalf("len = %d, want %d", len(got), CDFLen)
	}
	assertNear(t, got[0], 0)
	assertNear(t, got[100], 0.5)
	assertNear(t, got[200], 1)
}

func TestWeightedAvgCDFRejectsWrongLength(t *testing.T) {
	_, err := WeightedAvgCDF([]ActiveForecast{{StartTime: time.Unix(1700000000, 0), ContinuousCDF: []float64{0, 1}}})
	if err != ErrInvalidCDF {
		t.Fatalf("got %v, want ErrInvalidCDF", err)
	}
}

func linearCDF() []float64 {
	cdf := make([]float64, CDFLen)
	for k := range cdf {
		cdf[k] = float64(k) / float64(CDFLen-1)
	}
	return cdf
}
