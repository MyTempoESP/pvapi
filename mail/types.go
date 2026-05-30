package mail

// PickupItem represents a single product line item in a pickup group.
type PickupItem struct {
	ProductName string
	Quantity    int
}

// PickupGroup groups items under a single pickup location.
type PickupGroup struct {
	PickupName string
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
