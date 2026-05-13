package service

import (
	"strings"
	"unicode"
)

func normalizePEP3IEPImportDisplayText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(text))
	lastSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if lastSpace {
				continue
			}
			builder.WriteByte(' ')
			lastSpace = true
			continue
		}
		builder.WriteRune(r)
		lastSpace = false
	}
	return builder.String()
}

func normalizePEP3IEPImportMatchKey(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
