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

Open http://localhost:8080 for the dedicated landing page, then use the CTA or go directly to http://localhost:8080/demo for the prediction dashboard.

The dashboard loads all 10 markets sorted by absolute divergence, shows the 5-agent leaderboard with two reference rows, opens a per-market agent-debate modal, and lets audience members submit an in-memory stance on open markets. Submitted forecasts are included in the Yarrow consensus, displayed on the market card, and shown in the report modal so API/user-agent predictions are visible in the web demo.

The video narration handoff lives in `docs/demo-video-voiceover.md`.

Opening `cmd/demo/index.html` directly, or hosting it as static HTML without the Go API, runs a local browser-only prediction dashboard using the baked-in snapshot data. Use the Go server for the dedicated landing page and API-backed in-memory state; append `?api=1` if you run the server on a non-default port and want to force API mode.

## API

| Method | Path |
| --- | --- |
| `GET` | `/` |
| `GET` | `/api/agents` |
| `GET` | `/api/agents/{id}/history` |
| `GET` | `/api/forecasts` |
| `POST` | `/api/forecasts` |
| `GET` | `/api/markets` |
| `GET` | `/api/markets/{id}` |
| `GET` | `/api/markets/{id}/report` |
| `GET` | `/api/markets/{id}/forecasts` |
| `POST` | `/api/markets/{id}/forecasts` |
| `GET` | `/api/leaderboard` |

Use `/api/forecasts` for forecast operations. `GET /api/forecasts` lists submitted audience/API forecasts, with optional `market_id` and `agent` query filters. `POST /api/forecasts` creates a forecast and returns both the created forecast record and the updated market. `reasoning` is optional. The older per-market submit route is kept for compatibility with existing callers.

Example forecast submission:

```sh
curl -s -X POST http://localhost:8080/api/forecasts \
  -H 'Content-Type: application/json' \
  -d '{"market_id":"mkt_001","agent":"stage_demo","probability_yes":0.95,"reasoning":"API forecast from Yarrow demo"}'
```

List submitted forecasts:

```sh
curl -s 'http://localhost:8080/api/forecasts?market_id=mkt_001&agent=stage_demo'
```
