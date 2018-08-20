package repository

import (
	"github.com/hoernschen/goat/models"
)

var users []models.User

func GetUsers() []models.User {
	if users == nil {
		users = models.Users{}
	}
	return users
}

func GetUser(id string) models.User {
	user := models.User{}
	for _, u := range users {
		if u.ID == id {
			user = u
		}
	}
	return user
}

func CreateUser(user models.User) {
	users = append(users, user)
}

func UpdateUser(user models.User) {
	for i, u := range users {
		if u.ID == user.ID {
			users[i] = user
		}
	}
}

func DeleteUser(id string) {
	for i, u := range users {
		if u.ID == id {
			users = append(users[:i], users[i+1:]...)
		}
	}
}
