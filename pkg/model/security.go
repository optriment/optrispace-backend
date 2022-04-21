package model

type (
	// UserContext is current context user
	UserContext struct {
		Authenticated bool    `json:"authenticated"`
		Token         string  `json:"token,omitempty"`
		Subject       *Person `json:"subject,omitempty"`
	}
)

// InhouseRealm represents default inner realm
const InhouseRealm = "inhouse"
