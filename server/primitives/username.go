package primitives

import "strings"

type Username string

func (u *Username) Validate() []string {
	errors := make([]string, 0)

	normalizedUsername := strings.ToLower(strings.TrimSpace(string(*u)))
	*u = Username(normalizedUsername)

	if string(*u) == "" {
		errors = append(errors, "Username is required")
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
