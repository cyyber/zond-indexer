package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Cache struct {
	Key        []byte
	Value      []byte
	Expiration primitive.DateTime
	Type       string
}
