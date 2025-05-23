package validate

import (
	"fmt"
	"regexp"
)

func SQLIdentifier(i interface{}, k string) (warnings []string, errors []error) {
	v := i.(string)
	if (!regexp.MustCompile(`^[a-zA-Z0-9_.@#-]+$`).MatchString(v)) && (!regexp.MustCompile("SHARED ACCESS SIGNATURE").MatchString(v)) {
		errors = append(errors, fmt.Errorf(
			"invalid SQL identifier. SQL identifier allows letters, digits, @, $, #, . or _, start with letter, _, @ or # .Got %q", v))
	}

	if 1 > len(v) {
		errors = append(errors, fmt.Errorf("%q cannot be less than 1 character: %q", k, v))
	}

	if len(v) > 128 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 128 characters: %q %d", k, v, len(v)))
	}

	return
}

func SQLIdentifierPassword(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if len(v) < 8 {
		errors = append(errors, fmt.Errorf("length should equal to or greater than %d, got %d", 8, len(v)))
		return
	}

	if len(v) > 128 {
		errors = append(errors, fmt.Errorf("length should be equal to or less than %d, got %d", 128, len(v)))
		return
	}

	switch {
	case regexp.MustCompile(`^.*[a-z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[A-Z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[0-9]+.*$`).MatchString(v):
		return
	case regexp.MustCompile(`^.*[a-z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[A-Z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[\W]+.*$`).MatchString(v):
		return
	case regexp.MustCompile(`^.*[a-z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[\W]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[0-9]+.*$`).MatchString(v):
		return
	case regexp.MustCompile(`^.*[A-Z]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[\W]+.*$`).MatchString(v) && regexp.MustCompile(`^.*[0-9]+.*$`).MatchString(v):
		return
	default:
		errors = append(errors, fmt.Errorf("%q must contain characters from three of the categories - uppercase letters, lowercase letters, numbers and non-alphanumeric characters", k))
		return
	}
}
