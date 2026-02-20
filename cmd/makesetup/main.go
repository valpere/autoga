package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/val/autoga/internal/makecom"
)

type config struct {
	apiToken       string
	zone           string
	teamID         int
	folderID       int
	intervalSec    int
	scenarioName   string
	rssFeedURL     string
	scraperURL     string
	scraperAPIKey  string
	llmAPIURL      string
	llmAPIKey      string
	llmModel       string
	telegramChatID string
	activate       bool
}

func main() {
	cfg := loadConfig()

	client := makecom.NewClient(cfg.zone, cfg.apiToken)

	bp := makecom.BuildBlueprint(makecom.ScenarioConfig{
		Name:           cfg.scenarioName,
		RSSFeedURL:     cfg.rssFeedURL,
		ScraperURL:     cfg.scraperURL,
		ScraperAPIKey:  cfg.scraperAPIKey,
		LLMAPIUrl:      cfg.llmAPIURL,
		LLMAPIKey:      cfg.llmAPIKey,
		LLMModel:       cfg.llmModel,
		TelegramChatID: cfg.telegramChatID,
	})

	sched := makecom.Scheduling{
		Type:     "indefinitely",
		Interval: cfg.intervalSec,
	}

	log.Printf("creating scenario %q on team %d (zone: %s)...", cfg.scenarioName, cfg.teamID, cfg.zone)

	scenario, err := client.CreateScenario(cfg.teamID, bp, sched, cfg.folderID)
	if err != nil {
		log.Fatalf("create scenario: %v", err)
	}

	fmt.Printf("scenario created: id=%d name=%q\n", scenario.ID, scenario.Name)
	fmt.Printf("edit: https://%s.make.com/scenario/%d/edit\n", cfg.zone, scenario.ID)

	if cfg.activate {
		if err := client.ActivateScenario(scenario.ID); err != nil {
			log.Fatalf("activate scenario: %v", err)
		}
		fmt.Println("scenario activated")
	} else {
		fmt.Println("tip: set ACTIVATE=true to activate the scenario automatically")
		fmt.Println("note: set Telegram connection in Make.com UI before activating")
	}
}

func loadConfig() config {
	return config{
		apiToken:       requireEnv("MAKE_API_TOKEN"),
		zone:           getEnv("MAKE_ZONE", "eu1"),
		teamID:         requireInt("MAKE_TEAM_ID"),
		folderID:       getInt("MAKE_FOLDER_ID", 0),
		intervalSec:    getInt("MAKE_INTERVAL_SEC", 900),
		scenarioName:   getEnv("SCENARIO_NAME", "AutoGA Digest"),
		rssFeedURL:     requireEnv("RSS_FEED_URL"),
		scraperURL:     requireEnv("AUTOGA_URL"),
		scraperAPIKey:  getEnv("AUTOGA_API_KEY", ""),
		llmAPIURL:      requireEnv("LLM_API_URL"),
		llmAPIKey:      requireEnv("LLM_API_KEY"),
		llmModel:       requireEnv("LLM_MODEL"),
		telegramChatID: requireEnv("TELEGRAM_CHAT_ID"),
		activate:       getEnv("ACTIVATE", "false") == "true",
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func requireInt(key string) int {
	v := requireEnv(key)
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("env var %s must be an integer, got %q", key, v)
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
