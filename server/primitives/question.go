package primitives

import (
	"strings"
	"unicode/utf8"
)

type Question string

func (q *Question) Validate() []string {
	errs := make([]string, 0)

	normalizedQuestion := strings.TrimSpace(string(*q))
	*q = Question(normalizedQuestion)

	if normalizedQuestion == "" {
		errs = append(errs, "Poll question is cannot be empty")
	}

	if utf8.RuneCountInString(normalizedQuestion) > 255 {
		errs = append(errs, "Poll question cannot be longer than 255 characters")
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
