package pii

import (
	"context"
	"strings"
	"testing"
)

func TestTokenizer_Process(t *testing.T) {
	vault := NewMemoryVault()
	tokenizer := NewTokenizer(vault)
	ctx := context.Background()

	input := "User John Doe has a card 4111 1111 1111 1111 and email john@example.com."
	
	sanitized, err := tokenizer.Process(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(sanitized, "John Doe") {
		t.Errorf("expected Person to be tokenized")
	}
	if strings.Contains(sanitized, "4111 1111 1111 1111") {
		t.Errorf("expected PAN to be tokenized")
	}
	if strings.Contains(sanitized, "john@example.com") {
		t.Errorf("expected Email to be tokenized")
	}

	if !strings.Contains(sanitized, "[PERSON_001]") {
		t.Errorf("expected [PERSON_001] in sanitized output, got: %s", sanitized)
	}
	if !strings.Contains(sanitized, "[CARD_001]") {
		t.Errorf("expected [CARD_001] in sanitized output, got: %s", sanitized)
	}
	if !strings.Contains(sanitized, "[EMAIL_001]") {
		t.Errorf("expected [EMAIL_001] in sanitized output, got: %s", sanitized)
	}

	// Verify vault
	raw, _ := vault.Retrieve(ctx, "[CARD_001]")
	if raw != "4111 1111 1111 1111" {
		t.Errorf("expected raw card in vault, got: %s", raw)
	}
}

func TestLuhnValid(t *testing.T) {
	if !isLuhnValid("4111 1111 1111 1111") {
		t.Errorf("expected valid luhn to pass")
	}
	if isLuhnValid("4111 1111 1111 1112") {
		t.Errorf("expected invalid luhn to fail")
	}
}
