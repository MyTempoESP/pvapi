package categorymodel

type CategoryResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c Category) ToResource() CategoryResource {
	return CategoryResource{
		Name: c.Name,
		ID:   c.ID,
	}
}

