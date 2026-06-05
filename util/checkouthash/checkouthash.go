package checkouthash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Item is a checkout item entry for canonical hashing.
type Item struct {
	PackageID int `json:"package_id"`
	Quantity  int `json:"quantity"`
}

// Hash computes a SHA-256 content hash for a checkout request.
// Items are sorted by package_id for deterministic ordering.
// clientEmail is omitted from the hash when absent or empty.
func Hash(valuePaid string, items []Item, clientEmail string) (string, error) {
	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PackageID < sorted[j].PackageID
	})

	payload := struct {
		ValuePaid string `json:"value_paid"`
		Items     []Item `json:"items"`
		Email     *string `json:"client_email,omitempty"`
	}{
		ValuePaid: valuePaid,
		Items:     sorted,
	}

	if clientEmail != "" {
		payload.Email = &clientEmail
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal canonical payload: %w", err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
