package consignmentmodel

type CreateConsignmentRequest struct {
    Field any `json:"field" binding:"required"`
}

type UpdateConsignmentRequest struct {
    Field any `json:"field"`
}

func (c CreateConsignmentRequest) Prepare() ConsignmentBlueprint {
    return ConsignmentBlueprint{
        Field: c.Field,
    }
}

func (c UpdateConsignmentRequest) Prepare(id int) Consignment {
    return Consignment{
        ID:    id,
        Field: c.Field,
    }
}
