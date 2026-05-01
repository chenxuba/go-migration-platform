package service

import (
	"go-migration-platform/pkg/pep3template"
	"go-migration-platform/services/education/internal/model"
)

func pep3ItemRecordFields(itemNo int) []model.PEP3ItemRecordField {
	return convertPEP3ItemRecordFields(pep3template.ItemRecordFields(itemNo))
}

func pep3AllItemRecordFields() map[int][]model.PEP3ItemRecordField {
	return pep3ItemRecordFieldDefinitions()
}

func pep3ItemRecordFieldDefinitions() map[int][]model.PEP3ItemRecordField {
	definitions := pep3template.AllItemRecordFields()
	out := make(map[int][]model.PEP3ItemRecordField, len(definitions))
	for itemNo, fields := range definitions {
		out[itemNo] = convertPEP3ItemRecordFields(fields)
	}
	return out
}

func convertPEP3ItemRecordFields(fields []pep3template.ItemRecordField) []model.PEP3ItemRecordField {
	out := make([]model.PEP3ItemRecordField, 0, len(fields))
	for _, field := range fields {
		options := make([]model.PEP3ItemRecordFieldOption, 0, len(field.Options))
		for _, option := range field.Options {
			options = append(options, model.PEP3ItemRecordFieldOption{
				Value: option.Value,
				Label: option.Label,
			})
		}
		out = append(out, model.PEP3ItemRecordField{
			Key:         field.Key,
			Label:       field.Label,
			FieldType:   field.FieldType,
			DisplayType: field.DisplayType,
			Required:    field.Required,
			Placeholder: field.Placeholder,
			Options:     options,
		})
	}
	return out
}
