package makecom

import "encoding/json"

// ScenarioConfig holds all values needed to construct the scenario blueprint.
type ScenarioConfig struct {
	Name           string
	RSSFeedURL     string
	ScraperURL     string
	ScraperAPIKey  string
	LLMAPIUrl      string
	LLMAPIKey      string
	LLMModel       string
	TelegramChatID string
}

// BuildBlueprint constructs the Make.com scenario blueprint for the
// RSS → autoga scraper → LLM digest → Telegram flow.
//
// Data flow between modules uses Make.com template syntax: {{moduleId.field}}.
// Module IDs: 1=RSS, 2=scraper HTTP, 3=LLM HTTP, 4=Telegram.
//
// NOTE: article content is interpolated directly into the LLM JSON body.
// If articles contain characters that break JSON (quotes, backslashes),
// the LLM request may malform. Add a JSON:CreateJSON module in Make.com UI
// to handle encoding robustly if needed.
func BuildBlueprint(cfg ScenarioConfig) Blueprint {
	return Blueprint{
		Name: cfg.Name,
		Flow: []Module{
			rssModule(cfg.RSSFeedURL),
			scraperModule(cfg.ScraperURL, cfg.ScraperAPIKey),
			llmModule(cfg.LLMAPIUrl, cfg.LLMAPIKey, cfg.LLMModel),
			telegramModule(cfg.TelegramChatID),
		},
		Metadata: BlueprintMetadata{
			Version: 1,
			Scenario: ScenarioOptions{
				RoundTrips: 1,
				MaxErrors:  3,
				AutoCommit: true,
				Sequential: false,
			},
		},
	}
}

func rssModule(feedURL string) Module {
	return Module{
		ID:      1,
		Module:  "rss:watch",
		Version: 3,
		Mapper: map[string]any{
			"url":        feedURL,
			"maxResults": 1,
		},
		Metadata: ModuleMetadata{Designer: Designer{X: 0, Y: 0}},
	}
}

func scraperModule(scraperURL, apiKey string) Module {
	body := mustJSON(map[string]any{
		"urls": []string{"{{1.link}}"},
	})
	return Module{
		ID:      2,
		Module:  "http:ActionSendData",
		Version: 3,
		Mapper: map[string]any{
			"url":    scraperURL + "/scrape",
			"method": "POST",
			"headers": []map[string]string{
				{"name": "Authorization", "value": "Bearer " + apiKey},
				{"name": "Content-Type", "value": "application/json"},
			},
			"bodyType":      "raw",
			"contentType":   "application/json",
			"body":          body,
			"parseResponse": true,
		},
		Metadata: ModuleMetadata{Designer: Designer{X: 300, Y: 0}},
	}
}

func llmModule(apiURL, apiKey, model string) Module {
	// Make.com template variables are interpolated at runtime before the body
	// is sent to the LLM API. Module 2 (scraper) returns results[1] (1-indexed).
	const prompt = "Напиши короткий дайджест цієї статті українською мовою (3-5 речень)." +
		"\n\nЗаголовок: {{2.data.results[1].title}}" +
		"\n\nТекст: {{2.data.results[1].content}}"

	body := mustJSON(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	return Module{
		ID:      3,
		Module:  "http:ActionSendData",
		Version: 3,
		Mapper: map[string]any{
			"url":    apiURL,
			"method": "POST",
			"headers": []map[string]string{
				{"name": "Authorization", "value": "Bearer " + apiKey},
				{"name": "Content-Type", "value": "application/json"},
			},
			"bodyType":      "raw",
			"contentType":   "application/json",
			"body":          body,
			"parseResponse": true,
		},
		Metadata: ModuleMetadata{Designer: Designer{X: 600, Y: 0}},
	}
}

func telegramModule(chatID string) Module {
	// Module 3 (LLM) returns choices[1].message.content (1-indexed).
	const text = "{{2.data.results[1].title}}" +
		"\n\n{{3.data.choices[1].message.content}}" +
		"\n\nДжерело: {{1.link}}"

	return Module{
		ID:      4,
		Module:  "telegram:ActionSendMessage",
		Version: 1,
		Mapper: map[string]any{
			"chatId": chatID,
			"text":   text,
		},
		Metadata: ModuleMetadata{Designer: Designer{X: 900, Y: 0}},
	}
}

// mustJSON marshals v to a JSON string. Panics on error (only called with
// static, well-formed values so this should never happen in practice).
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
