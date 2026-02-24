package entity

// Couple ...
type Couple struct {
	Distance int   `json:"distance"`
	Profile  *User `json:"profile"`
}
