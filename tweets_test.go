package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

//TestEncoding
func TestEncoding(t *testing.T) {

	//random := rand.Int()

	httpMethod := "POST"

	consumerSecret := "ISZYKSJCXusW8oGuh6EMV4M3CdyCSCUTtZs8jM7zpvmpW4qPBc"
	tokenSecret := "7FGTxHps6bSVf4MGvSFPbmNfVOhUmBrMNmPH6lf8zc6Sr"
	oauth_consumer_key := "W29UcpkImGKIT31LhRuYwOzJJ"
	oauth_nonce := Nonce()
	oauth_timestamp := fmt.Sprintf("%d", time.Now().Unix())
	oauth_token := "414627346-y2t6XrmamYi21NYwupQzH5KTUhJZhXx486zhyNPx"

	//request url. 此url已老，换为1.1
	//fullUrl := "https://api.twitter.com/1/statuses/update.json?include_entities=true"
	//很奇怪，只能拼装到url后面作为参数，不能放到body里面
	fullUrl := "https://api.twitter.com/1.1/statuses/update.json?status=Hello%20USA"
	//collecting parameters
	vals := make(map[string]string)
	vals["status"] = "Hello USA" //fmt.Sprintf("%s #%d", "Hello Ladies + Gentlemen, a signed OAuth request!", random)
	//vals["include_entities"] = "true"
	vals["oauth_consumer_key"] = oauth_consumer_key
	vals["oauth_nonce"] = oauth_nonce
	vals["oauth_signature_method"] = "HMAC-SHA1"
	vals["oauth_timestamp"] = oauth_timestamp
	vals["oauth_token"] = oauth_token
	vals["oauth_version"] = "1.0"

	//https://dev.twitter.com/oauth/overview/creating-signatures
	//now we get the paramter string, used to generate signatures
	paramString := EncodeParams(vals)

	//we construct the signatures base string  HTTP method + base url + paramString
	signature_base_string := fmt.Sprintf("%s&%s&%s", httpMethod, UrlEncoded(BaseUrl(fullUrl)), UrlEncoded(paramString))

	signature := Signature(consumerSecret, tokenSecret, signature_base_string)
	vals["oauth_signature"] = signature

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
		authHeaderBuf.WriteString(UrlEncoded(vals[k]))
		authHeaderBuf.WriteString("\"")
		i++
	}

	authHeader := authHeaderBuf.String()
	//formParams := fmt.Sprintf("status=%s", UrlEncoded(vals["status"]))
	req, err := http.NewRequest("POST", fullUrl, nil) //strings.NewReader(formParams))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.OpenFile("D:\\log2.txt", os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	/*
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(formParams)))
			for k := range req.Header {
				t.Logf("%s: %s", k, req.Header.Get(k))
			}
			t.Log(formParams)
	*/
	req.Write(f)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("resp header", resp.Header)
	defer resp.Body.Close()
	f.WriteString("\n------------------------------------------\n")
	f.WriteString(fmt.Sprintf("%v %v\r\n", resp.Proto /* resp.ProtoMajor, resp.ProtoMinor,*/, resp.Status))
	resp.Header.Write(f)
	var bodyReader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "deflate":
		if bodyReader, err = zlib.NewReader(resp.Body); err != nil {
			t.Fatal("zlib create failed!", err)
		}
	case "gzip":
		if bodyReader, err = gzip.NewReader(resp.Body); err != nil {
			t.Fatal("gzip create failed!", err)
		}
	}
	f.WriteString("\r\n")
	io.Copy(f, bodyReader)
}

//TestNonce
func TestNonce(t *testing.T) {

	t.Logf("nonce is %s", Nonce())
	t.Logf("nonce is %s", Nonce())
	u := "http://www.baidu.com/a/b.json?a=b&c=d"
	uBase := u
	i := strings.Index(u, "?")
	if i != -1 {
		uBase = u[:i]
	}
	t.Log(uBase)

}

//TestSignature
func TestSignature(t *testing.T) {

	baseString := "POST&https%3A%2F%2Fapi.twitter.com%2F1%2Fstatuses%2Fupdate.json&include_entities%3Dtrue%26oauth_consumer_key%3Dxvz1evFS4wEEPTGEFPHBog%26oauth_nonce%3DkYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1318622958%26oauth_token%3D370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb%26oauth_version%3D1.0%26status%3DHello%2520Ladies%2520%252B%2520Gentlemen%252C%2520a%2520signed%2520OAuth%2520request%2521"
	s1 := "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw"
	s2 := "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE"
	t.Log(Signature(s1, s2, baseString))

	//测试获取oauth_token
	var method = "POST"
	var URL = "https://api.twitter.com/oauth/request_token"
	var params = make(map[string]string)
	params["oauth_consumer_key"] = "WqMhGBZr7MkjqPiN1qIyqOKUq"
	params["oauth_nonce"] = Nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	params["oauth_token"] = "414627346-1ctEs7wQ96aB2B20Wm8CPWFRcfsfSrGNDvof3Y0H"
	params["oauth_version"] = "1.0"
	//oauth_callback can be omitted
	//params["oauth_callback"] = "https://www.davidwang.site"
	params["x_auth_access_type"] = "read"
	//生成head string
	headString := EncodeParams(params)
	//组装signature base string
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, UrlEncoded(BaseUrl(URL)), UrlEncoded(headString))

	//生成签名
	//params["oauth_signature"] = SignatureJustOneSecret("XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", signatureBaseString)
	params["oauth_signature"] = Signature("XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", "5HdPyfbJQS07ay0Sj2v9qGCeRkpuhfZcOnc0P2WdtPGpw", signatureBaseString)

	//Make Authorization Header
	authHeaderBuf := bytes.NewBufferString("OAuth ")
	i := 0
	for _, k := range HeaderKeys {

		if strings.Index(k, "oauth_") != 0 {
			continue
		}
		if params[k] == "" {
			continue
		}

		if i > 0 {
			authHeaderBuf.WriteString(", ")
		}

		authHeaderBuf.WriteString(k)
		authHeaderBuf.WriteString("=")
		authHeaderBuf.WriteString("\"")
		authHeaderBuf.WriteString(UrlEncoded(params[k]))
		authHeaderBuf.WriteString("\"")
		i++
	}

	authHeader := authHeaderBuf.String()
	t.Logf("curl -v --proxy 127.0.0.1:9099 -XPOST -d 'x_auth_access_type=read' -H 'Authorization: %s' %s\n", authHeader, URL)
}

func TestChange4AccessToken(t *testing.T) {
	//测试获取oauth_token
	var method = "POST"
	var URL = "https://api.twitter.com/oauth/access_token"
	var params = make(map[string]string)
	params["oauth_consumer_key"] = "WqMhGBZr7MkjqPiN1qIyqOKUq"
	params["oauth_nonce"] = Nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	params["oauth_token"] = "LApWqgAAAAAA6TL7AAABY6GDvpY"
	params["oauth_version"] = "1.0"
	params["oauth_verifier"] = "60pHhP0IrRkRvHj4xJ4GpiG1Z6nCwaQh"
	//生成head string
	headString := EncodeParams(params)
	//组装signature base string
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, UrlEncoded(BaseUrl(URL)), UrlEncoded(headString))

	//生成签名
	//params["oauth_signature"] = SignatureJustOneSecret("XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", signatureBaseString)
	params["oauth_signature"] = Signature("XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", "Suj9o5Ry20noo2CIC04zNaeRjRYSuK2L", signatureBaseString)

	//Make Authorization Header
	authHeaderBuf := bytes.NewBufferString("OAuth ")
	i := 0
	for _, k := range HeaderKeys {

		if strings.Index(k, "oauth_") != 0 {
			continue
		}
		if params[k] == "" {
			continue
		}

		if i > 0 {
			authHeaderBuf.WriteString(", ")
		}

		authHeaderBuf.WriteString(k)
		authHeaderBuf.WriteString("=")
		authHeaderBuf.WriteString("\"")
		authHeaderBuf.WriteString(UrlEncoded(params[k]))
		authHeaderBuf.WriteString("\"")
		i++
	}

	authHeader := authHeaderBuf.String()
	t.Logf("curl -v --proxy 127.0.0.1:9099 -XPOST -d 'oauth_verifier=60pHhP0IrRkRvHj4xJ4GpiG1Z6nCwaQh' -H 'Authorization: %s' %s\n", authHeader, URL)
}

func TestChange4BearerToken(t *testing.T) {
	//测试获取oauth_token
	var method = "POST"
	var URL = "https://api.twitter.com/oauth2/token"
	var params = make(map[string]string)
	params["oauth_consumer_key"] = "WqMhGBZr7MkjqPiN1qIyqOKUq"
	params["oauth_nonce"] = Nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	params["oauth_token"] = "1000704938408726529-sfi246O9dBr2pfLBeGXBTggb6qYd9X"
	params["oauth_version"] = "1.0"
	params["grant_type"] = "client_credentials"
	//生成head string
	headString := EncodeParams(params)
	//组装signature base string
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, UrlEncoded(BaseUrl(URL)), UrlEncoded(headString))

	//生成签名
	params["oauth_signature"] = Signature("XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", "yqlfbgLzcKp3Z2oqh4gQ5mbChaXYZNSRJWTXlbTK9apRd", signatureBaseString)

	//Make Authorization Header
	authHeaderBuf := bytes.NewBufferString("OAuth ")
	i := 0
	for _, k := range HeaderKeys {

		if strings.Index(k, "oauth_") != 0 {
			continue
		}
		if params[k] == "" {
			continue
		}

		if i > 0 {
			authHeaderBuf.WriteString(", ")
		}

		authHeaderBuf.WriteString(k)
		authHeaderBuf.WriteString("=")
		authHeaderBuf.WriteString("\"")
		authHeaderBuf.WriteString(UrlEncoded(params[k]))
		authHeaderBuf.WriteString("\"")
		i++
	}

	authHeader := authHeaderBuf.String()
	t.Logf("curl -v --proxy 127.0.0.1:9099 -XPOST -d 'grant_type=client_credentials' -H 'Authorization: %s' %s\n", authHeader, URL)
}
func TestEscape(t *testing.T) {
	//under chrome
	line := "/NDp318/FCUVgekgGMe +xYXemZs0="
	t.Log(UrlEncoded(line))
	t.Log(url.QueryEscape(line))
}

func TestEscapeMysqlInvalidString(t *testing.T) {
	line := `RT @wu_yi_fan: 170206 Tencent:中国
#KrisWu is invited to the 59th Grammy Awards 2/12 as special guest (representing China)🌟

Livestrea…  `
	t.Log(EscapeMysqlInvalidString(line))
}

/*
#################debug log record################

POST /1/statuses/update.json?include_entities=true HTTP/1.1
Host: api.twitter.com
User-Agent: Go-http-client/1.1
Content-Length: 95
Authorization: OAuth oauth_nonce="HDFyEWA73KJVRQ8ad/5ex+6WJ+haKpJlzlpCdF54otI=", oauth_signature_method="HMAC-SHA1", oauth_version="1.0", oauth_signature="QTboWDnAHwxkdRXwn+O31bmu4Mc=", oauth_consumer_key="W29UcpkImGKIT31LhRuYwOzJJ", oauth_timestamp="58", oauth_token="414627346-y2t6XrmamYi21NYwupQzH5KTUhJZhXx486zhyNPx"
Content-Type: application/x-www-form-urlencoded

status=Hello%20Ladies%20+%20Gentlemen,%20a%20signed%20OAuth%20request%21%20#5577006791947779410
------------------------------------------
HTTP/2.0 410 Gone
Content-Encoding: deflate
Content-Length: 144
Content-Type: application/json;charset=utf-8
Date: Sat, 04 Feb 2017 06:31:08 GMT
Server: tsa_a
Set-Cookie: guest_id=v1%3A148618986888257746; Domain=.twitter.com; Path=/; Expires=Mon, 04-Feb-2019 06:31:08 UTC
Strict-Transport-Security: max-age=631138519
X-Connection-Hash: 4155a03d19064ffc84b8d43ef727d31b
X-Response-Time: 3

{"errors":[{"message":"The Twitter REST API v1 is no longer active. Please migrate to API v1.1. https://dev.twitter.com/docs/api/1.1/overview.","code":64}]}
*/
