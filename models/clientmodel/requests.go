package clientmodel

type CreateClientRequest struct {
    Field any `json:"field" binding:"required"`
}

type UpdateClientRequest struct {
    Field any `json:"field"`
}

func (c CreateClientRequest) Prepare() ClientBlueprint {
    return ClientBlueprint{
        Field: c.Field,
    }
}

func (c UpdateClientRequest) Prepare(id int) Client {
    return Client{
        ID:    id,
        Field: c.Field,
    }
}
