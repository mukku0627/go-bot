package mlog

import (
	"fmt"
	"labix.org/v2/mgo"
	"time"
)

type Mulog struct {
	session *mgo.Session
	Date    string
	Nick    string
	Message string
}

func GetMulog(s *mgo.Session) *Mulog {
	return &Mulog{s,"","",""}
}

func (self *Mulog) Insert(date,nick,message string) {
	session := self.session
	session.SetMode(mgo.Monotonic,true)

	c := session.DB("test").C("mulog")
	err := c.Insert(&Mulog{nil, date, nick ,message })
	if err != nil {
		panic(err)
	}
}

func (self *Mulog) Println(limit int) []Mulog {
	session := self.session
	c := session.DB("test").C("mulog")
	var results []Mulog
	err := c.Find(nil).Sort("-date").Limit(limit).All(&results)
	if err == nil {
		return results
	} else {
		return nil
	}
}

func (self *Mulog)viewData() string {
	t, _ := time.Parse("2006-01-02 15:04:05 -0700 JST",self.Date)
	return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d", t.Year(),t.Month(),t.Day(),t.Hour(),t.Minute(),t.Second())
}

func (self *Mulog)ParseMessage() string {
	return fmt.Sprintf("%s[%s] %s", self.viewData(),self.Nick,self.Message)
}
