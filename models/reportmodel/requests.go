package reportmodel

import (
	"encoding/json"
	"strconv"
)

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

// TODO: frontend sends orderId as a JSON string when occurrenceScope="product"
// (e.g. "1" instead of 1). FlexInt works around this until the frontend is fixed.
type FlexInt int

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexInt(i)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*f = FlexInt(i)
	return nil
}

type CreateReportRequest struct {
	Description string  `json:"description"`
	Moment      string  `json:"moment"`
	Monitor     string  `json:"monitor"`
	Scope       string  `json:"occurrenceScope"`
	OrderID     FlexInt `json:"orderId"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
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
