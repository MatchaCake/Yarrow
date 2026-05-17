"""Generate JSON fixtures from the Python reference implementation."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

if __package__ in {None, ""}:
    sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from algoref import reference as ref


def linear_cdf() -> list[float]:
    return [k / (ref.CDF_LEN - 1) for k in range(ref.CDF_LEN)]


def step_cdf(at: int) -> list[float]:
    return [0.0 if k < at else 1.0 for k in range(ref.CDF_LEN)]


def ramp_cdf(power: float) -> list[float]:
    return [(k / (ref.CDF_LEN - 1)) ** power for k in range(ref.CDF_LEN)]


def smooth_step_cdf(center: int, width: int) -> list[float]:
    out: list[float] = []
    low = center - width
    high = center + width
    for k in range(ref.CDF_LEN):
        if k <= low:
            out.append(0.0)
        elif k >= high:
            out.append(1.0)
        else:
            out.append((k - low) / (high - low))
    return out


def spec(name: str, op: str, **fields: Any) -> dict[str, Any]:
    return {"name": name, "op": op, **fields}


def build_specs() -> list[dict[str, Any]]:
    flat = linear_cdf()
    sharp_low = step_cdf(1)
    sharp_mid = step_cdf(100)
    sharp_high = step_cdf(199)
    wide_left = ramp_cdf(0.6)
    wide_right = ramp_cdf(1.8)
    smooth_mid = smooth_step_cdf(100, 20)
    smooth_low = smooth_step_cdf(35, 15)

    specs: list[dict[str, Any]] = []

    specs += [
        spec("baseline_binary_mid_yes", "baseline_binary", probability_yes=0.5, resolved_yes=True),
        spec("baseline_binary_perfect_yes", "baseline_binary", probability_yes=1.0, resolved_yes=True),
        spec("baseline_binary_perfect_no", "baseline_binary", probability_yes=0.0, resolved_yes=False),
        spec("baseline_binary_extreme_hit", "baseline_binary", probability_yes=0.999, resolved_yes=True),
        spec("baseline_binary_extreme_miss", "baseline_binary", probability_yes=0.001, resolved_yes=True),
        spec("baseline_binary_regular_no", "baseline_binary", probability_yes=0.73, resolved_yes=False),
    ]
    specs += [
        spec("brier_binary_mid_yes", "brier_binary", probability_yes=0.5, resolved_yes=True),
        spec("brier_binary_perfect_yes", "brier_binary", probability_yes=1.0, resolved_yes=True),
        spec("brier_binary_perfect_no", "brier_binary", probability_yes=0.0, resolved_yes=False),
        spec("brier_binary_extreme_hit", "brier_binary", probability_yes=0.999, resolved_yes=True),
        spec("brier_binary_extreme_miss", "brier_binary", probability_yes=0.001, resolved_yes=True),
        spec("brier_binary_regular_no", "brier_binary", probability_yes=0.73, resolved_yes=False),
    ]
    specs += [
        spec("peer_binary_n1", "peer_binary", probability_yes_self=0.7, gmp_outcome=0.5, n=1, resolved_yes=True),
        spec("peer_binary_n2_match", "peer_binary", probability_yes_self=0.6, gmp_outcome=0.6, n=2, resolved_yes=True),
        spec("peer_binary_n2_edge", "peer_binary", probability_yes_self=0.8, gmp_outcome=0.5, n=2, resolved_yes=True),
        spec("peer_binary_n5_no", "peer_binary", probability_yes_self=0.2, gmp_outcome=0.7, n=5, resolved_yes=False),
        spec("peer_binary_n50_hit", "peer_binary", probability_yes_self=0.999, gmp_outcome=0.55, n=50, resolved_yes=True),
        spec("peer_binary_n50_miss", "peer_binary", probability_yes_self=0.001, gmp_outcome=0.45, n=50, resolved_yes=True),
    ]
    specs += [
        spec("geometric_mean_two", "geometric_mean", values=[2.0, 8.0]),
        spec("geometric_mean_probs", "geometric_mean", values=[0.1, 0.9]),
        spec("geometric_mean_three", "geometric_mean", values=[0.2, 0.5, 0.8]),
        spec("geometric_mean_extreme_low", "geometric_mean", values=[0.001, 0.5, 0.999]),
        spec("geometric_mean_many", "geometric_mean", values=[0.15, 0.35, 0.55, 0.75, 0.95]),
        spec("geometric_mean_zero_clamp", "geometric_mean", values=[0.0, 0.25, 0.5]),
    ]
    specs += [
        spec("cdf_prob_at_flat_low", "cdf_prob_at", cdf=flat, k_resolution=0),
        spec("cdf_prob_at_flat_mid", "cdf_prob_at", cdf=flat, k_resolution=100),
        spec("cdf_prob_at_flat_high", "cdf_prob_at", cdf=flat, k_resolution=200),
        spec("cdf_prob_at_step_low", "cdf_prob_at", cdf=sharp_low, k_resolution=1),
        spec("cdf_prob_at_step_mid", "cdf_prob_at", cdf=sharp_mid, k_resolution=100),
        spec("cdf_prob_at_wide", "cdf_prob_at", cdf=wide_left, k_resolution=80),
    ]
    specs += [
        spec("brier_numeric_flat_low", "brier_numeric", cdf=flat, k_resolution=0),
        spec("brier_numeric_flat_mid", "brier_numeric", cdf=flat, k_resolution=100),
        spec("brier_numeric_flat_high", "brier_numeric", cdf=flat, k_resolution=200),
        spec("brier_numeric_sharp_hit", "brier_numeric", cdf=sharp_mid, k_resolution=100),
        spec("brier_numeric_sharp_miss", "brier_numeric", cdf=sharp_mid, k_resolution=10),
        spec("brier_numeric_wide", "brier_numeric", cdf=wide_right, k_resolution=150),
    ]
    specs += [
        spec("baseline_for_numeric_closed", "baseline_for_numeric_question", open_lower=False, open_upper=False),
        spec("baseline_for_numeric_lower_open", "baseline_for_numeric_question", open_lower=True, open_upper=False),
        spec("baseline_for_numeric_upper_open", "baseline_for_numeric_question", open_lower=False, open_upper=True),
        spec("baseline_for_numeric_both_open", "baseline_for_numeric_question", open_lower=True, open_upper=True),
        spec("baseline_for_numeric_repeat_closed", "baseline_for_numeric_question", open_lower=False, open_upper=False),
        spec("baseline_for_numeric_repeat_both", "baseline_for_numeric_question", open_lower=True, open_upper=True),
    ]
    specs += [
        spec("baseline_numeric_flat_mid", "baseline_numeric", cdf=flat, k_resolution=100, baseline=ref.baseline_for_numeric_question(False, False)),
        spec("baseline_numeric_flat_low_open", "baseline_numeric", cdf=flat, k_resolution=0, baseline=ref.baseline_for_numeric_question(True, False)),
        spec("baseline_numeric_flat_high_open", "baseline_numeric", cdf=flat, k_resolution=200, baseline=ref.baseline_for_numeric_question(False, True)),
        spec("baseline_numeric_sharp_hit", "baseline_numeric", cdf=sharp_mid, k_resolution=100, baseline=ref.baseline_for_numeric_question(False, False)),
        spec("baseline_numeric_sharp_miss", "baseline_numeric", cdf=sharp_mid, k_resolution=10, baseline=ref.baseline_for_numeric_question(False, False)),
        spec("baseline_numeric_wide_right", "baseline_numeric", cdf=wide_right, k_resolution=150, baseline=ref.baseline_for_numeric_question(True, True)),
    ]
    specs += [
        spec("peer_numeric_n1", "peer_numeric", p_self=0.2, gmp=0.2, n=1),
        spec("peer_numeric_n2_match", "peer_numeric", p_self=0.25, gmp=0.25, n=2),
        spec("peer_numeric_n2_edge", "peer_numeric", p_self=0.4, gmp=0.1, n=2),
        spec("peer_numeric_n5_low", "peer_numeric", p_self=0.02, gmp=0.1, n=5),
        spec("peer_numeric_n50_high", "peer_numeric", p_self=0.75, gmp=0.3, n=50),
        spec("peer_numeric_zero_clamp", "peer_numeric", p_self=0.0, gmp=0.2, n=3),
    ]
    specs += [
        spec("recency_weights_empty", "recency_weights", start_offsets=[]),
        spec("recency_weights_single", "recency_weights", start_offsets=[10]),
        spec("recency_weights_sorted_three", "recency_weights", start_offsets=[10, 20, 30]),
        spec("recency_weights_mixed_order", "recency_weights", start_offsets=[30, 10, 20]),
        spec("recency_weights_ties", "recency_weights", start_offsets=[20, 20, 10, 30]),
        spec("recency_weights_many", "recency_weights", start_offsets=[50, 10, 40, 20, 30, 60]),
    ]
    specs += [
        spec("weighted_median_binary_empty", "weighted_median_binary", probabilities=[], weights=[]),
        spec("weighted_median_binary_single", "weighted_median_binary", probabilities=[0.7], weights=[1.0]),
        spec("weighted_median_binary_equal_weights", "weighted_median_binary", probabilities=[0.1, 0.5, 0.9], weights=[1.0, 1.0, 1.0]),
        spec("weighted_median_binary_old_majority", "weighted_median_binary", probabilities=[0.1, 0.1, 0.9], weights=[1.0, 2.0, 3.0]),
        spec("weighted_median_binary_new_majority", "weighted_median_binary", probabilities=[0.1, 0.9, 0.9, 0.9], weights=[1.0, 2.0, 3.0, 4.0]),
        spec("weighted_median_binary_unsorted_probs", "weighted_median_binary", probabilities=[0.8, 0.2, 0.6, 0.4, 0.9], weights=[5.0, 1.0, 3.0, 2.0, 4.0]),
    ]
    specs += [
        spec("weighted_avg_cdf_empty", "weighted_avg_cdf", cdfs=[], weights=[]),
        spec("weighted_avg_cdf_single", "weighted_avg_cdf", cdfs=[flat], weights=[1.0]),
        spec("weighted_avg_cdf_two", "weighted_avg_cdf", cdfs=[flat, sharp_mid], weights=[1.0, 2.0]),
        spec("weighted_avg_cdf_three", "weighted_avg_cdf", cdfs=[wide_left, smooth_mid, sharp_high], weights=[1.0, 2.0, 3.0]),
        spec("weighted_avg_cdf_wide", "weighted_avg_cdf", cdfs=[wide_left, wide_right], weights=[2.0, 1.0]),
        spec("weighted_avg_cdf_smooth_low", "weighted_avg_cdf", cdfs=[smooth_low, flat, sharp_low], weights=[2.0, 3.0, 1.0]),
    ]
    specs += [
        spec("accumulate_brier_empty", "accumulate_brier", scores=[]),
        spec("accumulate_brier_single", "accumulate_brier", scores=[0.12]),
        spec("accumulate_brier_three", "accumulate_brier", scores=[0.1, 0.2, 0.3]),
        spec("accumulate_brier_perfect", "accumulate_brier", scores=[0.0, 0.0, 0.0]),
        spec("accumulate_brier_random", "accumulate_brier", scores=[0.25, 0.25, 0.25, 0.25]),
        spec("accumulate_brier_mixed", "accumulate_brier", scores=[0.01, 0.36, 0.09, 0.16, 0.49]),
    ]

    return specs


def weights_from_start_offsets(start_offsets: list[int]) -> list[float]:
    pairs = sorted(enumerate(start_offsets), key=lambda pair: pair[1])
    weights = [0.0 for _ in start_offsets]
    for rank, (idx, _) in enumerate(pairs):
        weights[idx] = float(rank + 1)
    return weights


def dispatch(item: dict[str, Any]) -> dict[str, Any]:
    op = item["op"]
    if op == "baseline_binary":
        return {"score": ref.baseline_binary(item["probability_yes"], item["resolved_yes"])}
    if op == "brier_binary":
        return {"score": ref.brier_binary(item["probability_yes"], item["resolved_yes"])}
    if op == "peer_binary":
        return {"score": ref.peer_binary(item["probability_yes_self"], item["gmp_outcome"], item["n"], item["resolved_yes"])}
    if op == "geometric_mean":
        return {"value": ref.geometric_mean(item["values"])}
    if op == "cdf_prob_at":
        return {"probability": ref.cdf_prob_at(item["cdf"], item["k_resolution"])}
    if op == "brier_numeric":
        return {"score": ref.brier_numeric(item["cdf"], item["k_resolution"])}
    if op == "baseline_for_numeric_question":
        return {"baseline": ref.baseline_for_numeric_question(item["open_lower"], item["open_upper"])}
    if op == "baseline_numeric":
        return {"score": ref.baseline_numeric(item["cdf"], item["k_resolution"], item["baseline"])}
    if op == "peer_numeric":
        return {"score": ref.peer_numeric(item["p_self"], item["gmp"], item["n"])}
    if op == "recency_weights":
        return {"weights": weights_from_start_offsets(item["start_offsets"])}
    if op == "weighted_median_binary":
        return {"median": ref.weighted_median_binary(item["probabilities"], item["weights"])}
    if op == "weighted_avg_cdf":
        return {"cdf": ref.weighted_avg_cdf(item["cdfs"], item["weights"])}
    if op == "accumulate_brier":
        return {"score": ref.accumulate_brier(item["scores"])}
    raise ValueError(f"unknown op {op}")


def write_json(path: Path, payload: dict[str, Any]) -> None:
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n")


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    input_dir = root / "fixtures" / "input"
    expected_dir = root / "fixtures" / "expected"
    input_dir.mkdir(parents=True, exist_ok=True)
    expected_dir.mkdir(parents=True, exist_ok=True)

    for item in build_specs():
        name = item["name"]
        input_payload = {k: v for k, v in item.items() if k != "name"}
        write_json(input_dir / f"{name}.json", input_payload)
        write_json(expected_dir / f"{name}.json", dispatch(item))

    print(f"wrote {len(build_specs())} fixture pairs")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
