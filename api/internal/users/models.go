package users

type User struct {
	Username string `json:"username"`
	Password string `json:"-"` // We never show the password
	Balance  *int64  `json:"balance,omitempty"`
	ID       int64  `json:"id"`
}
