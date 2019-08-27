package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var HeaderKeys = [8]string{"oauth_consumer_key", "oauth_callback", "oauth_nonce", "oauth_signature", "oauth_signature_method", "oauth_timestamp", "oauth_token", "oauth_version"}
var CONFIG = make(map[string]string)
var DEBUG bool

//TweetItemUser
type TweetItemUser struct {
	Id            int64  `json: "id"`
	Name          string `json: "name"`
	ScreenName    string `json:"screen_name"`
	Location      string `json: "location"`
	Description   string `json: "description"`
	StatusesCount int    `json: "statuses_count"`
}

//TweetItem
type TweetItem struct {
	CreatedAt string        `json: "created_at`
	Id        int64         `json: "id"`
	Text      string        `json: "text"`
	Lang      string        `json: "lang"`
	User      TweetItemUser `json: "user"`
}

//Map2Tweet convert map to tweetitem
func Map2Tweet(item map[string]interface{}) *TweetItem {
	ti := &TweetItem{}
	ti.CreatedAt = FormatDate(item["created_at"].(string))
	ti.Id = int64(item["id"].(float64))
	ti.Text = item["text"].(string)
	ti.Lang = item["lang"].(string)
	um := item["user"].(map[string]interface{})
	if um["description"] != nil {
		ti.User.Description = um["description"].(string)
	}
	ti.User.Name = um["name"].(string)
	ti.User.ScreenName = um["screen_name"].(string)
	if um["location"] != nil {
		ti.User.Location = um["location"].(string)
	}
	if um["description"] != nil {
		ti.User.Description = um["description"].(string)
	}
	if um["statuses_count"] != nil {
		ti.User.StatusesCount = int(um["statuses_count"].(float64))
	}
	return ti
}

//Nonce generator
//we should strip out all non-word characters
var nonWordReg = regexp.MustCompile("[^\\w]")

//Nonce create Nonce
func Nonce() string {
	bs := make([]byte, 32) //32bytes
	rand.Read(bs)
	line := base64.StdEncoding.EncodeToString(bs)
	return nonWordReg.ReplaceAllString(line, "")
}

//BaseUrl get base url
func BaseUrl(addr string) string {
	base := addr
	i := strings.Index(addr, "?")
	if i > -1 {
		base = addr[:i]
	}
	return base
}

//Signature generate signature
func Signature(consumerSecret, tokenSecret, signatureBaseString string) string {

	k := fmt.Sprintf("%s&%s", consumerSecret, tokenSecret)
	mac := hmac.New(sha1.New, []byte(k))
	mac.Write([]byte(signatureBaseString))
	bs := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(bs)

}
func SignatureJustOneSecret(secret, signatureBaseString string) string {

	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(signatureBaseString))
	bs := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(bs)

}

var urlPlusRegExp = regexp.MustCompile("\\+")

// UrlEncoded encodes a string like Javascript's encodeURIComponent()
func UrlEncoded(str string) string {
	/* 这个还不行
	u, err := url.Parse(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "UrlEncoded error %s %s", str, err)
		return ""
	}
	return u.String()
	*/
	line := url.QueryEscape(str)
	return urlPlusRegExp.ReplaceAllString(line, "%20")
}

//EncodeParams little like url.Values.Encode()
//http://stackoverflow.com/questions/13820280/encode-decode-urls
//由于url.QueryEscape将空格转成了+而不是%20，所以要使用这一个
func EncodeParams(v map[string]string) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		if buf.Len() > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(UrlEncoded(k) + "=")
		buf.WriteString(UrlEncoded(vs))
	}
	return buf.String()
}

//FormatDate format created_at to yyyy-MM-dd HH:mm:ss
func FormatDate(createdAt string) string {

	if ti, err := time.Parse(time.RubyDate, createdAt); err != nil {
		return createdAt
	} else {
		ti = ti.Add(time.Hour * 8) //时差8小时
		return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", ti.Year(), ti.Month(), ti.Day(), ti.Hour(), ti.Minute(), ti.Second())
	}
}

const SQL_INSERT = "insert into tb_tweets(tweet_id,from_user,create_time,db_time,text)values(?,?,?,NOW(),?)"

//StoreTweetToMySql
//create table tb_tweets(id bigint primary key auto_increment,from_user varchar(40),tweet_id bigint,create_time datetime,db_time datetime,text varchar(300))engine=MyISAM default charset=utf8;
func StoreTweetToMySql(item *TweetItem) {
	if _, err := db.Exec(SQL_INSERT, item.Id, item.User.ScreenName, item.CreatedAt, item.Text); err != nil {
		fmt.Println("Store mysql error", item.Text, err)
		if strings.Index(err.Error(), "1366") >= 0 { //证明是字符编码问题
			fmt.Println("We remove the invalid characters and try again")
			item.Text = EscapeMysqlInvalidString(item.Text)
			if _, err = db.Exec(SQL_INSERT, item.Id, item.User.ScreenName, item.CreatedAt, item.Text); err != nil {
				fmt.Println("Failed again! Omit saving to MySQL", item.Text, err)
			} else {
				fmt.Println("Store to mysql succeed after remove the invalid characters", item.Text)
			}
		}
	} else {
		//fmt.Printf("Insert %d row",result.RowsAffected)
	}
}

var InvalidRegexp = regexp.MustCompile("[^\u00bf-\uffff]")

func EscapeMysqlInvalidString(line string) string {
	r := make([]rune, 0, len(line))
	for _, v := range line {
		if v < 0xffff {
			r = append(r, v)
		}
	}
	//return InvalidRegexp.ReplaceAllString(line, "?")
	return string(r)
}
