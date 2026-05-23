package domain

type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Details []ValidationDetail
}

func (e ValidationError) Error() string {
	return "validation failed"
}

type ValidationCollector struct {
	details []ValidationDetail
}

func (v *ValidationCollector) Add(field, message string) {
	if v == nil {
		return
	}
	v.details = append(v.details, ValidationDetail{Field: field, Message: message})
}

func (v *ValidationCollector) AddIf(condition bool, field, message string) {
	if condition {
		v.Add(field, message)
	}
}

func (v *ValidationCollector) Err() error {
	if v == nil || len(v.details) == 0 {
		return nil
	}
	return ValidationError{Details: v.details}
}

func (v *ValidationCollector) HasErrors() bool {
	return v != nil && len(v.details) > 0
}

func (v *ValidationCollector) Details() []ValidationDetail {
	if v == nil {
		return []ValidationDetail{}
	}
	return v.details
}

func NewValidationCollector() *ValidationCollector {
	return &ValidationCollector{
		details: make([]ValidationDetail, 0),
	}
}
