package main

import (
	ur "github.com/caselongo/user-registration-go/user-registration"
)

type UserSource struct {
	users map[string]ur.User
}

func NewUserSource() *UserSource {
	return &UserSource{
		users: make(map[string]ur.User),
	}
}

func (u *UserSource) Insert(user ur.User) error {
	u.users[user.Email] = user

	return nil
}

func (u *UserSource) Update(user ur.User) error {
	u.users[user.Email] = user

	return nil
}

func (u *UserSource) Delete(email string) error {
	delete(u.users, email)

	return nil
}

func (u *UserSource) Select(email string) (*ur.User, error) {
	user, ok := u.users[email]
	if ok {
		return &user, nil
	}

	return nil, nil
}
