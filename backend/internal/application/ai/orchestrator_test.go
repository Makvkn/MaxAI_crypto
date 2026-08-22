package ai

import "testing"

func TestValidateLLMAnswerJSON(t *testing.T) {
	got, ok := validateLLMAnswer(`{"answer":"Portfolio is $10."}`)
	if !ok || got != "Portfolio is $10." {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestValidateLLMAnswerRejectsEmptyJSON(t *testing.T) {
	if _, ok := validateLLMAnswer(`{"answer":""}`); ok {
		t.Fatal("expected rejection")
	}
}

func TestValidateLLMAnswerPlainText(t *testing.T) {
	got, ok := validateLLMAnswer("Hello")
	if !ok || got != "Hello" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}
