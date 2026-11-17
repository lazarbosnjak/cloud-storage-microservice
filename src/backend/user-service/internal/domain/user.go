package domain

type User struct {
	MongoID   string `bson:"_id"`
	ID        string `bson:"userId"`
	Email     string `bson:"email"`
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}
