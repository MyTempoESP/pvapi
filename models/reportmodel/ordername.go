package reportmodel

import (
	"fmt"
	"strings"

	"linhazero.com/pvapi/models/consignmentmodel"
)

func BuildConsignmentOrderName(monitor string, pickups []consignmentmodel.Pickup) string {
	if len(pickups) == 0 {
		return "N/A"
	}

	names := make([]string, 0, len(pickups))
	for _, p := range pickups {
		names = append(names, p.Name)
	}

	if len(names) <= 3 {
		return fmt.Sprintf("%s para %s", monitor, strings.Join(names, ", "))
	}

	display := strings.Join(names[:3], ", ")
	remaining := len(names) - 3
	return fmt.Sprintf("%s para %s, +%d outros", monitor, display, remaining)
}
