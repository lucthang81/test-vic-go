package details_error

type DetailsError struct {
	message string
	details map[string]interface{}
}

func (err *DetailsError) Error() string {
	return err.message
}

func NewError(message string, details map[string]interface{}) *DetailsError {
	return &DetailsError{
		message: message,
		details: details,
	}
}

func (err *DetailsError) Details() map[string]interface{} {
	return err.details
}
