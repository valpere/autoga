package makecom

import "encoding/json"

// ScenarioConfig holds all values needed to construct the scenario blueprint.
type ScenarioConfig struct {
	Name                 string
	Zone                 string
	RSSFeedURL           string
	ScraperURL           string
	ScraperAPIKey        string
	LLMAPIUrl            string
	LLMAPIKey            string
	LLMModel             string
	TelegramChatID       string
	TelegramConnectionID int
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
			telegramModule(cfg.TelegramChatID, cfg.TelegramConnectionID),
		},
		Metadata: BlueprintMetadata{
			Instant: false,
			Version: 1,
			Scenario: ScenarioOptions{
				RoundTrips:            1,
				MaxErrors:             3,
				AutoCommit:            true,
				AutoCommitTriggerLast: true,
				Sequential:            false,
			},
			Designer: DesignerMeta{Orphans: []any{}},
			Zone:     cfg.Zone + ".make.com",
			Notes:    []any{},
		},
	}
}

func rssModule(feedURL string) Module {
	return Module{
		ID:      1,
		Module:  "rss:ActionReadArticles",
		Version: 4,
		Parameters: map[string]any{
			"include": []any{},
		},
		Mapper: map[string]any{
			"url":            feedURL,
			"username":       "",
			"password":       "",
			"filterDateFrom": "",
			"filterDateTo":   "",
			"maxResults":     "10",
			"gzip":           true,
		},
		Metadata: ModuleMetadata{Designer: Designer{X: 0, Y: 0}},
	}
}

func scraperModule(scraperURL, apiKey string) Module {
	body := mustJSON(map[string]any{
		"urls": []string{"{{1.url}}"},
	})
	return Module{
		ID:      2,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "noAuth",
			"tlsType":            "",
			"proxyKeychain":      "",
		},
		Mapper: map[string]any{
			"url":    scraperURL + "/scrape",
			"method": "post",
			"headers": []map[string]string{
				{"name": "Authorization", "value": "Bearer " + apiKey},
			},
			"contentType":              "custom",
			"contentTypeValue":         "application/json",
			"rawBodyContent":           body,
			"parseResponse":            true,
			"stopOnHttpError":          true,
			"allowRedirects":           true,
			"shareCookies":             false,
			"requestCompressedContent": true,
			"timeout":                  90,
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 300, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "No authentication"},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
				},
				"expect": map[string]any{
					"method":                   map[string]any{"label": "POST"},
					"headers":                  map[string]any{"mode": "chose"},
					"contentType":              map[string]any{"label": "Custom"},
					"parseResponse":            map[string]any{"mode": "chose"},
					"stopOnHttpError":          map[string]any{"mode": "chose"},
					"allowRedirects":           map[string]any{"mode": "chose"},
					"shareCookies":             map[string]any{"mode": "chose"},
					"requestCompressedContent": map[string]any{"mode": "chose"},
					"paginationType":           map[string]any{"label": "Empty"},
				},
			},
		},
	}
}

func llmModule(apiURL, apiKey, model string) Module {
	// Module 2 (scraper) returns results[1] (1-indexed per Make.com convention).
	// Article content is normalized (no newlines) by the scraper, so embedding
	// it in a JSON string via rawBodyContent is safe.
	const prompt = "Напиши короткий дайджест цієї статті українською мовою (3-5 речень)." +
		"\n\nЗаголовок: {{2.data.results[1].title}}" +
		"\n\nТекст: {{2.data.results[1].content}}"

	body := mustJSON(map[string]any{
		"model":  model,
		"stream": false,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	})

	return Module{
		ID:      3,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "noAuth",
			"tlsType":            "",
			"proxyKeychain":      "",
		},
		Mapper: map[string]any{
			"url":    apiURL,
			"method": "post",
			"headers": []map[string]string{
				{"name": "Authorization", "value": "Bearer " + apiKey},
			},
			"contentType":              "custom",
			"contentTypeValue":         "application/json",
			"rawBodyContent":           body,
			"parseResponse":            true,
			"stopOnHttpError":          true,
			"allowRedirects":           true,
			"shareCookies":             false,
			"requestCompressedContent": true,
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 600, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "No authentication"},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
				},
				"expect": map[string]any{
					"method":                   map[string]any{"label": "POST"},
					"headers":                  map[string]any{"mode": "chose"},
					"contentType":              map[string]any{"label": "Custom"},
					"parseResponse":            map[string]any{"mode": "chose"},
					"stopOnHttpError":          map[string]any{"mode": "chose"},
					"allowRedirects":           map[string]any{"mode": "chose"},
					"shareCookies":             map[string]any{"mode": "chose"},
					"requestCompressedContent": map[string]any{"mode": "chose"},
					"paginationType":           map[string]any{"label": "Empty"},
				},
			},
		},
	}
}

func telegramModule(chatID string, connectionID int) Module {
	// Module 3 (LLM/Ollama) returns message.content.
	const text = "{{2.data.results[1].title}}" +
		"\n\n{{3.data.message.content}}" +
		"\n\nДжерело: {{2.data.results[1].url}}"

	m := Module{
		ID:      4,
		Module:  "telegram:SendReplyMessage",
		Version: 1,
		Mapper: map[string]any{
			"chatId":                  chatID,
			"text":                    text,
			"messageThreadId":         "",
			"parseMode":               "",
			"replyToMessageId":        "",
			"replyMarkupAssembleType": "reply_markup_enter",
			"replyMarkup":             "",
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 900, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"__IMTCONN__": map[string]any{
						"label": "My Telegram Bot connection",
						"data": map[string]any{
							"scoped":     "true",
							"connection": "telegram",
						},
					},
				},
				"expect": map[string]any{
					"parseMode":               map[string]any{"label": "Empty"},
					"disableNotification":     map[string]any{"mode": "chose"},
					"replyMarkupAssembleType": map[string]any{"label": "Enter the Reply Markup"},
				},
			},
		},
	}
	if connectionID != 0 {
		m.Parameters = map[string]any{"__IMTCONN__": connectionID}
	}
	return m
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
