package user

type Email string

func (e Email) String() string { return string(e) }

type User struct {
	ID          string
	Email       Email
	DisplayName string
}
