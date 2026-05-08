package erxinscore

import (
	"fmt"
	"math"
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
		days += 30
	}
	if months < 0 {
		years--
		months += 12
	}
	totalMonths := float64(years*12+months) + float64(days)/30.0
	return Age{
		Years:              years,
		Months:             months,
		Days:               days,
		TotalMonths:        totalMonths,
		TotalMonthsRounded: math.Round(totalMonths*10) / 10,
	}, nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
