package main

/*
set mysql=root:root@tcp(127.0.0.1:3306)/test
set HTTP_PROXY=127.0.0.1:8380

set CONSUMER_SECRET=yous
set CONSUMER_KEY=xxx
set TOKEN_SECRET=xxxxdfd
set TOKEN=xcxdf
*/

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const usage = `
--------------------------------------------------------
#get specified user's twitter
tweets_stream.exe -person -q somename -c 10 -mysql %mysql%

#query twitter according the keyword
tweets_stream.exe -s -q trump -c 10 -mysql %mysql%

#post twitter [update status]
tweets_stream.exe -p -m "make earth great again"

#stream filter the twitter according the keyword
tweets_stream.exe -stream -q trump -mysql %mysql%
--------------------------------------------------------
##if you specify mysql, the tweet will store in mysql
--------------------------------------------------------
`

var keyword string
var count int
var message string
var isUpdateStatus bool
var isQueryTwitter bool
var isStream bool
var isGetPersonTwitter bool
var mySql string
var db *sql.DB
var mysqlPool = make(chan *TweetItem, 100)

//ready load config from os env
func ready() {
	flag.BoolVar(&DEBUG, "d", false, "debug mode")
	flag.StringVar(&keyword, "q", "", "query keyword")
	flag.IntVar(&count, "c", 1, "count numbers")
	flag.StringVar(&message, "m", "", "message text")
	flag.BoolVar(&isUpdateStatus, "p", false, "update twitter status")
	flag.BoolVar(&isQueryTwitter, "s", false, "search twitter")
	flag.BoolVar(&isStream, "stream", false, "stream filter")
	flag.BoolVar(&isGetPersonTwitter, "person", false, "Specify person")
	flag.StringVar(&mySql, "mysql", "", "mysql url [user:passwd@tcp(localhost:3306)/dbname], you should make sure [create table tb_tweets(id bigint primary key auto_increment,tweet_id bigint,from_user varchar(40),create_time datetime,db_time datetime,text varchar(300))engine=MyISAM default charset=utf8;]")

	flag.Parse()
	keys := []string{"CONSUMER_KEY", "CONSUMER_SECRET", "TOKEN", "TOKEN_SECRET"}
	for _, k := range keys {
		if v := os.Getenv(k); v == "" {
			log.Fatalf("Please provide environment value for [%s]\n", k)
		} else {
			CONFIG[k] = v
		}
	}
	if mySql == "" || os.Getenv("MYSQL") != "" {
		mySql = os.Getenv("MYSQL")
		log.Println("We take mysql url from environment by key MYSQL")
	}

}
func main() {
	fmt.Println(myName)
	ready()
	var err error
	if mySql != "" {
		if db, err = sql.Open("mysql", mySql); err != nil {
			log.Fatalln("Open mysql error!", err)
		}
		defer db.Close()
	}
	//mysql pool
	if db != nil {
		go func() {
			for {
				if item, ok := <-mysqlPool; ok {
					StoreTweetToMySql(item)
				} else {
					break
				}
			}
		}()
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println(usage)
	}

	if isUpdateStatus {
		if message == "" {
			flag.Usage()
			os.Exit(1)
		}
		if err = PostTwitter(message); err != nil {
			log.Fatalln(err)
		}
	} else if isQueryTwitter {
		if keyword == "" {
			flag.Usage()
			os.Exit(1)
		}
		if err = QueryTwitter(keyword, count); err != nil {
			log.Fatalln(err)
		}
	} else if isStream {
		if keyword == "" {
			flag.Usage()
			os.Exit(1)
		}
		if err = StreamFilterTwitter(keyword); err != nil {
			log.Fatalln(err)
		}
	} else if isGetPersonTwitter {
		if keyword == "" {
			flag.Usage()
			os.Exit(1)
		}
		if err = GetPersonTwitter(keyword, count); err != nil {
			log.Fatalln(err)
		}

	} else {
		flag.Usage()
	}

}

//PostTwitter post twitter to my account
func PostTwitter(status string) error {
	params := map[string]string{
		"status": status,
	}

	if result, err := DoPost("https://api.twitter.com/1.1/statuses/update.json", params); err != nil {
		return err
	} else {
		log.Println(result)
		return nil
	}
}

//QueryTwitter query twitter
func QueryTwitter(keyword string, count int) error {
	params := map[string]string{
		"q":                keyword,
		"count":            strconv.Itoa(count),
		"include_entities": "false",
	}

	if result, err := DoGet("https://api.twitter.com/1.1/search/tweets.json", params); err != nil {
		return err
	} else {
		var inf interface{} = result["statuses"]
		for i, v := range inf.([]interface{}) {
			m := v.(map[string]interface{})
			tweet := Map2Tweet(m)
			if db != nil {
				//mysqlPool <- tweet
				StoreTweetToMySql(tweet)
			}
			fmt.Printf("%d: [id=%d] [%v] [%v] %v\r\n", i+1, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName, tweet.Text)
		}
		return nil
	}
}

//GetPersonTwitter get person's twitter
func GetPersonTwitter(screenName string, count int) error {
	params := map[string]string{
		"screen_name":      screenName,
		"count":            strconv.Itoa(count),
		"include_entities": "false",
	}

	if items, err := DoGets("https://api.twitter.com/1.1/statuses/user_timeline.json", params); err != nil {
		return err
	} else {
		for i, item := range items {
			tweet := Map2Tweet(item)
			if db != nil {
				//mysqlPool <- tweet
				StoreTweetToMySql(tweet)
			}
			fmt.Printf("%d: [id=%d] [%v] [%v] %v\r\n", i+1, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName, tweet.Text)
		}
		return nil
	}
}

//StreamFilter stream filter
func StreamFilterTwitter(keyword string) error {
	params := map[string]string{
		"track":            keyword,
		"language":         "en,zh",
		"include_entities": "false",
		//"filter_level": "low",
		//"delimited": "length", //如果加上此参数，则twitter服务器返回的chunked每一个chunk前会有 长度+\r\n， 当前不加这个参数，则是直接的chunk+\r\n, 所以直接readLine(or ReadBytes('\n'))就可以了
	}

	i := 0

	return DoStream("POST", "https://stream.twitter.com/1.1/statuses/filter.json", params, func(item *TweetItem) {
		i++
		fmt.Printf("%d: [id=%v] [%v] [%v] %v\r\n", i, item.Id, item.CreatedAt, item.User.ScreenName, item.Text)
	})
}

//http://www.kammerl.de/ascii/AsciiSignature.php
//http://patorjk.com/software/taag/#p=display&c=bash&f=Big&t=David
//use font big
const myName = `
   _____              _     _   _______                _       
  |  __ \            (_)   | | |__   __|              | |      
  | |  | | __ ___   ___  __| |    | |_      _____  ___| |_ ___ 
  | |  | |/ _' \ \ / / |/ _' |    | \ \ /\ / / _ \/ _ \ __/ __|
  | |__| | (_| |\ V /| | (_| |    | |\ V  V /  __/  __/ |_\__ \
  |_____/ \__,_| \_/ |_|\__,_|    |_| \_/\_/ \___|\___|\__|___/

`
