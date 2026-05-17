package algo

import (
	"errors"
	"sort"
	"time"
)

const CDFLen = 201

type ActiveForecast struct {
	UserID         string
	StartTime      time.Time
	ProbabilityYes float64
	ContinuousCDF  []float64
}

var ErrInvalidCDF = errors.New("continuous_cdf must have length 201")

func RecencyWeights(fs []ActiveForecast) []float64 {
	if len(fs) == 0 {
		return nil
	}
	type entry struct {
		idx       int
		startTime time.Time
	}
	entries := make([]entry, len(fs))
	for i, forecast := range fs {
		entries[i] = entry{idx: i, startTime: forecast.StartTime}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].startTime.Before(entries[j].startTime)
	})

	weights := make([]float64, len(fs))
	for rank, item := range entries {
		weights[item.idx] = float64(rank + 1)
	}
	return weights
}

func WeightedMedianBinary(fs []ActiveForecast) (median float64, ok bool) {
	if len(fs) == 0 {
		return 0.5, false
	}
	if len(fs) == 1 {
		return fs[0].ProbabilityYes, true
	}

	weights := RecencyWeights(fs)
	type pair struct {
		probability float64
		weight      float64
	}
	pairs := make([]pair, len(fs))
	total := 0.0
	for i, forecast := range fs {
		pairs[i] = pair{probability: forecast.ProbabilityYes, weight: weights[i]}
		total += weights[i]
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].probability < pairs[j].probability
	})

	half := total / 2
	cumulative := 0.0
	for _, item := range pairs {
		cumulative += item.weight
		if cumulative >= half {
			return item.probability, true
		}
	}
	return pairs[len(pairs)-1].probability, true
}

func WeightedAvgCDF(fs []ActiveForecast) ([]float64, error) {
	if len(fs) == 0 {
		out := make([]float64, CDFLen)
		for k := range out {
			out[k] = float64(k) / float64(CDFLen-1)
		}
		return out, nil
	}
	for _, forecast := range fs {
		if len(forecast.ContinuousCDF) != CDFLen {
			return nil, ErrInvalidCDF
		}
	}

	weights := RecencyWeights(fs)
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	out := make([]float64, CDFLen)
	for i, forecast := range fs {
		for k := 0; k < CDFLen; k++ {
			out[k] += weights[i] * forecast.ContinuousCDF[k]
		}
	}
	for k := 0; k < CDFLen; k++ {
		out[k] /= totalWeight
	}
	for k := 1; k < CDFLen; k++ {
		if out[k] < out[k-1] {
			out[k] = out[k-1]
		}
	}
	return out, nil
}
