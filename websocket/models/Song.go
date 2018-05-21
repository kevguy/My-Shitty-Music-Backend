package models

import "gopkg.in/mgo.v2/bson"

// Song Represents a song, we uses bson keyword to tell the mgo driver how to name
// the properties in mongodb document
type Song struct {
	ID             bson.ObjectId `bson:"_id" json:"id"`
	Name           string        `bson:"name" json:"name"`
	ThumbnailMed   string        `bson:"thumbnail_med" json:"thumbnail_med"`
	ThumbnailLarge string        `bson:"thumbnail_large" json:"thumbnail_large"`
	Description    string        `bson:"description" json:"description"`
	URL            string        `bson:"url" json:"url"`
	Upvotes        int           `bson:"upvotes" json:"upvotes"`
}

// Movie Represents a movie, we uses bson keyword to tell the mgo driver how to name
// the properties in mongodb document
type Movie struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	CoverImage  string        `bson:"cover_image" json:"cover_image"`
	Description string        `bson:"description" json:"description"`
}
