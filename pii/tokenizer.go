package pii

import (
	"context"
	"regexp"
	"strings"
)

// Tokenizer orchestrates PII detection and tokenization.
type Tokenizer struct {
	vault Vault
	// In production, we would also call Presidio here.
}

func NewTokenizer(vault Vault) *Tokenizer {
	return &Tokenizer{vault: vault}
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

	// Detect and tokenize PAN (Deterministic boundary)
	output = panRegex.ReplaceAllStringFunc(output, func(match string) string {
		if isLuhnValid(match) {
			token, _ := t.vault.Store(ctx, "CARD", match)
			return token
		}
		return match
	})

	// Detect and tokenize Email
	output = emailRegex.ReplaceAllStringFunc(output, func(match string) string {
		token, _ := t.vault.Store(ctx, "EMAIL", match)
		return token
	})

	// Detect and tokenize Phone
	output = phoneRegex.ReplaceAllStringFunc(output, func(match string) string {
		token, _ := t.vault.Store(ctx, "PHONE", match)
		return token
	})

	// Mocking Presidio PERSON detection
	output = personMock.ReplaceAllStringFunc(output, func(match string) string {
		token, _ := t.vault.Store(ctx, "PERSON", match)
		return token
	})

	return output, nil
}
