package dto

// LoginResult 登录结果
type LoginResult struct {
	User    *UserDTO    `json:"user"`
	Account *AccountDTO `json:"account"`
	Token   string      `json:"token"`
	Session *SessionDTO `json:"session"`
}
