package pep3score

import (
	"fmt"
	"time"
)

func AgeAt(birthDate, assessmentDate time.Time) (Age, error) {
	birthDate = dateOnly(birthDate)
	assessmentDate = dateOnly(assessmentDate)
	if assessmentDate.Before(birthDate) {
		return Age{}, fmt.Errorf("assessment date %s is before birth date %s", assessmentDate.Format(time.DateOnly), birthDate.Format(time.DateOnly))
	}

	years := assessmentDate.Year() - birthDate.Year()
	months := int(assessmentDate.Month()) - int(birthDate.Month())
	days := assessmentDate.Day() - birthDate.Day()
	if days < 0 {
		months--
		// PEP-3 tester manual examples borrow 30 days when subtracting dates.
		days += 30
	}
	if months < 0 {
		years--
		months += 12
	}
	return Age{
		Years:              years,
		Months:             months,
		Days:               days,
		TotalMonthsForNorm: years*12 + months,
	}, nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
