package util

import (
	"strings"

	"linhazero.com/pvapi/constant"
)

func CutShipmentPrefix(
	searchQuery string,
) (id string, recordType string) {
	id, isShipment := strings.CutPrefix(
		searchQuery,
		constant.ShipmentPrefix,
	)

	if isShipment {
		recordType = "shipment"
	} else {
		recordType = "consignment"
	}

	return
}

