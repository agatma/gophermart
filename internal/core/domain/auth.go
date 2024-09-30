package domain

type User struct {
	ID       int
	Login    string
	Password string
}

type UserIn struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	PasswordHash string `json:"-"`
}
