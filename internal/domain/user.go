package domain

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DBUser struct {
	ID           int64
	Login        string
	PasswordHash string
}
