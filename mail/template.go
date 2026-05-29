package mail

import (
	_ "embed"
	"html/template"
	"strings"
)

//go:embed welcome.html
var welcomeTemplateSource string

// welcomeTemplateData holds the data passed to the welcome email template.
type welcomeTemplateData struct {
	Name      string
	Email     string
	Password  string
	LoginURL  string
	ResetURL  string
}

var welcomeTemplate *template.Template

func init() {
	src := strings.TrimSpace(welcomeTemplateSource)
	tpl := template.Must(template.New("welcome").Parse(src))
	welcomeTemplate = tpl
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
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
