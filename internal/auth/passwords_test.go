package auth

import "testing"

func TestSufficientlyStrong(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"foobar", false},
		{"foooooooooo", false},
		{"12345678", true},
		{"abcdeabcdeabcde", false},
		{"abcdefghijklmno", true},
	}
	for _, test := range tests {
		if SufficientlyStrong(test.password) != test.expected {
			t.Errorf(`expected SufficientlyStrong("%s") to be %v, was %v\n`, test.password, test.expected, !test.expected)
		}
	}
}
