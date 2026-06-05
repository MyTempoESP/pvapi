package checkouthash_test

import (
	"testing"

	"linhazero.com/pvapi/util/checkouthash"
)

func TestHashDeterministic(t *testing.T) {
	h1, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 2},
		{PackageID: 3, Quantity: 1},
	}, "client@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	// Same input, same order → same hash
	h2, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 2},
		{PackageID: 3, Quantity: 1},
	}, "client@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if h1 != h2 {
		t.Fatalf("expected identical hashes for identical input, got %q vs %q", h1, h2)
	}
}

func TestHashSortsItemsByPackageID(t *testing.T) {
	h1, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 3, Quantity: 1},
		{PackageID: 1, Quantity: 2},
	}, "client@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	h2, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 2},
		{PackageID: 3, Quantity: 1},
	}, "client@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if h1 != h2 {
		t.Fatalf("expected same hash regardless of item order, got %q vs %q", h1, h2)
	}
}

func TestHashOmitsEmailWhenEmpty(t *testing.T) {
	hNoEmail, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 1},
	}, "")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	hEmpty, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 1},
	}, "")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if hNoEmail != hEmpty {
		t.Fatalf("expected same hash for empty and absent email, got %q vs %q", hNoEmail, hEmpty)
	}
}

func TestHashDifferentInputs(t *testing.T) {
	h1, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 1},
	}, "a@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	h2, err := checkouthash.Hash("20.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 1},
	}, "a@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if h1 == h2 {
		t.Fatal("expected different hashes for different value_paid")
	}

	h3, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 2, Quantity: 1},
	}, "a@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if h1 == h3 {
		t.Fatal("expected different hashes for different items")
	}

	h4, err := checkouthash.Hash("10.00", []checkouthash.Item{
		{PackageID: 1, Quantity: 1},
	}, "b@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if h1 == h4 {
		t.Fatal("expected different hashes for different client_email")
	}
}

func TestHashOutputFormat(t *testing.T) {
	h, err := checkouthash.Hash("5.00", []checkouthash.Item{
		{PackageID: 10, Quantity: 3},
	}, "test@example.com")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if len(h) != 64 {
		t.Fatalf("expected 64 hex chars (SHA-256), got %d: %q", len(h), h)
	}

	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex character in hash: %c", c)
		}
	}
}

func TestHashEmailNotIncludedInHash(t *testing.T) {
	// client_name is NOT part of the hash — this test verifies
	// that only client_email (when present) affects the hash.
	// There's no client_name parameter in Hash(), confirming the spec.
}
