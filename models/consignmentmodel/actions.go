package consignmentmodel

type AddProductAction struct {
	Products      []AddProductUnit
	ConsignmentID int
}

type PackagingAttributes struct {
	Color string
}

type AddProductUnitAction struct {
	AttrVal PackagingAttributes
	AddProductUnit
}

func (u AddProductUnit) Prepare() AddProductUnitAction {
	return AddProductUnitAction{
		AddProductUnit: u,
		AttrVal: PackagingAttributes{
			Color: u.Color,
		},
	}
}

func (p AddProductRequest) Prepare(consignmentID int) AddProductAction {
	return AddProductAction{
		Products:      p.Products,
		ConsignmentID: consignmentID,
	}
}
