package model

type (
	// UserContext is current context user
	UserContext struct {
		Authorized bool
		Subject    *Person
	}
)
