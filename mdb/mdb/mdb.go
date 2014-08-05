package mdb

import (
	"labix.org/v2/mgo"
)

type Db struct {
	url string
	Msession *mgo.Session
}

func GetInstance(url string) *Db {
	return &Db{url, nil}
}

func (db *Db) GetSession() *mgo.Session {
	if db.Msession == nil {
		var err error
		db.Msession, err = mgo.Dial(db.url)
		if err != nil {
			panic(err)
		}
	}
	return db.Msession.Clone()
}
