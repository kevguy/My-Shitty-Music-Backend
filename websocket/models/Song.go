package models

import "gopkg.in/mgo.v2/bson"

// Thumbnails represencts
type Thumbnails struct {
	Small  string `bson:"small" json:"small"`
	Medium string `bson:"medium" json:"medium"`
	Large  string `bson:"large" json:"large"`
}

// Song Represents a song, we uses bson keyword to tell the mgo driver how to name
// the properties in mongodb document
type Song struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Artist      string        `bson:"artist" json:"artist"`
	Country     string        `bson:"country" json:"country"`
	Date        string        `bson:"date" json:"date"`
	Description string        `bson:"description" json:"description"`
	Thumbnails  Thumbnails    `bson:"thumbnails" json:"thumbnails"`
	URL         string        `bson:"url" json:"url"`
	YouTubeID   string        `bson:"youtube_id" json:"youtube_id"`
	Upvotes     int           `bson:"upvotes" json:"upvotes"`
}
