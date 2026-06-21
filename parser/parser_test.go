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

func TestRequiredExperience(t *testing.T) {
	got := RequiredExperience("Experience:\nOne (1) year of experience analyzing enterprise systems.")
	if !got.Found {
		t.Fatal("RequiredExperience() did not find requirement")
	}
	if got.Min != 1 {
		t.Fatalf("RequiredExperience().Min = %d, want 1", got.Min)
	}
}

func TestRequiredExperienceIgnoresPreferredAndSubstitution(t *testing.T) {
	text := `Experience:
One (1) year of experience analyzing enterprise systems.
Substitution:
Additional experience may be substituted up to a maximum of two (2) years.
Preferred Qualifications
3-5 years of experience with Python programming.`

	got := RequiredExperience(text)
	if !got.Found || got.Min != 1 {
		t.Fatalf("RequiredExperience() = %+v, want min 1", got)
	}
}

func TestRequiredExperienceTwoYears(t *testing.T) {
	got := RequiredExperience("Two (2) years of verifiable experience as a Licensed Clinical Social Worker.")
	if !got.Found || got.Min != 2 {
		t.Fatalf("RequiredExperience() = %+v, want min 2", got)
	}
}
