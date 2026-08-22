// Package openapi_test validates the API contract. The document is the source
// of truth, so a broken contract must fail the build rather than surface as a
// runtime mismatch with the frontend.
package openapi_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func loadDocument(t *testing.T) *openapi3.T {
	t.Helper()

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile("openapi.yaml")
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}
	return doc
}

func TestContractIsValid(t *testing.T) {
	doc := loadDocument(t)
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("contract is invalid: %v", err)
	}
}

// TestEveryFrontendRouteExists pins the routes the frontend already calls. The
// backend must serve exactly these paths for VITE_API_MODE to switch from mock
// to real without touching feature code.
func TestEveryFrontendRouteExists(t *testing.T) {
	doc := loadDocument(t)

	required := []string{
		"/auth/guest",
		"/auth/email/register",
		"/auth/email/login",
		"/auth/google",
		"/auth/refresh",
		"/auth/upgrade",
		"/auth/logout",
		"/auth/session",
		"/wallets",
		"/wallets/{walletId}",
		"/wallets/{walletId}/portfolio",
		"/wallets/{walletId}/performance",
		"/wallets/{walletId}/transactions",
		"/wallets/{walletId}/transactions/{transactionId}",
		"/ai/usage",
		"/ai/scenarios",
		"/ai/conversations",
		"/ai/conversations/{conversationId}/messages",
	}

	for _, path := range required {
		if doc.Paths.Find(path) == nil {
			t.Errorf("contract is missing %s, which the frontend already calls", path)
		}
	}
}

// TestMonetaryFieldsAreStrings guards the rule that money is never a JSON
// number: a float would silently lose precision (§97, §112).
func TestMonetaryFieldsAreStrings(t *testing.T) {
	doc := loadDocument(t)

	decimal := doc.Components.Schemas["Decimal"]
	if decimal == nil {
		t.Fatal("contract has no Decimal schema")
	}
	if !decimal.Value.Type.Is("string") {
		t.Fatalf("Decimal must serialize as a string, got %v", decimal.Value.Type)
	}
}

// TestPaginationIsCursorOnly guards the rule that collections never expose page
// or offset parameters (§109).
func TestPaginationIsCursorOnly(t *testing.T) {
	doc := loadDocument(t)

	forbidden := map[string]bool{"page": true, "offset": true, "page_size": true}
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			for _, param := range op.Parameters {
				if param.Value == nil || param.Value.In != "query" {
					continue
				}
				if forbidden[param.Value.Name] {
					t.Errorf("%s %s exposes offset pagination parameter %q", method, path, param.Value.Name)
				}
			}
		}
	}
}
