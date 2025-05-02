package primitives

import "github.com/alexedwards/argon2id"

type Password string

func (p *Password) Validate() []string {
	errors := make([]string, 0)

	if string(*p) == "" {
		errors = append(errors, "Password is required")
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (p *Password) Hash() string {
	hash, _ := argon2id.CreateHash(string(*p), argon2id.DefaultParams)
	return hash
}

func (p *Password) Verify(hash string) bool {
	match, _ := argon2id.ComparePasswordAndHash(string(*p), hash)
	return match
}
