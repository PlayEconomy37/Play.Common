package validator

// Define a new Validator type which contains a map of validation errors
type Validator struct {
	Errors map[string]string `json:",omitempty"`
}

// New is a helper which creates a new Validator instance with an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Returns true if the errors map contains any entries
func (v Validator) HasErrors() bool {
	return len(v.Errors) != 0
}

// Adds an error message to the map (so long as no entry already exists for the given key)
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is not 'ok'
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}
