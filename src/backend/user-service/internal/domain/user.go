package domain

type User struct {
	ID           string `bson:"_id"`
	Email        string `bson:"email"`
	FirstName    string `bson:"firstName"`
	LastName     string `bson:"lastName"`
	StorageQuota int64  `bson:"storageQuota"`
	StorageUsed  int64  `bson:"storageUsed"`
}

type StorageInfo struct {
	Quota int64 `bson:"storageQuota"`
	Used  int64 `bson:"storageUsed"`
}
