package parser

import "testing"

func TestHTMLToText(t *testing.T) {
	got := HTMLToText("<p>Hello&nbsp;<strong>world</strong></p><ul><li>SQL</li></ul>")
	want := "Hello world\nSQL"
	if got != want {
		t.Fatalf("HTMLToText() = %q, want %q", got, want)
	}
}

func TestSalaryRange(t *testing.T) {
	min, max, ok := SalaryRange("Annual Salary: $115,830-$145,704")
	if !ok {
		t.Fatal("SalaryRange() did not find range")
	}
	if min != 115830 || max != 145704 {
		t.Fatalf("SalaryRange() = %d, %d; want 115830, 145704", min, max)
	}
}
