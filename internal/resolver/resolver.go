package resolver

import (
	"fmt"
	"strings"

	"github.com/lovelyJason/openskills/internal/resource"
)

type AmbiguousError struct {
	Name    string
	Matches []resource.Resource
}

func (e *AmbiguousError) Error() string {
	var refs []string
	for _, m := range e.Matches {
		refs = append(refs, m.QualifiedName())
	}
	return fmt.Sprintf("ambiguous resource %q found in multiple marketplaces: %s",
		e.Name, strings.Join(refs, ", "))
}

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource %q not found in any marketplace", e.Name)
}

func Resolve(input string, available []resource.Resource) (*resource.Resource, error) {
	if strings.Contains(input, "@") {
		name, version := splitVersion(input)

		for i := range available {
			if available[i].QualifiedName() == name {
				r := available[i]
				if version != "" {
					r.Version = version
				}
				return &r, nil
			}
		}

		// "foo@1.2.3" where foo is a short name: version was extracted,
		// name is now just "foo", fall through to short-name resolution.
		if version != "" && !strings.Contains(name, "@") {
			return resolveShortName(name, version, available)
		}

		return nil, &NotFoundError{Name: input}
	}

	return resolveShortName(input, "", available)
}

func resolveShortName(name, version string, available []resource.Resource) (*resource.Resource, error) {
	var matches []resource.Resource
	for _, r := range available {
		if r.Name == name {
			matches = append(matches, r)
		}
	}

	switch len(matches) {
	case 0:
		return nil, &NotFoundError{Name: name}
	case 1:
		r := matches[0]
		if version != "" {
			r.Version = version
		}
		return &r, nil
	default:
		return nil, &AmbiguousError{Name: name, Matches: matches}
	}
}

func ResolveMany(inputs []string, available []resource.Resource) (resolved []resource.Resource, errors []error) {
	for _, input := range inputs {
		r, err := Resolve(input, available)
		if err != nil {
			errors = append(errors, err)
		} else {
			resolved = append(resolved, *r)
		}
	}
	return
}

// splitVersion splits "name@version" into (name, version).
// "foo@marketplace@1.2.3" → ("foo@marketplace", "1.2.3")
// "foo@1.2.3" where foo is a short name → handled by caller context
func splitVersion(input string) (string, string) {
	parts := strings.Split(input, "@")
	if len(parts) <= 1 {
		return input, ""
	}
	last := parts[len(parts)-1]
	if looksLikeVersion(last) {
		return strings.Join(parts[:len(parts)-1], "@"), last
	}
	return input, ""
}

func looksLikeVersion(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == 'v' || s[0] == 'V' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	return s[0] >= '0' && s[0] <= '9'
}
