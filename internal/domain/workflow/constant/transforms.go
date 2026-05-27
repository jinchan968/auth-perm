package constant

const (
	TransformRegexExtract   = "regex_extract"
	TransformRegexReplace   = "regex_replace"
	TransformTrim           = "trim"
	TransformMarkdownToText = "markdown_to_text"
	TransformExtractJSON    = "extract_json"
	TransformTruncate       = "truncate"
	TransformToUppercase    = "to_uppercase"
	TransformToLowercase    = "to_lowercase"
)

var validTransforms = map[string]bool{
	TransformRegexExtract:   true,
	TransformRegexReplace:   true,
	TransformTrim:           true,
	TransformMarkdownToText: true,
	TransformExtractJSON:    true,
	TransformTruncate:       true,
	TransformToUppercase:    true,
	TransformToLowercase:    true,
}

func IsValidTransform(t string) bool {
	return validTransforms[t]
}
