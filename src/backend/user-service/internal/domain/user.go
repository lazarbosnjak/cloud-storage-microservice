package domain

type User struct {
	ID        string `bson:"_id"`
	Email     string `bson:"email"`
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}
