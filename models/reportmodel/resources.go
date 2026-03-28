package reportmodel

type ReportResource struct {
    Field any `json:"field"`
}

func (r Report) ToResource() ReportResource {
    return ReportResource{
        Field: r.Field,
    }
}
