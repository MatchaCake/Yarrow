package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type forecastRequest struct {
	MarketID       string   `json:"market_id"`
	Agent          string   `json:"agent"`
	ProbabilityYes *float64 `json:"probability_yes"`
	Reasoning      string   `json:"reasoning"`
}

func NewRouter(store *ForecastStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, store.Agents())
	})
	mux.HandleFunc("/api/agents/", agentRoute(store))
	mux.HandleFunc("/api/markets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, store.Markets())
	})
	mux.HandleFunc("/api/forecasts", forecastOperationsRoute(store))
	mux.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, store.Leaderboard())
	})
	mux.HandleFunc("/api/markets/", marketRoute(store))
	return mux
}

func agentRoute(store *ForecastStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/agents/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] != "history" {
			http.NotFound(w, r)
			return
		}
		history, ok := store.AgentHistory(parts[0])
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, history)
	}
}

func forecastOperationsRoute(store *ForecastStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			records, err := store.Forecasts(
				strings.TrimSpace(r.URL.Query().Get("market_id")),
				strings.TrimSpace(r.URL.Query().Get("agent")),
			)
			if errors.Is(err, errMarketNotFound) {
				http.NotFound(w, r)
				return
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, records)
		case http.MethodPost:
			operation, ok := createForecast(w, r, store, "")
			if ok {
				writeJSON(w, operation)
			}
		default:
			methodNotAllowed(w)
		}
	}
}

func marketRoute(store *ForecastStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/markets/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		id := parts[0]
		if len(parts) == 1 {
			if r.Method != http.MethodGet {
				methodNotAllowed(w)
				return
			}
			market, ok := store.Market(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, market)
			return
		}
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		switch parts[1] {
		case "report":
			if r.Method != http.MethodGet {
				methodNotAllowed(w)
				return
			}
			report, ok := store.Report(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, report)
		case "forecasts":
			if r.Method == http.MethodGet {
				records, err := store.Forecasts(id, "")
				if errors.Is(err, errMarketNotFound) {
					http.NotFound(w, r)
					return
				}
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, records)
				return
			}
			if r.Method != http.MethodPost {
				methodNotAllowed(w)
				return
			}
			handlePostForecast(w, r, store, id)
		default:
			http.NotFound(w, r)
		}
	}
}

func handlePostForecast(w http.ResponseWriter, r *http.Request, store *ForecastStore, id string) {
	operation, ok := createForecast(w, r, store, id)
	if ok {
		writeJSON(w, operation.Market)
	}
}

func createForecast(w http.ResponseWriter, r *http.Request, store *ForecastStore, routeMarketID string) (ForecastOperation, bool) {
	defer r.Body.Close()
	var payload forecastRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return ForecastOperation{}, false
	}
	marketID := strings.TrimSpace(routeMarketID)
	if marketID == "" {
		marketID = strings.TrimSpace(payload.MarketID)
	}
	if marketID == "" {
		http.Error(w, "market_id is required", http.StatusBadRequest)
		return ForecastOperation{}, false
	}
	payload.Agent = strings.TrimSpace(payload.Agent)
	if payload.Agent == "" {
		http.Error(w, "agent is required", http.StatusBadRequest)
		return ForecastOperation{}, false
	}
	if payload.ProbabilityYes == nil {
		http.Error(w, "probability_yes is required", http.StatusBadRequest)
		return ForecastOperation{}, false
	}
	if *payload.ProbabilityYes < 0 || *payload.ProbabilityYes > 1 {
		http.Error(w, "probability_yes must be between 0 and 1", http.StatusBadRequest)
		return ForecastOperation{}, false
	}
	operation, err := store.SubmitForecast(marketID, Forecast{
		Agent:          payload.Agent,
		ProbabilityYes: *payload.ProbabilityYes,
		Reasoning:      strings.TrimSpace(payload.Reasoning),
		SubmittedAt:    time.Now().UTC(),
	})
	if errors.Is(err, errMarketNotFound) {
		http.NotFound(w, r)
		return ForecastOperation{}, false
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ForecastOperation{}, false
	}
	return operation, true
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
