package reportmodel

type CreateReportRequest struct {
    Field any `json:"field" binding:"required"`
}

type UpdateReportRequest struct {
    Field any `json:"field"`
}

func (r CreateReportRequest) Prepare() ReportBlueprint {
    return ReportBlueprint{
        Field: r.Field,
    }
}

func (r UpdateReportRequest) Prepare(id int) Report {
    return Report{
        ID:    id,
        Field: r.Field,
    }
}
