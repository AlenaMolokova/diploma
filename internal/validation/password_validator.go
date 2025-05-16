package validation

type PasswordValidator interface {
	ValidatePassword(password string) bool
}

type DefaultPasswordValidator struct{}

func NewDefaultPasswordValidator() *DefaultPasswordValidator {
	return &DefaultPasswordValidator{}
}

func (v *DefaultPasswordValidator) ValidatePassword(password string) bool {
	return len(password) >= 8
}
