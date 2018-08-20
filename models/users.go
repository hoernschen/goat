package models

type Users []User

//User to go into Rooms
type User struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	ip   string
}
