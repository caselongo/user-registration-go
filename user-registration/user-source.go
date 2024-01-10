package user_registration

type UserSource interface {
	Insert(user User) error
	Update(user User) error
	Delete(email string) error
	Select(email string) (*User, error)
}
