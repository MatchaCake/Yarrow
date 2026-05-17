# Yarrow Demo Video Voiceover

## 75-second script

Prediction markets are becoming real information infrastructure. In 2025, sector trading volume crossed 50 billion dollars, with Polymarket alone around 21.5 billion dollars.

But the market still mostly ranks capital. If a bettor has size, timing, or one lucky trade, they can look smart. What we actually need for AI forecasters is a reputation layer: who is calibrated, who disagrees with price, and whose reasoning is worth inspecting.

This is Yarrow. The demo starts with Polymarket-style market snapshots, then asks five specialized AI agents to forecast each event: a momentum quant, a contrarian, a macro analyst, a technical fundamentals agent, and a political analyst.

Yarrow aggregates their forecasts, compares the consensus against market-implied probability, and sorts the dashboard by the biggest divergence. Each market opens into an agent debate: every agent shows its probability, whether it is above or below the market, and the reasoning behind the stance.

The scoring layer is not hand-wavy. Yarrow uses Metaculus-style Brier scoring, with Go math diff-tested against an independent Python reference at one part in one million.

For the live workflow, a user agent can submit a new forecast through the Yarrow API. The web page records that prediction on the market card and in the report, so the demo shows the full loop: API forecast in, updated consensus out, visible audit trail on screen.

The point is simple: prediction markets price events, but Yarrow evaluates forecasters. That turns AI agents from opaque opinions into measurable, debatable, and rankable forecasting systems.

## Shot List

1. Landing page at `/`: show the market-size stats and the core problem statement.
2. Click into `/demo`: show markets sorted by absolute Yarrow-versus-Polymarket divergence.
3. Open a report: highlight the agent debate cards and the above-market / below-market tags.
4. Submit a forecast through the form or API: show the new submitted forecast appearing on the market card.
5. Open the same report again: show the submitted forecast recorded alongside the agent debate.
6. Close on the leaderboard: explain that Yarrow ranks forecasters by calibration, not wallet size.

## Source Notes

- Market volume stats: linked from the landing page sources.
- Scoring context: Metaculus FAQ and Metaculus `score_math.py`, linked from the landing page and `REFERENCES.md`.
- Math validation: local Go tests compare against the independent Python fixtures under `algoref/` and `fixtures/`.
