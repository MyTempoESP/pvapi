package checkoutclient

import (
	"regexp"
	"strings"

	"linhazero.com/pvapi/util"
)

// CheckoutClientOutput holds the resolved identity fields derived from
// raw checkout request inputs. It is produced by Resolve and consumed
// by the controller and client model layers.
type CheckoutClientOutput struct {
	EffectiveEmail   string
	Password         string
	NormalizedPhone  string
	ValidatedCPF     string
	ClientName       string
	IsSyntheticEmail bool
}

var nonDigit = regexp.MustCompile(`[^0-9]`)

// stripNonDigits removes all non-digit characters.
func stripNonDigits(s string) string {
	return nonDigit.ReplaceAllString(s, "")
}

// validateCPF returns the CPF digits if exactly 11 digits, else empty.
func validateCPF(raw string) string {
	digits := stripNonDigits(raw)
	if len(digits) == 11 {
		return digits
	}
	return ""
}

// deriveName returns the provided name if non-empty, otherwise derives
// one from the effective email address.
func deriveName(provided, effectiveEmail string, nameGiven bool) string {
	if nameGiven {
		return provided
	}
	return util.DeriveClientName(effectiveEmail)
}

// Resolve takes raw checkout request fields and returns the resolved
// identity output. It is a pure function with no database or HTTP
// dependencies.
func Resolve(email, telephone, cpf, name string) CheckoutClientOutput {
	normalizedPhone := stripNonDigits(telephone)
	validCPF := validateCPF(cpf)
	validPhone := normalizedPhone

	nameProvided := strings.TrimSpace(name) != ""
	emailProvided := strings.TrimSpace(email) != ""
	phoneProvided := validPhone != ""

	var (
		effectiveEmail   string
		password         string
		clientName       string
		isSyntheticEmail bool
	)

	switch {
	case emailProvided:
		effectiveEmail = email
		isSyntheticEmail = false
		clientName = deriveName(name, effectiveEmail, nameProvided)

	case validCPF != "":
		effectiveEmail = validCPF + "@br"
		isSyntheticEmail = true
		clientName = deriveName(name, effectiveEmail, nameProvided)

	case phoneProvided:
		effectiveEmail = validPhone + "@tele"
		isSyntheticEmail = true
		clientName = deriveName(name, effectiveEmail, nameProvided)

	default:
		return CheckoutClientOutput{}
	}

	if phoneProvided {
		password = validPhone
	} else {
		pwd, err := util.GenerateRandomPassword(16)
		if err != nil {
			password = ""
		} else {
			password = pwd
		}
	}

	return CheckoutClientOutput{
		EffectiveEmail:   effectiveEmail,
		Password:         password,
		NormalizedPhone:  validPhone,
		ValidatedCPF:     validCPF,
		ClientName:       clientName,
		IsSyntheticEmail: isSyntheticEmail,
	}
}
