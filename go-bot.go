package main

import (
	"fmt"
	"time"
	"strings"
	"strconv"
	"crypto/tls"
	"math/rand"
	"github.com/robfig/cron"
	irc "github.com/thoj/go-ircevent"
	"mdb/mdb"
	"mlog/mlog"
	"code.google.com/p/gcfg"
)
type Config struct {
	Global struct {
		Channel   string
		NickName  string
		Name      string
		IrcServer string
		IrcPort   string
		IrcPassword string
		MongoServer string
	}
	Notice struct {
		Mtg string
		Doc string
	}
}

func main() {

	// メンバー
	member := make(map[string]string)

	// コンフィグ
	var cfg Config
	if err := gcfg.ReadFileInto(&cfg, "src/config.gcfg"); err != nil {
		panic(err)
	}
	// systemグループチャネル
	system_chan := cfg.Global.Channel
	my_nick     := cfg.Global.NickName
	my_name     := cfg.Global.Name

	ircobj := irc.IRC(my_nick, my_name)
	ircobj.UseTLS    = true
	ircobj.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	ircobj.Password  = cfg.Global.IrcPassword
	if err := ircobj.Connect(fmt.Sprintf("%s:%s",cfg.Global.IrcServer,cfg.Global.IrcPort)); err != nil {
		fmt.Println(cfg.Global.IrcServer)
		fmt.Println(cfg.Global.IrcPort)
		fmt.Println(err)
	}

	// mongoDB
	mongo   := mdb.GetInstance(cfg.Global.MongoServer)
	session := mongo.GetSession()
	defer session.Close()
	log     := mlog.GetMulog(session)

	// ウェルカム処理
	ircobj.AddCallback("001", func (e *irc.Event) {
		ircobj.Join(system_chan)
	})

	// 入室
	ircobj.AddCallback("JOIN", func (e *irc.Event) {
		time.Sleep(100 * time.Millisecond)
		ircobj.Notice(system_chan, fmt.Sprintf("こんにちは、%sさん\n", e.Nick))
		if !strings.Contains(e.Nick, my_nick) {
			member[e.Nick] = e.Nick
		}
	})

	// プライベートメッセージ
	ircobj.AddCallback("PRIVMSG", func (e *irc.Event) {
		if e.Message() == ":ping" {
			// ピンポン
			time.Sleep(100 * time.Millisecond)
			ircobj.Notice(system_chan, "PONG")

		} else if e.Message() == ":dice" {
			// ダイス
			time.Sleep(100 * time.Millisecond)
			ircobj.Notice(system_chan, fmt.Sprintf("%sさんの出目は、%dです\n", e.Nick, rand.Intn(100)))

		} else if e.Message() == ":member" {
			// メンバー参照
			for _,v := range member {
				time.Sleep(10 * time.Millisecond)
				ircobj.Notice(system_chan, fmt.Sprintf(v))
			}

		} else if strings.Contains(e.Message(), ":add") {
			// メンバー追加
			name := strings.Fields(e.Message())[1:]
			if len(name) == 0 {
				member[e.Nick] = e.Nick
			} else {
				for _,v := range name {
					member[v] = v
				}
			}

		} else if strings.Contains(e.Message(), ":delete") {
			// メンバー削除
			name := strings.Fields(e.Message())[1:]
			for _,v := range name {
				delete(member, v)
			}

		} else if e.Message() == ":decide" {
			// 1人決める
			dice := rand.Intn(len(member))
			ii := 0
			for _, v := range member {
				if (ii == dice) {
					ircobj.Notice(system_chan, fmt.Sprintf("%sさんが選ばれました\n", v))
					break
				}
				ii += 1
			}

		} else if strings.Contains(e.Message(), ":history") {
			tmp := strings.Fields(e.Message())
			var limit int
			if len(tmp) > 1 {
				s,err := strconv.Atoi(tmp[1])
				if err != nil {
					limit = 5
				} else {
					limit = s
				}
			} else {
				limit = 5
			}
			results := log.Println(limit)
			for _,mulog := range results {
				ircobj.Notice(system_chan, mulog.ParseMessage())
			}
		} else {
			// ログ
			log.Insert(fmt.Sprint(time.Now()), e.Nick, e.Message())
		}
	})

	c := cron.New()
	c.AddFunc("0 5 */1 * * *", func() {
			ircobj.Notice(system_chan, "------------------------------------------")
			ircobj.Notice(system_chan, ":ping          ピンポン出来ます")
			ircobj.Notice(system_chan, ":dice          100面対のサイコロを振ります")
			ircobj.Notice(system_chan, ":decide        メンバーから1人選びます")
			ircobj.Notice(system_chan, ":member        メンバー一覧")
			ircobj.Notice(system_chan, ":add {str}     メンバー追加")
			ircobj.Notice(system_chan, ":delete  {str} メンバー削除")
			ircobj.Notice(system_chan, ":history {int} ヒストリー表示")
			ircobj.Notice(system_chan, "------------------------------------------")
		})

	//c.AddFunc("0 50 16 * * MON", func() {
	c.AddFunc("0 00 20 * * *", func() {
			ircobj.Notice(system_chan, "-------------------------")
			ircobj.Notice(system_chan, cfg.Notice.Mtg)
			ircobj.Notice(system_chan, "-------------------------")
		})

	c.AddFunc("0 2 20 * * *", func() {
			ircobj.Notice(system_chan, "---------------------")
			ircobj.Notice(system_chan, cfg.Notice.Doc)
			ircobj.Notice(system_chan, "---------------------")
		})

	c.Start()
	ircobj.Loop()

}
