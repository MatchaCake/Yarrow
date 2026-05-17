package main

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"yarrow/demo/algo"
)

type Agent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Persona string `json:"persona"`
}

type Stance struct {
	Probability float64 `json:"probability"`
	Reasoning   string  `json:"reasoning"`
}

type Market struct {
	ID                  string            `json:"id"`
	Title               string            `json:"title"`
	Slug                string            `json:"slug"`
	PolymarketYes       float64           `json:"polymarket_yes"`
	Resolution          string            `json:"resolution"`
	Stances             map[string]Stance `json:"stances"`
	RiskFactors         []string          `json:"risk_factors"`
	UserForecasts       []Forecast        `json:"user_forecasts"`
	CommunityPrediction float64           `json:"community_prediction"`
	Divergence          float64           `json:"divergence"`
}

type Forecast struct {
	Agent          string    `json:"agent"`
	ProbabilityYes float64   `json:"probability_yes"`
	SubmittedAt    time.Time `json:"submitted_at"`
}

type Report struct {
	Market             Market      `json:"market"`
	AgentBreakdown     []AgentLine `json:"agent_breakdown"`
	ConfidenceInterval [2]float64  `json:"confidence_interval"`
}

type AgentLine struct {
	AgentID     string  `json:"agent_id"`
	AgentName   string  `json:"agent_name"`
	Probability float64 `json:"probability"`
	Reasoning   string  `json:"reasoning"`
	Brier       float64 `json:"brier"`
}

type LeaderEntry struct {
	AgentID       string  `json:"agent_id"`
	AgentName     string  `json:"agent_name"`
	ResolvedCount int     `json:"resolved_count"`
	BrierMean     float64 `json:"brier_mean"`
	BaselineTotal float64 `json:"baseline_total"`
}

type AgentHistory struct {
	Agent       Agent             `json:"agent"`
	Predictions []AgentPrediction `json:"predictions"`
}

type AgentPrediction struct {
	MarketID    string  `json:"market_id"`
	MarketTitle string  `json:"market_title"`
	Probability float64 `json:"probability"`
	Reasoning   string  `json:"reasoning"`
	Resolution  string  `json:"resolution"`
	Resolved    bool    `json:"resolved"`
	Result      string  `json:"result"`
	Brier       float64 `json:"brier"`
}

type ForecastStore struct {
	mu      sync.RWMutex
	agents  []Agent
	markets map[string]Market
}

var errMarketNotFound = errors.New("market not found")

func LoadStore(marketsPath, agentsPath string) (*ForecastStore, error) {
	agentData, err := os.ReadFile(agentsPath)
	if err != nil {
		return nil, err
	}
	var agents []Agent
	if err := json.Unmarshal(agentData, &agents); err != nil {
		return nil, err
	}

	marketData, err := os.ReadFile(marketsPath)
	if err != nil {
		return nil, err
	}
	var markets []Market
	if err := json.Unmarshal(marketData, &markets); err != nil {
		return nil, err
	}

	store := &ForecastStore{
		agents:  append([]Agent(nil), agents...),
		markets: make(map[string]Market, len(markets)),
	}
	for _, market := range markets {
		market.UserForecasts = nil
		store.markets[market.ID] = market
	}
	return store, nil
}

func (s *ForecastStore) Agents() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Agent(nil), s.agents...)
}

func (s *ForecastStore) Markets() []Market {
	s.mu.RLock()
	defer s.mu.RUnlock()
	markets := make([]Market, 0, len(s.markets))
	for _, market := range s.markets {
		markets = append(markets, s.enrichedMarketLocked(market))
	}
	sort.SliceStable(markets, func(i, j int) bool {
		return abs(markets[i].Divergence) > abs(markets[j].Divergence)
	})
	return markets
}

func (s *ForecastStore) Market(id string) (Market, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	market, ok := s.markets[id]
	if !ok {
		return Market{}, false
	}
	return s.enrichedMarketLocked(market), true
}

func (s *ForecastStore) AddForecast(id string, forecast Forecast) (Market, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	market, ok := s.markets[id]
	if !ok {
		return Market{}, errMarketNotFound
	}
	if forecast.SubmittedAt.IsZero() {
		forecast.SubmittedAt = time.Now().UTC()
	}
	market.UserForecasts = append(market.UserForecasts, forecast)
	s.markets[id] = market
	return s.enrichedMarketLocked(market), nil
}

func (s *ForecastStore) Report(id string) (Report, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	market, ok := s.markets[id]
	if !ok {
		return Report{}, false
	}
	enriched := s.enrichedMarketLocked(market)
	lines := make([]AgentLine, 0, len(s.agents))
	minProbability := math.Inf(1)
	maxProbability := math.Inf(-1)
	resolved, resolvedYes := resolutionOutcome(enriched.Resolution)
	for _, agent := range s.agents {
		stance, ok := enriched.Stances[agent.ID]
		if !ok {
			continue
		}
		if stance.Probability < minProbability {
			minProbability = stance.Probability
		}
		if stance.Probability > maxProbability {
			maxProbability = stance.Probability
		}
		line := AgentLine{
			AgentID:     agent.ID,
			AgentName:   agent.Name,
			Probability: stance.Probability,
			Reasoning:   stance.Reasoning,
		}
		if resolved {
			line.Brier = algo.BrierBinary(stance.Probability, resolvedYes)
		}
		lines = append(lines, line)
	}
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].Probability > lines[j].Probability
	})
	if len(lines) == 0 {
		minProbability = 0
		maxProbability = 0
	}
	return Report{
		Market:             enriched,
		AgentBreakdown:     lines,
		ConfidenceInterval: [2]float64{minProbability, maxProbability},
	}, true
}

func (s *ForecastStore) Leaderboard() []LeaderEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type totals struct {
		name          string
		brierScores   []float64
		baselineTotal float64
	}
	byAgent := make(map[string]*totals, len(s.agents))
	for _, agent := range s.agents {
		byAgent[agent.ID] = &totals{name: agent.Name}
	}

	for _, market := range s.markets {
		resolved, resolvedYes := resolutionOutcome(market.Resolution)
		if !resolved {
			continue
		}
		for _, agent := range s.agents {
			stance, ok := market.Stances[agent.ID]
			if !ok {
				continue
			}
			entry := byAgent[agent.ID]
			entry.brierScores = append(entry.brierScores, algo.BrierBinary(stance.Probability, resolvedYes))
			entry.baselineTotal += algo.BaselineBinary(stance.Probability, resolvedYes)
		}
	}

	leaders := make([]LeaderEntry, 0, len(s.agents))
	for _, agent := range s.agents {
		entry := byAgent[agent.ID]
		leaders = append(leaders, LeaderEntry{
			AgentID:       agent.ID,
			AgentName:     entry.name,
			ResolvedCount: len(entry.brierScores),
			BrierMean:     algo.AccumulateBrier(entry.brierScores),
			BaselineTotal: entry.baselineTotal,
		})
	}
	sort.SliceStable(leaders, func(i, j int) bool {
		return leaders[i].BrierMean < leaders[j].BrierMean
	})
	return leaders
}

func (s *ForecastStore) AgentHistory(id string) (AgentHistory, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var selected Agent
	found := false
	for _, agent := range s.agents {
		if agent.ID == id {
			selected = agent
			found = true
			break
		}
	}
	if !found {
		return AgentHistory{}, false
	}

	markets := make([]Market, 0, len(s.markets))
	for _, market := range s.markets {
		markets = append(markets, market)
	}
	sort.SliceStable(markets, func(i, j int) bool {
		return markets[i].ID < markets[j].ID
	})

	predictions := make([]AgentPrediction, 0, len(markets))
	for _, market := range markets {
		stance, ok := market.Stances[id]
		if !ok {
			continue
		}
		resolved, resolvedYes := resolutionOutcome(market.Resolution)
		result := "open"
		brier := 0.0
		if resolved {
			result = market.Resolution
			brier = algo.BrierBinary(stance.Probability, resolvedYes)
		}
		predictions = append(predictions, AgentPrediction{
			MarketID:    market.ID,
			MarketTitle: market.Title,
			Probability: stance.Probability,
			Reasoning:   stance.Reasoning,
			Resolution:  market.Resolution,
			Resolved:    resolved,
			Result:      result,
			Brier:       brier,
		})
	}

	return AgentHistory{
		Agent:       selected,
		Predictions: predictions,
	}, true
}

func (s *ForecastStore) enrichedMarketLocked(market Market) Market {
	market.Stances = copyStances(market.Stances)
	market.RiskFactors = append([]string(nil), market.RiskFactors...)
	market.UserForecasts = append([]Forecast(nil), market.UserForecasts...)

	forecasts := make([]algo.ActiveForecast, 0, len(s.agents)+len(market.UserForecasts))
	base := time.Unix(1700000000, 0).UTC()
	for idx, agent := range s.agents {
		stance, ok := market.Stances[agent.ID]
		if !ok {
			continue
		}
		forecasts = append(forecasts, algo.ActiveForecast{
			UserID:         agent.ID,
			StartTime:      base.Add(time.Duration(idx) * time.Second),
			ProbabilityYes: stance.Probability,
		})
	}
	for _, forecast := range market.UserForecasts {
		forecasts = append(forecasts, algo.ActiveForecast{
			UserID:         forecast.Agent,
			StartTime:      forecast.SubmittedAt,
			ProbabilityYes: forecast.ProbabilityYes,
		})
	}
	cp, _ := algo.WeightedMedianBinary(forecasts)
	market.CommunityPrediction = cp
	market.Divergence = cp - market.PolymarketYes
	return market
}

func copyStances(in map[string]Stance) map[string]Stance {
	out := make(map[string]Stance, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func resolutionOutcome(resolution string) (resolved bool, resolvedYes bool) {
	switch resolution {
	case "yes":
		return true, true
	case "no":
		return true, false
	default:
		return false, false
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
