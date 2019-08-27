package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func doInternal(method, URL string, params map[string]string) (*http.Response, error) {

	//组装头部参数,头部的要多一些oauth_xxx参数
	headParams := make(map[string]string)
	for k, v := range params {
		headParams[k] = v
	}
	headParams["oauth_consumer_key"] = CONFIG["CONSUMER_KEY"]
	headParams["oauth_nonce"] = Nonce()
	headParams["oauth_signature_method"] = "HMAC-SHA1"
	headParams["oauth_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	headParams["oauth_token"] = CONFIG["TOKEN"]
	headParams["oauth_version"] = "1.0"
	//生成head string
	headString := EncodeParams(headParams)
	//组装signature base string
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, UrlEncoded(BaseUrl(URL)), UrlEncoded(headString))

	//生成签名
	headParams["oauth_signature"] = Signature(CONFIG["CONSUMER_SECRET"], CONFIG["TOKEN_SECRET"], signatureBaseString)
	//Make Authorization Header
	authHeaderBuf := bytes.NewBufferString("OAuth ")
	i := 0
	for _, k := range HeaderKeys {

		if strings.Index(k, "oauth_") != 0 {
			continue
		}

		if i > 0 {
			authHeaderBuf.WriteString(", ")
		}

		authHeaderBuf.WriteString(k)
		authHeaderBuf.WriteString("=")
		authHeaderBuf.WriteString("\"")
		authHeaderBuf.WriteString(UrlEncoded(headParams[k]))
		authHeaderBuf.WriteString("\"")
		i++
	}

	authHeader := authHeaderBuf.String()
	fullUrl := fmt.Sprintf("%s?%s", URL, EncodeParams(params))
	req, err := http.NewRequest(method, fullUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if DEBUG {
		req.Write(os.Stdout)
	}

	return http.DefaultClient.Do(req)
}

//定义twitter Rest接口

//DoHttp make post to twitter server
//we should append the params after the URL as parameters
//注意此处的url是干净的url无?
func DoHttp(method, URL string, params map[string]string) (map[string]interface{}, error) {
	resp, err := doInternal(method, URL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var bodyReader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "deflate":
		if bodyReader, err = zlib.NewReader(resp.Body); err != nil {
			return nil, err
		}
	case "gzip":
		if bodyReader, err = gzip.NewReader(resp.Body); err != nil {
			return nil, err
		}
	}

	if DEBUG {
		fmt.Fprint(os.Stdout, "\n------------------------------------------\n")
		fmt.Fprint(os.Stdout, fmt.Sprintf("%v %v\r\n", resp.Proto /* resp.ProtoMajor, resp.ProtoMinor,*/, resp.Status))
		resp.Header.Write(os.Stdout)
		fmt.Fprint(os.Stdout, "\r\n")
		if bs, err := ioutil.ReadAll(bodyReader); err != nil {
			return nil, err
		} else {
			os.Stdout.Write(bs)
			fmt.Fprint(os.Stdout, "\r\n")
			bodyReader = bytes.NewBuffer(bs)
		}
	}

	result := make(map[string]interface{})
	if err = json.NewDecoder(bodyReader).Decode(&result); err != nil {
		return nil, err
	}

	var errors interface{} = result["errors"]
	if errors != nil {
		items, _ := errors.([]interface{})
		item := items[0].(map[string]interface{})
		return nil, fmt.Errorf("%s [code:%v]", item["message"], item["code"])
	}

	return result, nil
}

//DoPost make post to twitter server
//we should append the params after the URL as parameters
//注意此处的url是干净的url无?
func DoPost(URL string, params map[string]string) (map[string]interface{}, error) {
	return DoHttp("POST", URL, params)
}

//DoGet make post to twitter server
//we should append the params after the URL as parameters
//注意此处的url是干净的url无?
func DoGet(URL string, params map[string]string) (map[string]interface{}, error) {
	return DoHttp("GET", URL, params)
}

//DoGets make post to twitter server
//we should append the params after the URL as parameters
//注意此处的url是干净的url无?
func DoGets(URL string, params map[string]string) ([]map[string]interface{}, error) {
	resp, err := doInternal("GET", URL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var bodyReader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "deflate":
		if bodyReader, err = zlib.NewReader(resp.Body); err != nil {
			return nil, err
		}
	case "gzip":
		if bodyReader, err = gzip.NewReader(resp.Body); err != nil {
			return nil, err
		}
	}

	if DEBUG {
		fmt.Fprint(os.Stdout, "\n------------------------------------------\n")
		fmt.Fprint(os.Stdout, fmt.Sprintf("%v %v\r\n", resp.Proto /* resp.ProtoMajor, resp.ProtoMinor,*/, resp.Status))
		resp.Header.Write(os.Stdout)
		fmt.Fprint(os.Stdout, "\r\n")
		if bs, err := ioutil.ReadAll(bodyReader); err != nil {
			return nil, err
		} else {
			os.Stdout.Write(bs)
			fmt.Fprint(os.Stdout, "\r\n")
			bodyReader = bytes.NewBuffer(bs)
		}
	}

	bufReader := bufio.NewReader(bodyReader)
	head, _ := bufReader.ReadByte()
	bufReader.UnreadByte()
	if head == '{' { //error happend
		result := make(map[string]interface{})
		if err = json.NewDecoder(bodyReader).Decode(&result); err != nil {
			return nil, err
		}

		var errors interface{} = result["errors"]
		if errors != nil {
			items, _ := errors.([]interface{})
			item := items[0].(map[string]interface{})
			return nil, fmt.Errorf("%s [code:%v]", item["message"], item["code"])
		}
	}

	result := make([]map[string]interface{}, 10)
	if err = json.NewDecoder(bufReader).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

//DoStream do stream api with twitter api
func DoStream(method, URL string, params map[string]string, callback func(item *TweetItem)) error {
	resp, err := doInternal(method, URL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//debug to console terminal
	//io.Copy(os.Stdout, resp.Body)
	//but we analyst

	//channel
	pool := make(chan *TweetItem, 100)
	//read from the channel and invoke the callback
	go func() {
		for {
			if item, ok := <-pool; ok {
				callback(item)
			} else {
				break
			}
		}
	}()

	br := bufio.NewReader(resp.Body)
	item := make(map[string]interface{})
	chunk, err := br.ReadBytes('\n')
	for err == nil {
		if len(chunk) < 10 {
			fmt.Println("chunk length is too small, we continue....", chunk)
			chunk, err = br.ReadBytes('\n')
			continue
		}
		chunk = TrimChunk(chunk)
		if err = json.Unmarshal(chunk, &item); err != nil {
			log.Println("json unmarshal error", chunk, string(chunk))
			return err
		}

		/*
			if err = json.NewDecoder(bytes.NewReader(chunk)).Decode(&item); err != nil {
				return err
			}
		*/

		//pool <- []string{item["created_at"].(string), item["user"].(map[string]interface{})["screen_name"].(string), item["text"].(string)}
		ti := Map2Tweet(item)
		pool <- ti
		if db != nil {
			mysqlPool <- ti
		}

		chunk, err = br.ReadBytes('\n')
	}

	log.Println("read chunk error", err)
	return err
}

//trimChunk trim the leading and after white space
func TrimChunk(chunk []byte) []byte {

	//the leading
	i, j := 0, len(chunk)
loop:
	for k, v := range chunk {
		switch v {
		case '\t', '\r', '\n', ' ':
			i = k + 1
		default:
			break loop
		}
	}

	//the after
loop2:
	for j > 0 {
		switch chunk[j-1] {
		case '\t', '\r', '\n', ' ':
		default:
			break loop2
		}
		j--
	}

	return chunk[i:j]
}
