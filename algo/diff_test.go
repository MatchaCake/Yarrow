package algo

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

const fixtureTolerance = 1e-6

func TestDiff_ReferenceFixtures(t *testing.T) {
	root := findRepoRoot(t)
	inputDir := filepath.Join(root, "fixtures", "input")
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) < 60 {
		t.Fatalf("expected at least 60 fixtures, got %d", len(names))
	}

	for _, name := range names {
		t.Run(strings.TrimSuffix(name, ".json"), func(t *testing.T) {
			runDiffFixture(t, root, name)
		})
	}
}

func runDiffFixture(t *testing.T, root, name string) {
	t.Helper()
	input := loadMap(t, filepath.Join(root, "fixtures", "input", name))
	expected := loadMap(t, filepath.Join(root, "fixtures", "expected", name))

	switch input["op"] {
	case "baseline_binary":
		assertFixtureScalar(t, BaselineBinary(asFloat(input["probability_yes"]), asBool(input["resolved_yes"])), expected, "score")
	case "brier_binary":
		assertFixtureScalar(t, BrierBinary(asFloat(input["probability_yes"]), asBool(input["resolved_yes"])), expected, "score")
	case "peer_binary":
		assertFixtureScalar(t, PeerBinary(asFloat(input["probability_yes_self"]), asFloat(input["gmp_outcome"]), asInt(input["n"]), asBool(input["resolved_yes"])), expected, "score")
	case "geometric_mean":
		assertFixtureScalar(t, GeometricMean(asFloatSlice(input["values"])), expected, "value")
	case "cdf_prob_at":
		assertFixtureScalar(t, CDFProbAt(asFloatSlice(input["cdf"]), asInt(input["k_resolution"])), expected, "probability")
	case "brier_numeric":
		assertFixtureScalar(t, BrierNumeric(asFloatSlice(input["cdf"]), asInt(input["k_resolution"])), expected, "score")
	case "baseline_for_numeric_question":
		assertFixtureScalar(t, BaselineForNumericQuestion(asBool(input["open_lower"]), asBool(input["open_upper"])), expected, "baseline")
	case "baseline_numeric":
		assertFixtureScalar(t, BaselineNumeric(asFloatSlice(input["cdf"]), asInt(input["k_resolution"]), asFloat(input["baseline"])), expected, "score")
	case "peer_numeric":
		assertFixtureScalar(t, PeerNumeric(asFloat(input["p_self"]), asFloat(input["gmp"]), asInt(input["n"])), expected, "score")
	case "recency_weights":
		got := RecencyWeights(forecastsFromStartOffsets(asIntSlice(input["start_offsets"])))
		assertFixtureSlice(t, got, expected, "weights")
	case "weighted_median_binary":
		fs := forecastsFromProbabilitiesAndWeights(asFloatSlice(input["probabilities"]), asFloatSlice(input["weights"]))
		got, _ := WeightedMedianBinary(fs)
		assertFixtureScalar(t, got, expected, "median")
	case "weighted_avg_cdf":
		fs := cdfForecastsFromWeights(asCDFs(input["cdfs"]), asFloatSlice(input["weights"]))
		got, err := WeightedAvgCDF(fs)
		if err != nil {
			t.Fatal(err)
		}
		assertFixtureSlice(t, got, expected, "cdf")
	case "accumulate_brier":
		assertFixtureScalar(t, AccumulateBrier(asFloatSlice(input["scores"])), expected, "score")
	default:
		t.Fatalf("unknown op %q", input["op"])
	}
}

func loadMap(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return out
}

func assertFixtureScalar(t *testing.T, got float64, expected map[string]any, key string) {
	t.Helper()
	want := asFloat(expected[key])
	if math.Abs(got-want) > fixtureTolerance {
		t.Fatalf("%s: got %v, want %v, diff %g", key, got, want, got-want)
	}
}

func assertFixtureSlice(t *testing.T, got []float64, expected map[string]any, key string) {
	t.Helper()
	want := asFloatSlice(expected[key])
	if len(got) != len(want) {
		t.Fatalf("%s length: got %d, want %d", key, len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > fixtureTolerance {
			t.Fatalf("%s[%d]: got %v, want %v, diff %g", key, i, got[i], want[i], got[i]-want[i])
		}
	}
}

func forecastsFromStartOffsets(offsets []int) []ActiveForecast {
	fs := make([]ActiveForecast, len(offsets))
	for i, offset := range offsets {
		fs[i] = ActiveForecast{UserID: "u", StartTime: time.Unix(1700000000+int64(offset), 0).UTC()}
	}
	return fs
}

func forecastsFromProbabilitiesAndWeights(probabilities, weights []float64) []ActiveForecast {
	fs := make([]ActiveForecast, len(probabilities))
	for rank, idx := range orderByWeight(weights) {
		fs[idx] = ActiveForecast{
			UserID:         "u",
			StartTime:      time.Unix(1700000000+int64(rank), 0).UTC(),
			ProbabilityYes: probabilities[idx],
		}
	}
	return fs
}

func cdfForecastsFromWeights(cdfs [][]float64, weights []float64) []ActiveForecast {
	fs := make([]ActiveForecast, len(cdfs))
	for rank, idx := range orderByWeight(weights) {
		fs[idx] = ActiveForecast{
			UserID:        "u",
			StartTime:     time.Unix(1700000000+int64(rank), 0).UTC(),
			ContinuousCDF: cdfs[idx],
		}
	}
	return fs
}

func orderByWeight(weights []float64) []int {
	idxs := make([]int, len(weights))
	for i := range idxs {
		idxs[i] = i
	}
	sort.SliceStable(idxs, func(i, j int) bool { return weights[idxs[i]] < weights[idxs[j]] })
	return idxs
}

func asFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	}
	return 0
}

func asInt(v any) int {
	return int(asFloat(v))
}

func asBool(v any) bool {
	got, _ := v.(bool)
	return got
}

func asFloatSlice(v any) []float64 {
	raw, _ := v.([]any)
	out := make([]float64, len(raw))
	for i, item := range raw {
		out[i] = asFloat(item)
	}
	return out
}

func asIntSlice(v any) []int {
	raw, _ := v.([]any)
	out := make([]int, len(raw))
	for i, item := range raw {
		out[i] = asInt(item)
	}
	return out
}

func asCDFs(v any) [][]float64 {
	raw, _ := v.([]any)
	out := make([][]float64, len(raw))
	for i, item := range raw {
		out[i] = asFloatSlice(item)
	}
	return out
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "fixtures")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("fixtures directory not found")
		}
		dir = parent
	}
}
