package reportmodel

import "time"

/*
 * {
      "id": "123",
      "orderId": "98765",
      "customerName": "Nome do Monitor Aqui",
      "moment": "exposicao",
      "type": "avaria",
      "status": "aberto",
      "description": "Produto chegou com o ecrã riscado",
      "priority": "alta",
      "occurrenceScope": "product",
      "createdAt": "2023-10-25T14:30:00.000000Z",
      "updatedAt": "2023-10-25T14:30:00.000000Z"
    }
*/

type ReportResource struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Moment      string    `json:"moment"`
	Monitor     string    `json:"customerName"`
	Monitor2    string    `json:"monitorName"`
	Scope       string    `json:"occurrenceScope"`
	OrderID     int       `json:"orderId"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (r Report) ToResource() ReportResource {
	orderID := 0
	if r.Scope == "product" && r.ProductID != nil {
		orderID = *r.ProductID
	} else if r.ConsignmentID != nil {
		orderID = *r.ConsignmentID
	}

	return ReportResource{
		ID:          r.ID,
		Description: r.Description,
		Moment:      r.Moment,
		Monitor:     r.ConsigneeName,
		Scope:       r.Scope,
		OrderID:     orderID,
		Priority:    r.Priority,
		Status:      r.Status,
		Type:        r.Type,
		CreatedAt:   r.CreatedAt,
	}
}
