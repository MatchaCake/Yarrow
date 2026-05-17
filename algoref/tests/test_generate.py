from algoref import generate


def test_build_specs_covers_required_fixture_volume_and_ops():
    specs = generate.build_specs()
    ops = {spec["op"] for spec in specs}

    assert len(specs) >= 60
    assert {
        "baseline_binary",
        "brier_binary",
        "peer_binary",
        "geometric_mean",
        "cdf_prob_at",
        "brier_numeric",
        "baseline_for_numeric_question",
        "baseline_numeric",
        "peer_numeric",
        "recency_weights",
        "weighted_median_binary",
        "weighted_avg_cdf",
        "accumulate_brier",
    }.issubset(ops)


def test_dispatch_returns_expected_shape_for_vector_and_scalar_ops():
    by_name = {spec["name"]: spec for spec in generate.build_specs()}

    assert set(generate.dispatch(by_name["baseline_binary_mid_yes"])) == {"score"}
    assert set(generate.dispatch(by_name["weighted_avg_cdf_empty"])) == {"cdf"}
    assert set(generate.dispatch(by_name["recency_weights_mixed_order"])) == {"weights"}
