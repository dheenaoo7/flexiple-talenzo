package main

import (
	"context"
	"log"
	"net/http"

	"talent-search-rubric-arm/backend/internal/api"
	"talent-search-rubric-arm/backend/internal/config"
	"talent-search-rubric-arm/backend/internal/data"
	"talent-search-rubric-arm/backend/internal/llm"
	"talent-search-rubric-arm/backend/internal/session"
	"talent-search-rubric-arm/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	profiles, err := data.LoadProfiles(cfg.ProfilesPath)
	if err != nil {
		log.Fatalf("loading profiles: %v", err)
	}
	log.Printf("loaded %d profiles from %s", len(profiles.Profiles), cfg.ProfilesPath)

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}
	defer db.Close()

	llmClient, err := llm.New(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatalf("creating gemini client: %v", err)
	}

	svc := session.New(db, llmClient, profiles)
	a := api.New(svc)

	addr := ":" + cfg.Port
	log.Printf("listening on %s (model=%s)", addr, cfg.GeminiModel)
	if err := http.ListenAndServe(addr, a.Router()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
