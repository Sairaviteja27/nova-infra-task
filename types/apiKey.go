package types

type APIKey struct {
	Key     string `bson:"key"`
	Name    string `bson:"name,omitempty"`
	Revoked bool   `bson:"revoked,omitempty"`
}