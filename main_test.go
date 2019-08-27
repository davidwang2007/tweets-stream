package main

import "testing"
import "time"

func TestTime(t *testing.T) {
	value := "Mon Sep 24 03:35:21 +0000 2012"
	// Writing down the way the standard time would look like formatted our way
	layout := time.RubyDate
	/*
		value := "Thu, 05/19/11, 10:47PM"

		// Writing down the way the standard time would look like formatted our way
		layout := "Mon, 01/02/06, 03:04PM"
	*/
	if ti, err := time.Parse(layout, value); err != nil {
		t.Fatal(err)
	} else {
		t.Log(ti)
	}

}
