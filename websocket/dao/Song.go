package dao

import (
	"log"

	. "Redis-Exploration/websocket/models"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ShittyMusicDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "songs"
)

// Connect establishes a connection to database
func (m *ShittyMusicDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// FindAllSongs finds list of songs
func (m *ShittyMusicDAO) FindAllSongs() ([]Song, error) {
	var songs []Song
	err := db.C(COLLECTION).Find(bson.M{}).All(&songs)
	return songs, err
}

/**
package dao

import (
	"log"

	. "github.com/mlabouardy/movies-restapi/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)


// Find a movie by its id
func (m *MoviesDAO) FindById(id string) (Movie, error) {
	var movie Movie
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&movie)
	return movie, err
}

// Insert a movie into database
func (m *MoviesDAO) Insert(movie Movie) error {
	err := db.C(COLLECTION).Insert(&movie)
	return err
}

// Delete an existing movie
func (m *MoviesDAO) Delete(movie Movie) error {
	err := db.C(COLLECTION).Remove(&movie)
	return err
}

// Update an existing movie
func (m *MoviesDAO) Update(movie Movie) error {
	err := db.C(COLLECTION).UpdateId(movie.ID, &movie)
	return err
}
*/
