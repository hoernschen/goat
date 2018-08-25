package models

type Room struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Members []User `json:"members,omitempty"`
	public  bool
}
