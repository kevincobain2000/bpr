package pkg

import "strings"

type Slices struct{}

func NewSlices() *Slices {
	return &Slices{}
}

func (s *Slices) FilterByCSV(slice []string, csv string) []string {
	if csv == "" {
		return slice
	}

	csvMap := make(map[string]struct{})
	for _, item := range strings.Split(csv, ",") {
		csvMap[strings.TrimSpace(item)] = struct{}{}
	}

	// Filter the slice
	var result []string
	for _, item := range slice {
		if _, exists := csvMap[item]; exists {
			result = append(result, item)
		}
	}

	return result
}
