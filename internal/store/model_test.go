package store

import "testing"

func TestWordCount(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"empty", "", 0},
		{"single word", "bonjour", 1},
		{"sentence", "Préciser la transition entre ces deux paragraphes.", 7},
		{"repeated whitespace", "un   deux\tdeux\n\ntrois", 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := wordCount(c.body); got != c.want {
				t.Errorf("wordCount(%q) = %d, want %d", c.body, got, c.want)
			}
		})
	}
}
