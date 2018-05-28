package dao

import (
	"fmt"
	"log"
	"time"

	. "Redis-Exploration/websocket/models"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ShittyMusicDAO contains info needed to set up communication with MongoDB
type ShittyMusicDAO struct {
	Server   string
	Database string
	Addr     string
	Username string
	Password string
}

var db *mgo.Database

// The collection name
const (
	COLLECTION = "songs"
)

// Connect establishes a connection to database
func (m *ShittyMusicDAO) Connect() {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{m.Addr},
		Timeout:  60 * time.Second,
		Database: m.Database,
		Username: m.Username,
		Password: m.Password,
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatalf("CreateSession: %s\n", err)
	}

	db = mongoSession.DB(m.Database)

	fmt.Println("Connected to MongoDB")

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	// 	mongoSession.SetMode(mgo.Monotonic, true)
	//
	// 	// Create a wait group to manage the goroutines.
	// 	var waitGroup sync.WaitGroup
	//
	// 	// Perform 10 concurrent queries against the database.
	// 	waitGroup.Add(10)
	// 	for query := 0; query < 10; query++ {
	// 		go RunQuery(query, &waitGroup, mongoSession)
	// 	}
	//
	// 	// Wait for all the queries to complete.
	// 	waitGroup.Wait()
	// 	log.Println("All Queries Completed")
	// }
	//
}

// FindAllSongs finds list of songs
func (m *ShittyMusicDAO) FindAllSongs() ([]Song, error) {
	var songs []Song
	err := db.C(COLLECTION).Find(bson.M{}).All(&songs)
	return songs, err
}

// FindSongByID finds song by ID
func (m *ShittyMusicDAO) FindSongByID(id string) (Song, error) {
	var song Song
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&song)
	return song, err
}

// InsertSong inserts a song
func (m *ShittyMusicDAO) InsertSong(song Song) error {
	err := db.C(COLLECTION).Insert(&song)
	return err
}

// DeleteSong deletes a song
func (m *ShittyMusicDAO) DeleteSong(song Song) error {
	err := db.C(COLLECTION).Remove(&song)
	return err
}

// UpdateSong updates a song
func (m *ShittyMusicDAO) UpdateSong(song Song) error {
	err := db.C(COLLECTION).UpdateId(song.ID, &song)
	return err
}
