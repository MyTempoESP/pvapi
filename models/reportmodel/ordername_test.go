package reportmodel

import (
	"testing"

	"linhazero.com/pvapi/models/consignmentmodel"
)

func TestBuildConsignmentOrderName(t *testing.T) {
	monitor := "João Silva"

	t.Run("zero_pickups_returns_na", func(t *testing.T) {
		got := BuildConsignmentOrderName(monitor, nil)
		if got != "N/A" {
			t.Fatalf("expected %q, got %q", "N/A", got)
		}

		got = BuildConsignmentOrderName(monitor, []consignmentmodel.Pickup{})
		if got != "N/A" {
			t.Fatalf("expected %q, got %q", "N/A", got)
		}
	})

	t.Run("one_pickup", func(t *testing.T) {
		got := BuildConsignmentOrderName(monitor, []consignmentmodel.Pickup{
			{Name: "Loja Centro"},
		})
		want := "João Silva para Loja Centro"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("two_pickups", func(t *testing.T) {
		got := BuildConsignmentOrderName(monitor, []consignmentmodel.Pickup{
			{Name: "Loja Centro"},
			{Name: "Loja Norte"},
		})
		want := "João Silva para Loja Centro, Loja Norte"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("three_pickups", func(t *testing.T) {
		got := BuildConsignmentOrderName(monitor, []consignmentmodel.Pickup{
			{Name: "Loja Centro"},
			{Name: "Loja Norte"},
			{Name: "Loja Sul"},
		})
		want := "João Silva para Loja Centro, Loja Norte, Loja Sul"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("five_pickups_shows_plus_remaining", func(t *testing.T) {
		got := BuildConsignmentOrderName(monitor, []consignmentmodel.Pickup{
			{Name: "Loja Centro"},
			{Name: "Loja Norte"},
			{Name: "Loja Sul"},
			{Name: "Loja Leste"},
			{Name: "Loja Oeste"},
		})
		want := "João Silva para Loja Centro, Loja Norte, Loja Sul, +2 outros"
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
