package primitives

import (
	"net"
	"strings"
)

type Email string

func (e *Email) Validate() []string {
	errors := make([]string, 0)

	normalizedEmail := strings.ToLower(strings.TrimSpace(string(*e)))
	*e = Email(normalizedEmail)

	if string(*e) == "" {
		errors = append(errors, "Email is required")
	}

	at := strings.LastIndex(normalizedEmail, "@")
	if at == -1 {
		errors = append(errors, "Email format is invalid")
	}

	host := normalizedEmail[at+1:]
	if _, err := net.LookupMX(host); err != nil {
		errors = append(errors, "Email is not reachable")
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
