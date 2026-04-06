package reportmodel

/*
description	"a"
moment	"exposicao"
monitor	"Rodrigo"
occurrenceScope	"product"
orderId	1
priority	"media"
status	"aberto"
type	"falta_peca"
*/

type CreateReportRequest struct {
	Description   string `json:"description"`
	Moment        string `json:"moment"`
	Monitor       string `json:"monitor"`
	Scope         string `json:"occurrenceScope"`
	ConsignmentID int    `json:"orderId"`
	Priority      string `json:"priority"`
	Status        string `json:"status"`
	Type          string `json:"type"`
}

type UpdateReportRequest struct {
	Field any `json:"field"`
}

func (r CreateReportRequest) Prepare() ReportBlueprint {
	return ReportBlueprint{
		CreateReportRequest: r,
	}
}

func (r UpdateReportRequest) Prepare(id int) Report {
	return Report{
		ID: id,
	}
}
