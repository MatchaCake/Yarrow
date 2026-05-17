package algo

import "math"

const minLogProb = 1e-200

func BaselineBinary(pYes float64, resolvedYes bool) float64 {
	p := pYes
	if !resolvedYes {
		p = 1 - pYes
	}
	if p < minLogProb {
		p = minLogProb
	}
	return 100 * math.Log(p*2) / math.Log(2)
}

func BrierBinary(pYes float64, resolvedYes bool) float64 {
	outcome := 0.0
	if resolvedYes {
		outcome = 1
	}
	d := pYes - outcome
	return d * d
}

func PeerBinary(pSelfYes, gmpOutcome float64, n int, resolvedYes bool) float64 {
	if n < 2 {
		return 0
	}
	pSelf := pSelfYes
	if !resolvedYes {
		pSelf = 1 - pSelfYes
	}
	if pSelf < minLogProb {
		pSelf = minLogProb
	}
	if gmpOutcome < minLogProb {
		gmpOutcome = minLogProb
	}
	return 100 * (float64(n) / float64(n-1)) * math.Log(pSelf/gmpOutcome)
}

func GeometricMean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		if x < minLogProb {
			x = minLogProb
		}
		sum += math.Log(x)
	}
	return math.Exp(sum / float64(len(xs)))
}

func CDFProbAt(cdf []float64, kResolution int) float64 {
	if len(cdf) != CDFLen {
		return 0
	}
	if kResolution <= 0 {
		return cdf[1] - cdf[0]
	}
	if kResolution >= CDFLen-1 {
		return cdf[CDFLen-1] - cdf[CDFLen-2]
	}
	return (cdf[kResolution+1] - cdf[kResolution-1]) / 2
}

func BrierNumeric(cdf []float64, kResolution int) float64 {
	if len(cdf) != CDFLen {
		return 0
	}
	if kResolution < 0 {
		kResolution = 0
	}
	if kResolution >= CDFLen {
		kResolution = CDFLen - 1
	}

	sum := 0.0
	for k := 0; k < CDFLen; k++ {
		p := CDFProbAt(cdf, k)
		outcome := 0.0
		if k == kResolution {
			outcome = 1
		}
		d := p - outcome
		sum += d * d
	}
	return sum
}

func BaselineForNumericQuestion(openLower, openUpper bool) float64 {
	openMass := 0.0
	if openLower {
		openMass += 0.05
	}
	if openUpper {
		openMass += 0.05
	}
	return (1 - openMass) / float64(CDFLen-2)
}

func BaselineNumeric(cdf []float64, kResolution int, baseline float64) float64 {
	p := CDFProbAt(cdf, kResolution)
	if p < minLogProb {
		p = minLogProb
	}
	if baseline < minLogProb {
		baseline = minLogProb
	}
	return 100 * math.Log(p/baseline) / 2
}

func PeerNumeric(pSelf, gmp float64, n int) float64 {
	if n < 2 {
		return 0
	}
	if pSelf < minLogProb {
		pSelf = minLogProb
	}
	if gmp < minLogProb {
		gmp = minLogProb
	}
	return 100 * (float64(n) / float64(n-1)) * math.Log(pSelf/gmp) / 2
}

func AccumulateBrier(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}
