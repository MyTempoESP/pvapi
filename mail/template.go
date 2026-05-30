package mail

import (
	_ "embed"
	"encoding/base64"
	"html/template"
	"strings"
)

//go:embed welcome.html
var welcomeTemplateSource string

//go:embed logo.png
var logoPNG []byte

// welcomeTemplateData holds the data passed to the welcome email template.
type welcomeTemplateData struct {
	Name       string
	Email      string
	Password   string
	LoginURL   string
	ResetURL   string
	LogoSrc    template.URL
}

var welcomeTemplate *template.Template

func init() {
	src := strings.TrimSpace(welcomeTemplateSource)
	tpl := template.Must(template.New("welcome").Parse(src))
	welcomeTemplate = tpl
}

// logoDataURI returns the logo as a base64 data URI.
func logoDataURI() template.URL {
	enc := base64.StdEncoding.EncodeToString(logoPNG)
	return template.URL("data:image/png;base64," + enc)
}

// renderWelcome renders the welcome email HTML template with the given data.
func renderWelcome(name, email, password, loginURL, resetURL string) (string, error) {
	var buf strings.Builder
	err := welcomeTemplate.Execute(&buf, welcomeTemplateData{
		Name:      name,
		Email:     email,
		Password:  password,
		LoginURL:  loginURL,
		ResetURL:  resetURL,
		LogoSrc:   logoDataURI(),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
