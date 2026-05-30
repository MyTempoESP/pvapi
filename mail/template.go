package mail

import (
	_ "embed"
	"html/template"
	"strings"
)

const logoSrc = "https://pvauto.online/logo.png"

//go:embed pickup_confirmation.html
var pickupConfirmationSource string

// pickupTemplateData holds the data passed to the pickup confirmation email template.
type pickupTemplateData struct {
	OrderIDShort string // first 8 chars of UUID
	OrderIDFull  string // full UUID
	HasQrCode    bool
	Groups       []PickupGroup
	IsNewUser    bool
	Name         string
	Email        string
	Password     string
	LoginURL     string
	ResetURL     string
	LogoSrc      string
}

var pickupTemplate *template.Template

func init() {
	src := strings.TrimSpace(pickupConfirmationSource)
	tpl := template.Must(template.New("pickup_confirmation").Parse(src))
	pickupTemplate = tpl
}

// shortOrderID returns the first 8 chars of a UUID for display.
func shortOrderID(orderID string) string {
	if len(orderID) > 8 {
		return orderID[:8]
	}
	return orderID
}

// renderPickupConfirmation renders the pickup confirmation email HTML template.
func renderPickupConfirmation(data PickupEmailData) (string, error) {
	var buf strings.Builder
	err := pickupTemplate.Execute(&buf, pickupTemplateData{
		OrderIDShort: shortOrderID(data.OrderID),
		OrderIDFull:  data.OrderID,
		HasQrCode:    data.QrCode != "",
		Groups:       data.Groups,
		IsNewUser:    data.IsNewUser,
		Name:         data.Name,
		Email:        data.Email,
		Password:     data.Password,
		LoginURL:     data.LoginURL,
		ResetURL:     data.ResetURL,
		LogoSrc:      logoSrc,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
