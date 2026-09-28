package cmd

import "testing"

func TestColourSpellingOnlyRewritesTheSuffix(t *testing.T) {
	if got := colourSpelling(nil, "no-color-thing"); string(got) != "no-color-thing" {
		t.Errorf("got %q", got)
	}
	if got := colourSpelling(nil, "from-color"); string(got) != "from-colour" {
		t.Errorf("got %q", got)
	}
}
