package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "hello world ",
			expected: []string{"hello"},
		},
		{
			input:    "bulbasaur PIKACHU sQuirtle",
			expected: []string{"bulbasaur"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual[0]) != len(c.expected) {
			t.Errorf("Length does not match")
			t.Fail()
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[0]
			if word != expectedWord {
				t.Errorf("Expected word is not the same")
				t.Fail()
			}
		}
	}
}
