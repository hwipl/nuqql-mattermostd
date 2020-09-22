package cmd

// buddy stores information about a buddy
type buddy struct {
	user   string
	name   string
	status string
}

// newBuddy creates a new buddy with user (ID), name and status
func newBuddy(user, name, status string) *buddy {
	return &buddy{
		user:   user,
		name:   name,
		status: status,
	}
}
