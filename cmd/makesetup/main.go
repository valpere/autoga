package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/val/autoga/internal/makecom"
)

type feed struct {
	name string
	url  string
}

type config struct {
	apiToken      string
	zone          string
	teamID        int
	folderID      int
	intervalSec   int
	scenarioName  string
	rssFeedsFile  string
	scraperURL    string
	scraperAPIKey string
	llmAPIURL     string
	llmAPIKey     string
	llmModel      string
	telegramChatID       string
	telegramConnectionID int
}

func main() {
	cfg := loadConfig()

	feeds, err := readFeeds(cfg.rssFeedsFile)
	if err != nil {
		log.Fatalf("read feeds: %v", err)
	}
	if len(feeds) == 0 {
		log.Fatalf("%s is empty", cfg.rssFeedsFile)
	}

	client := makecom.NewClient(cfg.zone, cfg.apiToken)
	sched := makecom.Scheduling{Type: "indefinitely", Interval: cfg.intervalSec}

	for _, f := range feeds {
		name := cfg.scenarioName + ": " + f.name

		bp := makecom.BuildBlueprint(makecom.ScenarioConfig{
			Name:                 name,
			Zone:                 cfg.zone,
			RSSFeedURL:           f.url,
			ScraperURL:           cfg.scraperURL,
			ScraperAPIKey:        cfg.scraperAPIKey,
			LLMAPIUrl:            cfg.llmAPIURL,
			LLMAPIKey:            cfg.llmAPIKey,
			LLMModel:             cfg.llmModel,
			TelegramChatID:       cfg.telegramChatID,
			TelegramConnectionID: cfg.telegramConnectionID,
		})

		log.Printf("creating scenario %q...", name)

		scenario, err := client.CreateScenario(cfg.teamID, bp, sched, cfg.folderID)
		if err != nil {
			log.Printf("ERROR %q: %v", name, err)
			continue
		}

		fmt.Printf("created: id=%d name=%q\n", scenario.ID, scenario.Name)
		fmt.Printf("  edit: https://%s.make.com/scenario/%d/edit\n", cfg.zone, scenario.ID)
	}

	fmt.Println("\nnote: activate scenarios manually in Make.com UI")
}

func readFeeds(path string) ([]feed, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var feeds []feed
	for i, row := range records {
		if len(row) < 2 {
			return nil, fmt.Errorf("line %d: expected 2 columns (name, url), got %d", i+1, len(row))
		}
		name := strings.TrimSpace(row[0])
		url := strings.TrimSpace(row[1])
		if name == "" || url == "" {
			continue
		}
		feeds = append(feeds, feed{name: name, url: url})
	}
	return feeds, nil
}

func loadConfig() config {
	return config{
		apiToken:       requireEnv("MAKE_API_TOKEN"),
		zone:           getEnv("MAKE_ZONE", "eu1"),
		teamID:         requireInt("MAKE_TEAM_ID"),
		folderID:       getInt("MAKE_FOLDER_ID", 0),
		intervalSec:    getInt("MAKE_INTERVAL_SEC", 900),
		scenarioName:   getEnv("SCENARIO_NAME", "AutoGA Digest"),
		rssFeedsFile:   getEnv("RSS_FEEDS", ".rss_feeds.csv"),
		scraperURL:     requireEnv("AUTOGA_URL"),
		scraperAPIKey:  getEnv("AUTOGA_API_KEY", ""),
		llmAPIURL:      requireEnv("LLM_API_URL"),
		llmAPIKey:      requireEnv("LLM_API_KEY"),
		llmModel:       requireEnv("LLM_MODEL"),
		telegramChatID:       getEnv("TELEGRAM_CHAT_ID", ""),
		telegramConnectionID: getInt("TELEGRAM_CONNECTION_ID", 0),
	}
}

func requireEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
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
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
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
