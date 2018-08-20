package repository

import (
	"github.com/hoernschen/goat/models"
)

var rooms []models.Room

func GetRooms() []models.Room {
	return rooms
}

func GetRoom(id string) models.Room {
	room := models.Room{}
	for _, r := range rooms {
		if r.ID == id {
			room = r
		}
	}
	return room
}

func CreateRoom(room models.Room) {
	rooms = append(rooms, room)
}

func UpdateRoom(room models.Room) {
	for i, r := range rooms {
		if r.ID == room.ID {
			rooms[i] = room
		}
	}
}

func JoinRoom(roomid string, user models.User) {
	for i, r := range rooms {
		if r.ID == roomid {
			rooms[i].Members = append(r.Members, user)
		}
	}
}

func LeaveRoom(roomid string, userid string) {
	for i, r := range rooms {
		if r.ID == roomid {
			for j, m := range r.Members {
				if m.ID == userid {
					rooms[i].Members = append(r.Members[:j], r.Members[j+1:]...)
				}
			}
		}
	}
}

func DeleteRoom(id string) {
	for i, r := range rooms {
		if r.ID == id {
			rooms = append(rooms[:i], rooms[i+1:]...)
		}
	}
}
