package constant

const (
	OpContains    = "contains"
	OpNotContains = "not_contains"
	OpEquals      = "equals"
	OpMatches     = "matches"
	OpStartsWith  = "starts_with"
	OpEndsWith    = "ends_with"
	OpGT          = "gt"
	OpGTE         = "gte"
	OpLT          = "lt"
	OpLTE         = "lte"
	OpIsEmpty     = "is_empty"
	OpNotEmpty    = "not_empty"
)

var validOperators = map[string]bool{
	OpContains:    true,
	OpNotContains: true,
	OpEquals:      true,
	OpMatches:     true,
	OpStartsWith:  true,
	OpEndsWith:    true,
	OpGT:          true,
	OpGTE:         true,
	OpLT:          true,
	OpLTE:         true,
	OpIsEmpty:     true,
	OpNotEmpty:    true,
}

func IsValidOperator(op string) bool {
	return validOperators[op]
}
