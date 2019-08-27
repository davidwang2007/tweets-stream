package main

import "testing"
import "encoding/json"

//TestRest test rest api
func TestRest(t *testing.T) {
	params := map[string]string{
		"status": "Hello David Wang",
	}

	if result, err := DoPost("https://api.twitter.com/1.1/statuses/update.json", params); err != nil {
		t.Fatal(err)
	} else {
		t.Log(result)
	}

}

func TestDoGet(t *testing.T) {
	//https://api.twitter.com/1.1/search/tweets.json
	params := map[string]string{
		"q":                "trump",
		"count":            "100",
		"include_entities": "false",
	}

	if result, err := DoGet("https://api.twitter.com/1.1/search/tweets.json", params); err != nil {
		t.Fatal(err)
	} else {
		var inf interface{} = result["statuses"]
		for i, v := range inf.([]interface{}) {
			m := v.(map[string]interface{})
			t.Logf("%d: %v", i+1, m["text"])
		}
	}

}

func TestTrimChunk(t *testing.T) {
	arr := []byte{' ', '\t', 1, 2, 3, 4, '\n', '\r'}
	t.Log(TrimChunk(arr))
}

func TestJson(t *testing.T) {
	line := "[{\"a\":1}]"
	v := make([]interface{}, 1)
	json.Unmarshal([]byte(line), &v)
	t.Log(v)

}
