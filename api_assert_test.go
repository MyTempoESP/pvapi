package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
)

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, rec.Code, rec.Body.String())
	}
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("decode json response: %v; body=%s", err, rec.Body.String())
	}
	return body
}

func asMap(t *testing.T, value any) map[string]any {
	t.Helper()
	v, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", value)
	}
	return v
}

func asSlice(t *testing.T, value any) []any {
	t.Helper()
	v, ok := value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", value)
	}
	return v
}

func asString(t *testing.T, value any) string {
	t.Helper()
	v, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	return v
}

func asBool(t *testing.T, value any) bool {
	t.Helper()
	v, ok := value.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", value)
	}
	return v
}

func asInt(t *testing.T, value any) int {
	t.Helper()

	switch v := value.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			t.Fatalf("convert json number to int: %v", err)
		}
		return int(i)
	case float64:
		return int(v)
	case int:
		return v
	default:
		t.Fatalf("expected number, got %T", value)
		return 0
	}
}

func containsName(items []any, want string) bool {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, ok := obj["orderName"]
		if !ok {
			continue
		}
		s, ok := name.(string)
		if ok && s == want {
			return true
		}
	}
	return false
}

func containsID(items []any, want int) bool {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, ok := obj["id"]
		if !ok {
			continue
		}
		switch v := id.(type) {
		case json.Number:
			n, err := v.Int64()
			if err == nil && int(n) == want {
				return true
			}
		case float64:
			if int(v) == want {
				return true
			}
		case string:
			if string(v) == strconv.Itoa(want) {
				return true
			}
		}
	}
	return false
}

func findIDByFieldValue(t *testing.T, items []any, field, want string) int {
	t.Helper()
	obj := findObjectByStringField(t, items, field, want)
	return asInt(t, obj["id"])
}

func findObjectByIntField(t *testing.T, items []any, field string, want int) map[string]any {
	t.Helper()
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		v := obj[field]
		if v == nil {
			continue
		}
		if asInt(t, v) == want {
			return obj
		}
	}
	t.Fatalf("could not find item where %s=%d", field, want)
	return nil
}

func findObjectByStringField(t *testing.T, items []any, field, value string) map[string]any {
	t.Helper()
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		got, ok := obj[field].(string)
		if ok && got == value {
			return obj
		}
	}
	t.Fatalf("could not find item where %s=%q", field, value)
	return nil
}

func requireField(t *testing.T, obj map[string]any, field string) any {
	t.Helper()
	v, ok := obj[field]
	if !ok {
		t.Fatalf("expected field %q in object", field)
	}
	return v
}

func containsMonitorAndPickups(orderName, monitor string, pickupCount int) bool {
	prefix := monitor + " para "
	if len(orderName) < len(prefix) {
		return false
	}
	if orderName[:len(prefix)] != prefix {
		return false
	}

	// Count commas in the pickup portion
	suffix := orderName[len(prefix):]
	if suffix == "N/A" {
		return pickupCount == 0
	}

	// For N pickups with no +N suffix: N-1 commas
	// For 4+ pickups showing first 3: 2 commas + "+N outros"
	if pickupCount <= 3 {
		// N pickups → N-1 commas (e.g. 3 pickups → 2 commas)
		commas := 0
		for _, c := range suffix {
			if c == ',' {
				commas++
			}
		}
		return commas == pickupCount-1
	}

	// For >3 pickups: expect 2 commas and "+N outros"
	commas := 0
	for _, c := range suffix {
		if c == ',' {
			commas++
		}
	}
	return commas == 2
}


