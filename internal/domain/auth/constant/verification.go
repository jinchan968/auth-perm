package constant

// VerificationStatus 验证状态
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusExpired  VerificationStatus = "expired"
)

func (s VerificationStatus) IsValid() bool {
	switch s {
	case VerificationStatusPending, VerificationStatusVerified, VerificationStatusExpired:
		return true
	default:
		return false
	}
}

func (s VerificationStatus) String() string {
	return string(s)
}
