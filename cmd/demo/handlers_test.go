package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T) http.Handler {
	t.Helper()
	store, err := LoadStore("../../data/markets.json", "../../data/agents.json")
	if err != nil {
		t.Fatal(err)
	}
	return NewRouter(store)
}

func getJSON[T any](t *testing.T, handler http.Handler, path string) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d body=%s", path, rec.Code, rec.Body.String())
	}
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return out
}

func TestMarketsEndpointReturnsTenSortedByAbsDivergence(t *testing.T) {
	handler := testServer(t)
	markets := getJSON[[]Market](t, handler, "/api/markets")
	if len(markets) != 10 {
		t.Fatalf("len = %d, want 10", len(markets))
	}
	for i := 1; i < len(markets); i++ {
		if abs(markets[i].Divergence) > abs(markets[i-1].Divergence) {
			t.Fatalf("markets not sorted by abs divergence at %d", i)
		}
	}
}

func TestLeaderboardReturnsFiveAgentsWithThreeResolvedMarkets(t *testing.T) {
	handler := testServer(t)
	leaders := getJSON[[]LeaderEntry](t, handler, "/api/leaderboard")
	if len(leaders) != 5 {
		t.Fatalf("len = %d, want 5", len(leaders))
	}
	for _, entry := range leaders {
		if entry.ResolvedCount != 3 {
			t.Fatalf("%s resolved_count = %d, want 3", entry.AgentID, entry.ResolvedCount)
		}
	}
	for i := 1; i < len(leaders); i++ {
		if leaders[i].BrierMean < leaders[i-1].BrierMean {
			t.Fatalf("leaderboard not sorted ascending by brier_mean")
		}
	}
}

func TestReportEndpointReturnsSortedAgentBreakdownAndCI(t *testing.T) {
	handler := testServer(t)
	report := getJSON[Report](t, handler, "/api/markets/mkt_001/report")
	if len(report.AgentBreakdown) != 5 {
		t.Fatalf("agent breakdown len = %d, want 5", len(report.AgentBreakdown))
	}
	for i := 1; i < len(report.AgentBreakdown); i++ {
		if report.AgentBreakdown[i].Probability > report.AgentBreakdown[i-1].Probability {
			t.Fatalf("agent breakdown not sorted descending")
		}
	}
	if report.ConfidenceInterval[0] >= report.ConfidenceInterval[1] {
		t.Fatalf("bad confidence interval: %v", report.ConfidenceInterval)
	}
}

func TestPostForecastUpdatesCommunityPrediction(t *testing.T) {
	handler := testServer(t)
	before := getJSON[Market](t, handler, "/api/markets/mkt_001")

	body := bytes.NewBufferString(`{"agent":"stage_demo","probability_yes":0.95}`)
	req := httptest.NewRequest(http.MethodPost, "/api/markets/mkt_001/forecasts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST status = %d body=%s", rec.Code, rec.Body.String())
	}
	var after Market
	if err := json.Unmarshal(rec.Body.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	if after.CommunityPrediction < before.CommunityPrediction {
		t.Fatalf("community prediction should rise or hold after high forecast: before=%v after=%v", before.CommunityPrediction, after.CommunityPrediction)
	}
	if len(after.UserForecasts) != 1 {
		t.Fatalf("user forecasts len = %d, want 1", len(after.UserForecasts))
	}
}
