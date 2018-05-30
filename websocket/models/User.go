package models

import "gopkg.in/mgo.v2/bson"

// User represents for Mongo
type User struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	UUID        string        `bson:"_uuid" json:"uuid"`
	Username    string        `bson:"username" json:"username"`
	Name        string        `bson:"name" json:"name"`
	DisplayName string        `bson:"display_name" json:"display_name"`
	Password    string        `bson:"password" json:"password"`
	Type        string        `bson:"type" json:"type"`
	GoogleID    string        `bson:"google_id" json:"google_id"`
	ProfilePic  string        `bson:"profile_pic" json:"profile_pic"`
	Email       string        `bson:"email" json:"email"`
	Hearts      []string      `bson:"hearts" json:"hearts"`
}

// GoogleProfile for Google Auth
type GoogleProfile struct {
	ID, DisplayName, FullName, ImageURL, Email string
}
