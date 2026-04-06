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
	ID            int       `json:"id"`
	Description   string    `json:"description"`
	Moment        string    `json:"moment"`
	Monitor       string    `json:"customerName"`
	Scope         string    `json:"occurrenceScope"`
	ConsignmentID int       `json:"orderId"`
	Priority      string    `json:"priority"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (r Report) ToResource() ReportResource {
	return ReportResource{
		ID:            r.ID,
		Description:   r.Description,
		Moment:        r.Moment,
		Monitor:       r.Monitor,
		Scope:         r.Scope,
		ConsignmentID: r.ConsignmentID,
		Priority:      r.Priority,
		Status:        r.Status,
		Type:          r.Type,
		CreatedAt:     r.CreatedAt,
	}
}
