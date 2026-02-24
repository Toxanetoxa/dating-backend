package entity

type Filter struct {
	Radius  int
	AgeFrom int
	AgeTo   int
	Sex     *UserSex
}
