package SHOUTCLOUD

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type SHOUTREQUEST struct {
	INPUT string
}

type SHOUTRESPONSE struct {
	INPUT  string
	OUTPUT string
}

func UPCASE(THING_TO_YELL string) (string, error) {
	REQUEST := &SHOUTREQUEST{THING_TO_YELL}
	ENCODED, ERR := json.Marshal(REQUEST)
	if ERR != nil {
		return "", errors.New("COULDN'T MARSHALL THE REQUEST")
	}
	READER := bytes.NewReader(ENCODED)

	// NO TLS, SO MUCH SADNESS.
	RESP, ERR := http.Post("http://API.SHOUTCLOUD.IO/V1/SHOUT",
		"application/json", READER)
	if ERR != nil {
		return "", errors.New("REQUEST FAILED CAN'T UPCASE ERROR MESSAGE HALP")
	}

	BODYBYTES, ERR := ioutil.ReadAll(RESP.Body)
	if ERR != nil {
		return "", errors.New("COULDN'T READ BODY HALP")
	}

	var BODY SHOUTRESPONSE
	if json.Unmarshal(BODYBYTES, &BODY) != nil {
		return "", errors.New("COULDN'T UNPACK RESPONSE")
	}

	return BODY.OUTPUT, nil
}
