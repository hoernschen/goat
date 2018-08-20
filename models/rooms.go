package models

//Room with active User
type Room struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Members    []User `json:"members,omitempty"`
	maxMembers int
}
