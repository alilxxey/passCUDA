package config

import "fmt"
import "slices"

type User struct {
	Name             string `yaml:"name"`
	Role             UserRole
	TgUserID         int64    `yaml:"tg_user_id"`
	TgAllowedChatIDs []int64  `yaml:"tg_allowed_chat_ids"`
	EmvFingerprints  []string `yaml:"emv_fingerprints"`
}

func (u *User) String() string {
	return fmt.Sprintf("User(name=%s, role=%s, tg_user_id=%d)", u.Name, u.Role, u.TgUserID)
}

func (c *Config) FindUserById(id int64) (*User, error) {
	idx := slices.IndexFunc(c.Users, func(c User) bool { return c.TgUserID == id })
	if idx < 0 {
		return nil, fmt.Errorf("user with id: `{%d}` not found", id)
	}
	return &c.Users[idx], nil
}

func (c *Config) FindUserByCardFingerprint(emvFingerprint string) (*User, error) {
	idx := slices.IndexFunc(c.Users, func(c User) bool { return slices.Contains(c.EmvFingerprints, emvFingerprint) })
	if idx < 0 {
		return nil, fmt.Errorf("user with EMV fingerprint: `{%s}` not found", emvFingerprint)
	}
	return &c.Users[idx], nil
}

func (u *User) IsChatAllowedForUser(chatID int64) bool {
	if len(u.TgAllowedChatIDs) == 0 {
		return true
	}
	return slices.Contains(u.TgAllowedChatIDs, chatID)
}
