package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type Person struct {
	Id    bson.ObjectId `bson:"_id,omitempty"`
	Name  string        `bson:"Name"`
	Phone string        `bson:"Phone"`
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("people")
	c.DropCollection()

	info := &mgo.CollectionInfo{
		Capped:   true,
		MaxBytes: 1000000,
	}

	err = c.Create(info)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for t := range ticker.C {
			err = c.Insert(&Person{
				Name:  "Ale",
				Phone: "+55 53 8116 9639"})
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Tick at", t)
		}
	}()

	time.Sleep(time.Second * 1)
	iter := c.Find(bson.M{}).Tail(1 * time.Second)

	result := Person{}

	for {
		lastId := 0
		for iter.Next(&result) {
			fmt.Println(result)
			lastId := result.Id
			fmt.Println(lastId)
		}
		if iter.Err() != nil {
			fmt.Println(iter.Err())
			iter.Close()
			return
		}
		if iter.Timeout() {
			continue
		}
		query := c.Find(bson.M{"_id": bson.M{"$gt": lastId}})
		iter = query.Sort("$natural").Tail(1 * time.Second)
	}
	iter.Close()

}
