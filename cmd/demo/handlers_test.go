package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func getText(t *testing.T, handler http.Handler, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d body=%s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestRootServesDedicatedLandingAndDemoServesDashboard(t *testing.T) {
	handler := testServer(t)

	landing := getText(t, handler, "/")
	if !strings.Contains(landing, `href="/demo"`) {
		t.Fatalf("landing page should link to /demo")
	}
	if strings.Contains(landing, `id="markets"`) {
		t.Fatalf("landing page should not include the prediction dashboard")
	}

	demo := getText(t, handler, "/demo")
	if !strings.Contains(demo, `id="markets"`) {
		t.Fatalf("demo page should include the prediction dashboard")
	}
	if strings.Contains(demo, `class="intro"`) {
		t.Fatalf("demo page should not include the landing intro banner")
	}
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

func TestAgentHistoryEndpointReturnsPredictionsAndResults(t *testing.T) {
	handler := testServer(t)
	history := getJSON[struct {
		Agent       Agent `json:"agent"`
		Predictions []struct {
			MarketID    string  `json:"market_id"`
			MarketTitle string  `json:"market_title"`
			Probability float64 `json:"probability"`
			Reasoning   string  `json:"reasoning"`
			Resolution  string  `json:"resolution"`
			Resolved    bool    `json:"resolved"`
			Result      string  `json:"result"`
			Brier       float64 `json:"brier"`
		} `json:"predictions"`
	}](t, handler, "/api/agents/momentum_quant/history")

	if history.Agent.ID != "momentum_quant" {
		t.Fatalf("agent id = %q, want momentum_quant", history.Agent.ID)
	}
	if len(history.Predictions) != 10 {
		t.Fatalf("predictions len = %d, want 10", len(history.Predictions))
	}
	for i := 1; i < len(history.Predictions); i++ {
		if history.Predictions[i].MarketID < history.Predictions[i-1].MarketID {
			t.Fatalf("predictions not sorted by market id at %d", i)
		}
	}

	var openFound, resolvedFound bool
	for _, prediction := range history.Predictions {
		switch prediction.MarketID {
		case "mkt_001":
			openFound = true
			if prediction.Resolved || prediction.Result != "open" || prediction.Brier != 0 {
				t.Fatalf("open prediction = %+v, want unresolved open with zero brier", prediction)
			}
			if prediction.Probability != 0.55 || prediction.Reasoning == "" || prediction.MarketTitle == "" {
				t.Fatalf("bad open prediction details: %+v", prediction)
			}
		case "mkt_003":
			resolvedFound = true
			if !prediction.Resolved || prediction.Result != "yes" {
				t.Fatalf("resolved prediction = %+v, want yes result", prediction)
			}
			if prediction.Probability != 0.72 {
				t.Fatalf("probability = %v, want 0.72", prediction.Probability)
			}
			if diff := abs(prediction.Brier - 0.0784); diff > 1e-9 {
				t.Fatalf("brier = %v, want 0.0784", prediction.Brier)
			}
		}
	}
	if !openFound || !resolvedFound {
		t.Fatalf("missing open=%v resolved=%v predictions", openFound, resolvedFound)
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

func TestForecastOperationsAPICreatesAndListsForecasts(t *testing.T) {
	handler := testServer(t)
	before := getJSON[Market](t, handler, "/api/markets/mkt_001")

	body := bytes.NewBufferString(`{"market_id":"mkt_001","agent":"api_agent","probability_yes":0.82,"reasoning":"API smoke forecast"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/forecasts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		Forecast ForecastRecord `json:"forecast"`
		Market   Market         `json:"market"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Forecast.MarketID != "mkt_001" || created.Forecast.MarketTitle == "" {
		t.Fatalf("forecast market fields = %+v", created.Forecast)
	}
	if created.Forecast.Agent != "api_agent" || created.Forecast.ProbabilityYes != 0.82 || created.Forecast.Reasoning != "API smoke forecast" {
		t.Fatalf("forecast fields = %+v", created.Forecast)
	}
	if created.Forecast.SubmittedAt.IsZero() {
		t.Fatalf("submitted_at was not set: %+v", created.Forecast)
	}
	if created.Market.ID != "mkt_001" || len(created.Market.UserForecasts) != 1 {
		t.Fatalf("updated market = %+v", created.Market)
	}
	if created.Market.CommunityPrediction <= before.CommunityPrediction {
		t.Fatalf("API forecast should be included in Yarrow prediction: before=%v after=%v", before.CommunityPrediction, created.Market.CommunityPrediction)
	}

	records := getJSON[[]ForecastRecord](t, handler, "/api/forecasts?market_id=mkt_001&agent=api_agent")
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if records[0].MarketID != "mkt_001" || records[0].Agent != "api_agent" || records[0].ProbabilityYes != 0.82 || records[0].Reasoning != "API smoke forecast" {
		t.Fatalf("record = %+v", records[0])
	}

	marketRecords := getJSON[[]ForecastRecord](t, handler, "/api/markets/mkt_001/forecasts")
	if len(marketRecords) != 1 {
		t.Fatalf("market records len = %d, want 1", len(marketRecords))
	}
}

func TestForecastOperationsAPIRequiresMarketID(t *testing.T) {
	handler := testServer(t)

	body := bytes.NewBufferString(`{"agent":"api_agent","probability_yes":0.82}`)
	req := httptest.NewRequest(http.MethodPost, "/api/forecasts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST status = %d body=%s, want 400", rec.Code, rec.Body.String())
	}
}

func TestForecastOperationsAPIRequiresProbability(t *testing.T) {
	handler := testServer(t)

	body := bytes.NewBufferString(`{"market_id":"mkt_001","agent":"api_agent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/forecasts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST status = %d body=%s, want 400", rec.Code, rec.Body.String())
	}
}
