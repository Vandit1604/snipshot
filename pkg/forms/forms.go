package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// we can use this variable from the forms package when we will use the MatchesPattern() function. So we don't have to recompile the regex everytime.
var EmailRX = regexp.MustCompile("^[a-zA-Z\\_\\-\\.]+@[a-zA-Z\\_\\-\\.]+$")

type Form struct {
	url.Values
	Errors errors
}

func New(data url.Values) *Form {
	return &Form{
		Values: data,
		Errors: errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		if strings.TrimSpace(f.Get(field)) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

func (f *Form) MaxLength(field string, requiredLen int) {
	value := f.Get(field)
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) > requiredLen {
		f.Errors.Add(field, fmt.Sprintf("This field is too long(Maximum length is %d)", requiredLen))
	}
}

func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}

	for _, opt := range opts {
		if value == opt {
			return
		}
	}

	f.Errors.Add(field, "This field is invalid")
}

func (f *Form) MinLength(field string, requiredLen int) {
	value := f.Get(field)
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) < requiredLen {
		f.Errors.Add(field, fmt.Sprintf("This field is too short(Minimum length is %d)", requiredLen))
	}
}

func (f *Form) MatchesPattern(field string, emailRegEx *regexp.Regexp) {
	// check if value is empty
	value := f.Get(field)
	if value == "" {
		return
	}

	// if not empty, check if its a valid email
	if !emailRegEx.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
