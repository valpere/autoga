package makecom

import (
	"encoding/json"
	"fmt"
)

// ScenarioConfig holds all values needed to construct the scenario blueprint.
type ScenarioConfig struct {
	Name                 string
	Zone                 string
	RSSFeedURL           string
	ScraperURL           string
	LLMAPIUrl            string
	LLMModel             string
	TelegramChatID       string
	TelegramConnectionID int
	UpstashURL           string
	UpstashTTLSec        int
	// Make.com keychain IDs — set up once in Make.com UI (Connections).
	// Auth is injected automatically at runtime; no API keys embedded in blueprint.
	KeychainUpstash int // Upstash REST API key
	KeychainScraper int // autoga API key
	KeychainLLM     int // Ollama Cloud API key
}

// BuildBlueprint constructs the Make.com scenario blueprint for the
// RSS → dedup check → autoga scraper → LLM digest → Telegram → dedup record flow.
//
// Module IDs:
//
//	1 = RSS feed
//	2 = Upstash GET (dedup check): skip if url hash already exists
//	3 = autoga /scrape
//	4 = Ollama LLM
//	5 = Telegram
//	6 = Upstash SET (record url hash with TTL)
func BuildBlueprint(cfg ScenarioConfig) Blueprint {
	return Blueprint{
		Name: cfg.Name,
		Flow: []Module{
			rssModule(cfg.RSSFeedURL),
			upstashCheckModule(cfg.UpstashURL, cfg.KeychainUpstash),
			scraperModule(cfg.ScraperURL, cfg.KeychainScraper),
			llmModule(cfg.LLMAPIUrl, cfg.KeychainLLM, cfg.LLMModel),
			telegramModule(cfg.TelegramChatID, cfg.TelegramConnectionID),
			upstashSetModule(cfg.UpstashURL, cfg.KeychainUpstash, cfg.UpstashTTLSec),
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

// upstashCheckModule calls GET /get/<md5(url)> on Upstash REST API.
// Returns {"result": null} if key does not exist (not yet seen).
// The downstream scraper module has a filter: process only when result is empty.
func upstashCheckModule(upstashURL string, keychainID int) Module {
	return Module{
		ID:      2,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "apiKey",
			"tlsType":            "",
			"proxyKeychain":      "",
			"apiKeyKeychain":     keychainID,
		},
		Mapper: map[string]any{
			"url":                      upstashURL + "/get/{{md5(1.url)}}",
			"method":                   "get",
			"parseResponse":            true,
			"stopOnHttpError":          true,
			"allowRedirects":           true,
			"shareCookies":             false,
			"requestCompressedContent": false,
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 300, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "API keyUse when the service requires an API key."},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
					"apiKeyKeychain":     map[string]any{"label": "Upstash API Key"},
				},
				"expect": map[string]any{
					"method":                   map[string]any{"label": "GET"},
					"headers":                  map[string]any{"mode": "chose"},
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

// upstashSetModule records the url hash in Upstash with the configured TTL.
// Uses GET /set/<key>/1/ex/<ttl> — Upstash REST supports commands as URL path.
func upstashSetModule(upstashURL string, keychainID int, ttlSec int) Module {
	return Module{
		ID:      6,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "apiKey",
			"tlsType":            "",
			"proxyKeychain":      "",
			"apiKeyKeychain":     keychainID,
		},
		Mapper: map[string]any{
			"url":                      fmt.Sprintf("%s/set/{{md5(1.url)}}/1/ex/%d", upstashURL, ttlSec),
			"method":                   "get",
			"parseResponse":            false,
			"stopOnHttpError":          false,
			"allowRedirects":           true,
			"shareCookies":             false,
			"requestCompressedContent": false,
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 1500, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "API keyUse when the service requires an API key."},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
					"apiKeyKeychain":     map[string]any{"label": "Upstash API Key"},
				},
				"expect": map[string]any{
					"method":                   map[string]any{"label": "GET"},
					"headers":                  map[string]any{"mode": "chose"},
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

func scraperModule(scraperURL string, keychainID int) Module {
	body := mustJSON(map[string]any{
		"urls": []string{"{{1.url}}"},
	})
	return Module{
		ID:      3,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "apiKey",
			"tlsType":            "",
			"proxyKeychain":      "",
			"apiKeyKeychain":     keychainID,
		},
		// Filter: only process if Upstash returned null (url not seen before).
		Filter: &ModuleFilter{
			Name: "Not yet seen",
			Conditions: [][]FilterCondition{{
				{A: "{{2.data.result}}", B: "", O: "notexist"},
			}},
		},
		Mapper: map[string]any{
			"url":                      scraperURL + "/scrape",
			"method":                   "post",
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
			Designer: Designer{X: 600, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "API keyUse when the service requires an API key."},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
					"apiKeyKeychain":     map[string]any{"label": "App API Key"},
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

func llmModule(apiURL string, keychainID int, model string) Module {
	// Module 3 (scraper) returns results[1] (1-indexed per Make.com convention).
	// Article content is normalized (no newlines) by the scraper, so embedding
	// it in a JSON string via rawBodyContent is safe.
	const prompt = "Напиши короткий дайджест цієї статті українською мовою (3-5 речень)." +
		"\n\nЗаголовок: {{3.data.results[1].title}}" +
		"\n\nТекст: {{3.data.results[1].content}}"

	body := mustJSON(map[string]any{
		"model":  model,
		"stream": false,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	})

	return Module{
		ID:      4,
		Module:  "http:MakeRequest",
		Version: 4,
		Parameters: map[string]any{
			"authenticationType": "apiKey",
			"tlsType":            "",
			"proxyKeychain":      "",
			"apiKeyKeychain":     keychainID,
		},
		Mapper: map[string]any{
			"url":                      apiURL,
			"method":                   "post",
			"contentType":              "custom",
			"contentTypeValue":         "application/json",
			"rawBodyContent":           body,
			"parseResponse":            true,
			"stopOnHttpError":          true,
			"allowRedirects":           true,
			"shareCookies":             false,
			"requestCompressedContent": true,
			"timeout":                  "90",
		},
		Metadata: ModuleMetadata{
			Designer: Designer{X: 900, Y: 0},
			Restore: map[string]any{
				"parameters": map[string]any{
					"authenticationType": map[string]any{"label": "API keyUse when the service requires an API key."},
					"tlsType":            map[string]any{"label": "Empty"},
					"proxyKeychain":      map[string]any{"label": "Choose a key"},
					"apiKeyKeychain":     map[string]any{"label": "Ollama API Key"},
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
	// Module 3 (scraper) and module 4 (LLM/Ollama).
	const text = "{{3.data.results[1].title}}" +
		"\n\n{{4.data.message.content}}" +
		"\n\nДжерело: {{3.data.results[1].url}}"

	m := Module{
		ID:      5,
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
			Designer: Designer{X: 1200, Y: 0},
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
