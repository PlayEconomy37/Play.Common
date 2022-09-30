package permissions

// Define a Permissions slice, which we will use to hold the permission codes for a single user
type Permissions []string

// Add a helper method to check whether the Permissions slice contains a specific
// permission code
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}
