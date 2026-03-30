package clientmodel

type ClientResource struct {
    Field any `json:"field"`
}

func (c Client) ToResource() ClientResource {
    return ClientResource{
        Field: c.Field,
    }
}
