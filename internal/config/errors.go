package config

type ValidationError struct {
}

func (e *ValidationError) Error() string {
	return "whatever needs to happen here"
}
