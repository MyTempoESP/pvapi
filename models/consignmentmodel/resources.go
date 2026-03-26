package consignmentmodel

type ConsignmentResource struct {
    Field any `json:"field"`
}

func (c Consignment) ToResource() ConsignmentResource {
    return ConsignmentResource{
        Field: c.Field,
    }
}
