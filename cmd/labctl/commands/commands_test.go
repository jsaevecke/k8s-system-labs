package commands

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		value string
		want  Command
		ok    bool
	}{
		{value: "start", want: Start, ok: true},
		{value: "delete", want: Delete, ok: true},
		{value: "list", want: List, ok: true},
		{value: "unsupported", ok: false},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, ok := Parse(test.value)
			if ok != test.ok || got != test.want {
				t.Fatalf("Parse(%q) = (%q, %t), want (%q, %t)", test.value, got, ok, test.want, test.ok)
			}
		})
	}
}
