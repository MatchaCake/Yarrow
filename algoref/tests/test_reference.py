import math

import pytest

from algoref import reference as ref


def test_baseline_binary_known_values():
    assert ref.baseline_binary(0.5, True) == pytest.approx(0.0)
    assert ref.baseline_binary(1.0, True) == pytest.approx(100.0)


def test_brier_binary_known_value():
    assert ref.brier_binary(0.5, True) == pytest.approx(0.25)


def test_geometric_mean_known_value():
    assert ref.geometric_mean([2, 8]) == pytest.approx(4.0)


def test_weighted_median_binary_known_value():
    assert ref.weighted_median_binary([0.1, 0.5, 0.9], [1, 1, 1]) == pytest.approx(0.5)


def test_cdf_prob_at_uniform_distribution():
    cdf = [k / 200 for k in range(ref.CDF_LEN)]
    assert ref.cdf_prob_at(cdf, 0) == pytest.approx(0.005)
    assert ref.cdf_prob_at(cdf, 100) == pytest.approx(0.005)
    assert ref.cdf_prob_at(cdf, 200) == pytest.approx(0.005)


def test_weighted_avg_cdf_empty_is_linear_prior():
    got = ref.weighted_avg_cdf([], [])
    assert len(got) == ref.CDF_LEN
    assert got[0] == pytest.approx(0.0)
    assert got[100] == pytest.approx(0.5)
    assert got[200] == pytest.approx(1.0)


def test_peer_scores_zero_with_single_forecaster():
    assert ref.peer_binary(0.8, 0.8, 1, True) == pytest.approx(0.0)
    assert ref.peer_numeric(0.2, 0.2, 1) == pytest.approx(0.0)


def test_accumulate_brier_mean_and_empty():
    assert ref.accumulate_brier([0.1, 0.2, 0.3]) == pytest.approx(0.2)
    assert ref.accumulate_brier([]) == pytest.approx(0.0)


def test_invalid_cdf_lengths_raise_for_weighted_avg():
    with pytest.raises(ValueError, match="length 201"):
        ref.weighted_avg_cdf([[0.0, 1.0]], [1.0])


def test_baseline_binary_clamps_zero_probability():
    got = ref.baseline_binary(0.0, True)
    assert math.isfinite(got)
    assert got < -1000
