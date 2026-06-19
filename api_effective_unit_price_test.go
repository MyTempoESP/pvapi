package main

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEffectiveUnitPrice(t *testing.T) {
	h := newAPIHarness(t)

	tests := []struct {
		name         string
		price        decimal.Decimal
		discount     decimal.Decimal
		discountType string
		want         decimal.Decimal
	}{
		{
			name:         "percentage_discount",
			price:        decimal.NewFromInt(100),
			discount:     decimal.NewFromInt(20),
			discountType: "percentage",
			want:         decimal.NewFromInt(80),
		},
		{
			name:         "fixed_discount",
			price:        decimal.NewFromInt(100),
			discount:     decimal.NewFromInt(20),
			discountType: "fixed",
			want:         decimal.NewFromInt(80),
		},
		{
			name:         "fixed_discount_exceeds_price",
			price:        decimal.NewFromInt(10),
			discount:     decimal.NewFromInt(15),
			discountType: "fixed",
			want:         decimal.NewFromInt(0),
		},
		{
			name:         "percentage_zero",
			price:        decimal.NewFromInt(100),
			discount:     decimal.NewFromInt(0),
			discountType: "percentage",
			want:         decimal.NewFromInt(100),
		},
		{
			name:         "percentage_full",
			price:        decimal.NewFromInt(100),
			discount:     decimal.NewFromInt(100),
			discountType: "percentage",
			want:         decimal.NewFromInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got decimal.Decimal
			err := h.pool.QueryRow(
				context.Background(),
				`SELECT effective_unit_price($1, $2, $3)`,
				tt.price, tt.discount, tt.discountType,
			).Scan(&got)
			if err != nil {
				t.Fatalf("SELECT effective_unit_price: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("effective_unit_price(%s, %s, %s) = %s, want %s",
					tt.price.String(), tt.discount.String(), tt.discountType,
					got.String(), tt.want.String())
			}
		})
	}
}
