package constant

// TOTPSecretStatus TOTP密钥状态
type TOTPSecretStatus string

const (
	TOTPStatusDisabled TOTPSecretStatus = "disabled"
	TOTPStatusEnabled  TOTPSecretStatus = "enabled"
	TOTPStatusPending  TOTPSecretStatus = "pending"
)

func (s TOTPSecretStatus) String() string {
	return string(s)
}

func (s TOTPSecretStatus) IsValid() bool {
	switch s {
	case TOTPStatusDisabled, TOTPStatusEnabled, TOTPStatusPending:
		return true
	default:
		return false
	}
}
