"""Independent Python reference implementation for the Yarrow demo math."""

from __future__ import annotations

import math
from collections.abc import Sequence

import numpy as np

CDF_LEN = 201
MIN_LOG_PROB = 1e-200


def recency_weights(n: int) -> list[float]:
    """Return weights 1..n for forecasts already ordered oldest to newest."""
    return [float(i) for i in range(1, n + 1)]


def weighted_median_binary(probabilities: Sequence[float], weights: Sequence[float]) -> float:
    if len(probabilities) == 0:
        return 0.5
    if len(probabilities) != len(weights):
        raise ValueError("probabilities and weights must have the same length")
    if len(probabilities) == 1:
        return float(probabilities[0])

    pairs = sorted(zip(probabilities, weights), key=lambda pair: pair[0])
    total = float(sum(weights))
    half = total / 2.0
    cumulative = 0.0
    for probability, weight in pairs:
        cumulative += float(weight)
        if cumulative >= half:
            return float(probability)
    return float(pairs[-1][0])


def weighted_avg_cdf(cdfs: Sequence[Sequence[float]], weights: Sequence[float]) -> list[float]:
    if len(cdfs) == 0:
        return [k / (CDF_LEN - 1) for k in range(CDF_LEN)]
    if len(cdfs) != len(weights):
        raise ValueError("cdfs and weights must have the same length")

    arr = np.asarray(cdfs, dtype=np.float64)
    if arr.ndim != 2 or arr.shape[1] != CDF_LEN:
        got = arr.shape[1] if arr.ndim == 2 else 0
        raise ValueError(f"each CDF must have length {CDF_LEN}, got {got}")

    w = np.asarray(weights, dtype=np.float64)
    agg = (arr * w[:, None]).sum(axis=0) / w.sum()
    np.maximum.accumulate(agg, out=agg)
    return agg.tolist()


def baseline_binary(p_yes: float, resolved_yes: bool) -> float:
    p = p_yes if resolved_yes else 1.0 - p_yes
    p = max(p, MIN_LOG_PROB)
    return 100.0 * math.log(p * 2.0) / math.log(2.0)


def brier_binary(p_yes: float, resolved_yes: bool) -> float:
    outcome = 1.0 if resolved_yes else 0.0
    return (p_yes - outcome) ** 2


def peer_binary(p_self_yes: float, gmp_outcome: float, n: int, resolved_yes: bool) -> float:
    if n < 2:
        return 0.0
    p_self = p_self_yes if resolved_yes else 1.0 - p_self_yes
    p_self = max(p_self, MIN_LOG_PROB)
    gmp = max(gmp_outcome, MIN_LOG_PROB)
    correction = n / (n - 1)
    return 100.0 * correction * math.log(p_self / gmp)


def geometric_mean(xs: Sequence[float]) -> float:
    if len(xs) == 0:
        return 0.0
    log_sum = sum(math.log(max(x, MIN_LOG_PROB)) for x in xs)
    return math.exp(log_sum / len(xs))


def cdf_prob_at(cdf: Sequence[float], k_resolution: int) -> float:
    if len(cdf) != CDF_LEN:
        return 0.0
    if k_resolution <= 0:
        return float(cdf[1] - cdf[0])
    if k_resolution >= CDF_LEN - 1:
        return float(cdf[CDF_LEN - 1] - cdf[CDF_LEN - 2])
    return float((cdf[k_resolution + 1] - cdf[k_resolution - 1]) / 2.0)


def brier_numeric(cdf: Sequence[float], k_resolution: int) -> float:
    if len(cdf) != CDF_LEN:
        return 0.0
    k_resolution = max(0, min(CDF_LEN - 1, k_resolution))
    total = 0.0
    for k in range(CDF_LEN):
        probability = cdf_prob_at(cdf, k)
        outcome = 1.0 if k == k_resolution else 0.0
        total += (probability - outcome) ** 2
    return total


def baseline_for_numeric_question(open_lower: bool, open_upper: bool) -> float:
    open_mass = (0.05 if open_lower else 0.0) + (0.05 if open_upper else 0.0)
    return (1.0 - open_mass) / float(CDF_LEN - 2)


def baseline_numeric(cdf: Sequence[float], k_resolution: int, baseline: float) -> float:
    p = max(cdf_prob_at(cdf, k_resolution), MIN_LOG_PROB)
    baseline = max(baseline, MIN_LOG_PROB)
    return 100.0 * math.log(p / baseline) / 2.0


def peer_numeric(p_self: float, gmp: float, n: int) -> float:
    if n < 2:
        return 0.0
    p_self = max(p_self, MIN_LOG_PROB)
    gmp = max(gmp, MIN_LOG_PROB)
    correction = n / (n - 1)
    return 100.0 * correction * math.log(p_self / gmp) / 2.0


def accumulate_brier(scores: Sequence[float]) -> float:
    if len(scores) == 0:
        return 0.0
    return float(sum(scores) / len(scores))
