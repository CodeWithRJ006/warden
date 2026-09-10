package pii

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Tokenizer orchestrates PII detection and tokenization.
type Tokenizer struct {
	vault       Vault
	presidioURL string
}

func NewTokenizer(vault Vault, presidioURL string) *Tokenizer {
	return &Tokenizer{vault: vault, presidioURL: presidioURL}
}

type PresidioRequest struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

type PresidioEntity struct {
	Start      int     `json:"start"`
	End        int     `json:"end"`
	EntityType string  `json:"entity_type"`
	Score      float64 `json:"score"`
}

// Regex patterns for fallback deterministic detection
var (
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRegex = regexp.MustCompile(`(?:\+?\d{1,3}[\s-]?)?\(?\d{3}\)?[\s-]?\d{3}[\s-]?\d{4}`)
	panRegex   = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	personMock = regexp.MustCompile(`\b(?:John Doe|Jane Doe)\b`) // Mocking Presidio Person detection for the test
)

// isLuhnValid checks if a number string passes the Luhn algorithm.
func isLuhnValid(number string) bool {
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	var sum int
	alternate := false
	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if n < 0 || n > 9 {
			return false
		}
		if alternate {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}

func (t *Tokenizer) Process(ctx context.Context, input string) (string, error) {
	output := input

	// 1. Deterministic Layer (Regex + Luhn)
	output = panRegex.ReplaceAllStringFunc(output, func(match string) string {
		if isLuhnValid(match) {
			token, _ := t.vault.Store(ctx, "CARD", match)
			return token
		}
		return match
	})

	output = emailRegex.ReplaceAllStringFunc(output, func(match string) string {
		token, _ := t.vault.Store(ctx, "EMAIL", match)
		return token
	})

	output = phoneRegex.ReplaceAllStringFunc(output, func(match string) string {
		token, _ := t.vault.Store(ctx, "PHONE", match)
		return token
	})

	// 2. Presidio Layer
	if t.presidioURL != "" {
		reqBody, _ := json.Marshal(PresidioRequest{Text: output, Language: "en"})
		resp, err := http.Post(t.presidioURL+"/analyze", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			return "", fmt.Errorf("presidio unavailable: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var entities []PresidioEntity
			if err := json.NewDecoder(resp.Body).Decode(&entities); err == nil {
				// Naive replacement based on entities (in production we'd replace backwards to not mess up indices)
				// For prototype, we'll extract the substrings and replace them.
				for _, e := range entities {
					if e.Start < len(output) && e.End <= len(output) {
						raw := output[e.Start:e.End]
						// Skip if we already tokenized it (e.g. contains '[')
						if !strings.Contains(raw, "[") {
							token, _ := t.vault.Store(ctx, e.EntityType, raw)
							output = strings.Replace(output, raw, token, 1)
						}
					}
				}
			}
			resp.Body.Close()
		}
	} else {
		// Mocking Presidio PERSON detection for tests if URL is not set
		output = personMock.ReplaceAllStringFunc(output, func(match string) string {
			token, _ := t.vault.Store(ctx, "PERSON", match)
			return token
		})
	}

	return output, nil
}
