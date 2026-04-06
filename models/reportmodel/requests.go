package reportmodel

import "strconv"

/*
description	"a"
moment	"exposicao"
monitor	"Rodrigo"
occurrenceScope	"product"
orderId	"1"
priority	"media"
status	"aberto"
type	"falta_peca"
*/

type CreateReportRequest struct {
	Description         string `json:"description"`
	Moment              string `json:"moment"`
	Monitor             string `json:"monitor"`
	Scope               string `json:"occurrenceScope"`
	ConsignmentIDSTRING string `json:"orderId"` // i hate my job
	Priority            string `json:"priority"`
	Status              string `json:"status"`
	Type                string `json:"type"`
}

type UpdateReportRequest struct {
	Field any `json:"field"`
}

func (r CreateReportRequest) Prepare() (ReportBlueprint, error) {
	var zero ReportBlueprint

	consignmentID, err := strconv.Atoi(r.ConsignmentIDSTRING)

	if err != nil {
		return zero, err
	}

	return ReportBlueprint{
		ConsignmentID:       consignmentID,
		CreateReportRequest: r,
	}, nil
}

func (r UpdateReportRequest) Prepare(id int) Report {
	return Report{
		ID: id,
	}
}
