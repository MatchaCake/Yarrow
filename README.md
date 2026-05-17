# Yarrow Demo

Standalone hackathon demo for Yarrow: Metaculus-style forecasting math over baked-in Polymarket market snapshots, plus a single-binary web demo showing where agent consensus diverges from market-implied probability.

This repo is intentionally not production Yarrow. It has no database, auth, importer, wallet flow, live Polymarket fetch, or JavaScript build step. Runtime state is in memory and resets when the server restarts.

## What Is Included

- `algo/`: pure Go scoring and aggregation kernel with pinned signatures from the Ralph Loop spec.
- `algoref/`: independent Python reference implementation plus fixture generator.
- `fixtures/`: 78 Python-generated input/expected fixture pairs consumed by Go diff tests at `1e-6` absolute tolerance.
- `data/`: 10 curated market snapshots and 5 agent personas copied from `~/Documents/git/Yarrow/hackathon`.
- `cmd/demo/`: Go HTTP server on `:8080`, in-memory forecast store, pinned JSON API, and one vanilla `index.html`.

## Math Function Inventory

| Area | Functions |
| --- | --- |
| Binary scoring | `BaselineBinary`, `BrierBinary`, `PeerBinary`, `GeometricMean` |
| Numeric scoring | `CDFProbAt`, `BrierNumeric`, `BaselineForNumericQuestion`, `BaselineNumeric`, `PeerNumeric` |
| Aggregation | `RecencyWeights`, `WeightedMedianBinary`, `WeightedAvgCDF` |
| Track record | `AccumulateBrier` |

References for the scoring formulas are listed in `REFERENCES.md`.

## Run Tests

```sh
algoref/.venv/bin/python -m pytest algoref/ -q
go test ./algo/... -count=1 -race
go test ./... -count=1
```

Regenerate fixtures:

```sh
algoref/.venv/bin/python algoref/generate.py
```

## Run The Demo

```sh
go run ./cmd/demo
```

Open http://localhost:8080.

The page loads all 10 markets sorted by absolute divergence, shows the 5-agent leaderboard with two reference rows, opens a per-market analysis modal, and lets audience members submit an in-memory stance on open markets.

Do not open `cmd/demo/index.html` directly from Finder or with a `file://` URL. The buttons call same-origin `/api/...` routes, so the Go server must be running.

## API

| Method | Path |
| --- | --- |
| `GET` | `/` |
| `GET` | `/api/agents` |
| `GET` | `/api/markets` |
| `GET` | `/api/markets/{id}` |
| `GET` | `/api/markets/{id}/report` |
| `POST` | `/api/markets/{id}/forecasts` |
| `GET` | `/api/leaderboard` |

Example forecast submission:

```sh
curl -s -X POST http://localhost:8080/api/markets/mkt_001/forecasts \
  -H 'Content-Type: application/json' \
  -d '{"agent":"stage_demo","probability_yes":0.95}'
```
