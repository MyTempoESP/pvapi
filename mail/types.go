package mail

// PickupItem represents a single product line item in a pickup group.
type PickupItem struct {
	ProductName string
	Quantity    int
}

// PickupAddress holds the formatted address of a pickup location.
type PickupAddress struct {
	Formatted string
}

// PickupGroup groups items under a single pickup location.
type PickupGroup struct {
	PickupName string
	Address    *PickupAddress
	Items      []PickupItem
}

// PickupEmailData holds all data needed to render the pickup confirmation email.
type PickupEmailData struct {
	OrderID   string
	QrCode    string
	Groups    []PickupGroup
	IsNewUser bool
	Name      string
	Email     string
	Password  string
	LoginURL  string
	ResetURL  string
}
