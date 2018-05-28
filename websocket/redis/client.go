package redisclient

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

var client *redis.Client

type ShittyMusicRedisDAO struct {
	Addr     string
	Password string
	DB       int
}

type Upvotes struct {
	SongID  string
	Upvotes int
}

type Plays struct {
	SongID string
	Plays  int
}

func (c *ShittyMusicRedisDAO) Connect() {
	fmt.Println("connecting to redis client")
	client = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password, // no password set
		DB:       c.DB,       // use default DB
	})

	pong, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong, err)
	// Output: PONG <nil>
}

func (c *ShittyMusicRedisDAO) InitSong(id string, plays int, upvotes int) error {
	// err := client.Set("key", "value", 0).Err()
	err := client.Set("song:"+id+":plays", plays, 0).Err()
	if err != nil {
		panic(err)
	}
	err = client.Set("song:"+id+":upvotes", upvotes, 0).Err()
	if err != nil {
		panic(err)
	}
	return err
}

func (c *ShittyMusicRedisDAO) PlaySong(id string) error {
	key := "song:" + id + ":plays"
	fmt.Println(key)

	fmt.Println(client)

	val, err := client.Exists(key).Result()
	if err != nil {
		fmt.Println("check keys exist failed")
		panic(err)
	}
	fmt.Println(val)
	if val == 0 {
		err = c.InitSong(id, 0, 0)
		if err != nil {
			fmt.Println("init song failed")
			panic(err)
		}
	}

	err = client.Incr(key).Err()
	if err != nil {
		fmt.Println("incr song play failed")
		panic(err)
	}

	return err
	// return nil
}

func (c *ShittyMusicRedisDAO) GetPlays() ([]Plays, error) {
	keys, _, err := client.Scan(0, "song:*:plays", 1000).Result()
	if err != nil {
		return nil, err
	}

	var result []Plays

	for index := 0; index < len(keys); index++ {
		val, err := client.Get(keys[index]).Result()
		if err != nil {
			return nil, err
		}
		songID := strings.Split(keys[index], ":")[1]
		numPlays, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		play := Plays{
			SongID: songID,
			Plays:  numPlays,
		}

		result = append(result, play)
		fmt.Println("key", val)
	}
	return result, err
}

func (c *ShittyMusicRedisDAO) UpvoteSong(id string) error {
	key := "song:" + id + ":upvotes"
	err := client.Incr(key).Err()
	if err != nil {
		panic(err)
	}
	return err
}

func ExampleClient() {
	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}

func CreateRedistClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
}
