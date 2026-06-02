package reportmodel

type ReportBlueprint struct {
	CreateReportRequest
}

type ReportWithConsigneeIDBlueprint struct {
	ConsigneeID int

	CreateReportRequest
}

type UpdateReportBlueprint struct {
	ID          int
	Status      *string
	Priority    *string
	Description *string
	Type        *string
}
