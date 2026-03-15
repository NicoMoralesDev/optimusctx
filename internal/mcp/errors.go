package mcp

const (
	errCodeValidation = -32002
	errCodeBounds     = -32003
)

type FieldErrorDetail struct {
	Field      string `json:"field"`
	Message    string `json:"message,omitempty"`
	Constraint string `json:"constraint,omitempty"`
	Max        int    `json:"max,omitempty"`
	Min        int    `json:"min,omitempty"`
	Received   any    `json:"received,omitempty"`
}

func newValidationError(message string, detail FieldErrorDetail) *ResponseError {
	code := errCodeValidation
	if detail.Constraint == "maximum" || detail.Constraint == "minimum" {
		code = errCodeBounds
	}

	data := map[string]any{
		"field": detail.Field,
	}
	if detail.Constraint != "" {
		data["constraint"] = detail.Constraint
	}
	if detail.Message != "" {
		data["details"] = detail.Message
	}
	if detail.Constraint == "maximum" || detail.Max != 0 {
		data["max"] = detail.Max
	}
	if detail.Constraint == "minimum" || detail.Min != 0 {
		data["min"] = detail.Min
	}
	if detail.Received != nil {
		data["received"] = detail.Received
	}

	return &ResponseError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func newMaximumExceededError(field string, max int, received any) *ResponseError {
	return newValidationError("request exceeds supported maximum", FieldErrorDetail{
		Field:      field,
		Constraint: "maximum",
		Max:        max,
		Received:   received,
	})
}

func newMinimumViolationError(field string, min int, received any) *ResponseError {
	return newValidationError("request is below the supported minimum", FieldErrorDetail{
		Field:      field,
		Constraint: "minimum",
		Min:        min,
		Received:   received,
	})
}

func newRequiredFieldError(field string) *ResponseError {
	return newValidationError("required field is missing", FieldErrorDetail{
		Field:      field,
		Constraint: "required",
	})
}

func newConflictFieldError(field string, received any, details string) *ResponseError {
	return newValidationError("request fields conflict", FieldErrorDetail{
		Field:      field,
		Constraint: "conflict",
		Received:   received,
		Message:    details,
	})
}
