package constant

import "time"

const (
	InvitationStatusActive      = "active"
	InvitationStatusUsed        = "used"
	InvitationStatusInvalidated = "invalidated"
	InvitationStatusExpired     = "expired"

	DefaultInvitationTTL = 7 * 24 * time.Hour
	InvitationCodeBytes  = 18
)
