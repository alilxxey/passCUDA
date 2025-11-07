package config

import "fmt"

type UserRole int

const (
	UserRoleUser UserRole = iota
	UserRoleAdmin
)

func (r UserRole) String() string {
	switch r {
	case UserRoleUser:
		return "user"
	case UserRoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func (r *UserRole) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch s {
	case "user":
		*r = UserRoleUser
	case "admin":
		*r = UserRoleAdmin
	default:
		return fmt.Errorf("unknown user role: %s", s)
	}
	return nil
}

func (r UserRole) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}
