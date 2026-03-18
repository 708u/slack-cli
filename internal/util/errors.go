package util

import "fmt"

// SlackCLIError is the base error type for all slack-cli errors.
type SlackCLIError struct {
	Message string
	Code    string
}

func (e *SlackCLIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s [%s]", e.Message, e.Code)
	}
	return e.Message
}

// ConfigurationError indicates a problem with user configuration.
type ConfigurationError struct {
	SlackCLIError
}

// NewConfigurationError creates a ConfigurationError with the given
// message.
func NewConfigurationError(message string) *ConfigurationError {
	return &ConfigurationError{
		SlackCLIError: SlackCLIError{
			Message: message,
			Code:    "CONFIGURATION_ERROR",
		},
	}
}

// ValidationError indicates invalid user input.
type ValidationError struct {
	SlackCLIError
}

// NewValidationError creates a ValidationError with the given message.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		SlackCLIError: SlackCLIError{
			Message: message,
			Code:    "VALIDATION_ERROR",
		},
	}
}

// ApiError indicates a Slack API failure.
type ApiError struct {
	SlackCLIError
}

// NewApiError creates an ApiError with the given message.
func NewApiError(message string) *ApiError {
	return &ApiError{
		SlackCLIError: SlackCLIError{
			Message: message,
			Code:    "API_ERROR",
		},
	}
}

// FileError indicates a file-system operation failure.
type FileError struct {
	SlackCLIError
}

// NewFileError creates a FileError with the given message.
func NewFileError(message string) *FileError {
	return &FileError{
		SlackCLIError: SlackCLIError{
			Message: message,
			Code:    "FILE_ERROR",
		},
	}
}
