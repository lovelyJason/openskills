package resolver

import (
	"testing"

	"github.com/lovelyJason/openskills/internal/resource"
)

func makeResources() []resource.Resource {
	return []resource.Resource{
		{Name: "jira-to-code", Type: resource.TypePlugin, Marketplace: "tsai"},
		{Name: "fe-workflow", Type: resource.TypePlugin, Marketplace: "tsai"},
		{Name: "auth-helper", Type: resource.TypePlugin, Marketplace: "tsai"},
		{Name: "auth-helper", Type: resource.TypePlugin, Marketplace: "community"},
		{Name: "git-commit", Type: resource.TypeSkill, Marketplace: "tsai"},
	}
}

func TestResolve_UniqueShortName(t *testing.T) {
	res, err := Resolve("jira-to-code", makeResources())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "jira-to-code" || res.Marketplace != "tsai" {
		t.Errorf("got %s@%s, want jira-to-code@tsai", res.Name, res.Marketplace)
	}
}

func TestResolve_QualifiedName(t *testing.T) {
	res, err := Resolve("auth-helper@tsai", makeResources())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Marketplace != "tsai" {
		t.Errorf("got marketplace %s, want tsai", res.Marketplace)
	}
}

func TestResolve_QualifiedNameOtherMarketplace(t *testing.T) {
	res, err := Resolve("auth-helper@community", makeResources())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Marketplace != "community" {
		t.Errorf("got marketplace %s, want community", res.Marketplace)
	}
}

func TestResolve_Ambiguous(t *testing.T) {
	_, err := Resolve("auth-helper", makeResources())
	if err == nil {
		t.Fatal("expected AmbiguousError, got nil")
	}
	ambErr, ok := err.(*AmbiguousError)
	if !ok {
		t.Fatalf("expected *AmbiguousError, got %T", err)
	}
	if len(ambErr.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(ambErr.Matches))
	}
}

func TestResolve_NotFound(t *testing.T) {
	_, err := Resolve("nonexistent", makeResources())
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected *NotFoundError, got %T", err)
	}
}

func TestResolve_NotFoundQualified(t *testing.T) {
	_, err := Resolve("jira-to-code@unknown", makeResources())
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
}

func TestResolve_WithVersion(t *testing.T) {
	res, err := Resolve("jira-to-code@1.2.3", makeResources())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "1.2.3" {
		t.Errorf("got version %q, want 1.2.3", res.Version)
	}
}

func TestResolve_QualifiedWithVersion(t *testing.T) {
	res, err := Resolve("auth-helper@tsai@v2.0.0", makeResources())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Marketplace != "tsai" {
		t.Errorf("got marketplace %s, want tsai", res.Marketplace)
	}
	if res.Version != "v2.0.0" {
		t.Errorf("got version %q, want v2.0.0", res.Version)
	}
}

func TestResolveMany(t *testing.T) {
	resolved, errs := ResolveMany(
		[]string{"jira-to-code", "nonexistent", "git-commit"},
		makeResources(),
	)
	if len(resolved) != 2 {
		t.Errorf("expected 2 resolved, got %d", len(resolved))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
	}{
		{"foo", "foo", ""},
		{"foo@bar", "foo@bar", ""},
		{"foo@1.0.0", "foo", "1.0.0"},
		{"foo@v1.0.0", "foo", "v1.0.0"},
		{"foo@bar@1.0.0", "foo@bar", "1.0.0"},
		{"foo@bar@baz", "foo@bar@baz", ""},
	}
	for _, tt := range tests {
		name, version := splitVersion(tt.input)
		if name != tt.wantName || version != tt.wantVersion {
			t.Errorf("splitVersion(%q) = (%q, %q), want (%q, %q)",
				tt.input, name, version, tt.wantName, tt.wantVersion)
		}
	}
}

func TestLooksLikeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"V2.3", true},
		{"abc", false},
		{"", false},
		{"v", false},
		{"marketplace-name", false},
	}
	for _, tt := range tests {
		got := looksLikeVersion(tt.input)
		if got != tt.want {
			t.Errorf("looksLikeVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
