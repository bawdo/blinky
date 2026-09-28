package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestSanitiseEscapesControlCharacters(t *testing.T) {
	cases := map[string]string{
		"desk":        "desk",
		"desk\tone":   `desk\tone`,
		"a\nb":        `a\nb`,
		"\x1b[31mred": `\x1b[31mred`,
		"书房":          "书房",
		"party 🎉":     "party 🎉",
	}
	for in, want := range cases {
		if got := Sanitise(in); got != want {
			t.Errorf("Sanitise(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTableAlignsASCII(t *testing.T) {
	var b bytes.Buffer
	err := Table(&b, []string{"ID", "NAME"}, [][]string{{"desk", "desk"}, {"BS073788-3.1", "-"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "ID" + strings.Repeat(" ", 12) + "NAME\n" +
		"desk" + strings.Repeat(" ", 10) + "desk\n" +
		"BS073788-3.1" + "  " + "-\n"
	if b.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", b.String(), want)
	}
}

func TestTableAlignsWideCharacters(t *testing.T) {
	var b bytes.Buffer
	rows := [][]string{{"书房", "BS1"}, {"🎉", "BS2"}, {"desk", "BS3"}}
	if err := Table(&b, []string{"NAME", "SERIAL"}, rows); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(b.String(), "\n"), "\n")
	want := -1
	for _, line := range lines {
		marker := "BS"
		if strings.HasPrefix(line, "NAME") {
			marker = "SERIAL"
		}
		col := runewidth.StringWidth(line[:strings.Index(line, marker)])
		if want == -1 {
			want = col
		}
		if col != want {
			t.Errorf("column starts at %d in %q, want %d", col, line, want)
		}
		if strings.HasSuffix(line, " ") {
			t.Errorf("trailing space in %q", line)
		}
	}
}

func TestTableEscapesControlCharacters(t *testing.T) {
	var b bytes.Buffer
	if err := Table(&b, []string{"NAME"}, [][]string{{"a\tb"}}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), "\t") || !strings.Contains(b.String(), `a\tb`) {
		t.Errorf("control character not escaped: %q", b.String())
	}
}

func TestFieldsAlignsValues(t *testing.T) {
	var b bytes.Buffer
	err := Fields(&b, "desk", [][2]string{{"Serial", "BS1"}, {"Manufacturer", "Agile"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "desk\n" +
		"  Serial:" + strings.Repeat(" ", 8) + "BS1\n" +
		"  Manufacturer:  Agile\n"
	if b.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", b.String(), want)
	}
}

func TestJSONIndentsAndKeepsEmptyArrays(t *testing.T) {
	var b bytes.Buffer
	if err := JSON(&b, []int{}); err != nil {
		t.Fatal(err)
	}
	if b.String() != "[]\n" {
		t.Errorf("got %q", b.String())
	}
	b.Reset()
	if err := JSON(&b, map[string]string{"id": "<desk>"}); err != nil {
		t.Fatal(err)
	}
	if b.String() != "{\n  \"id\": \"<desk>\"\n}\n" {
		t.Errorf("got %q", b.String())
	}
}
