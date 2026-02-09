package main

import (
	"fmt"
	"rtsp-inspector/clients/rtsp"
)

func main() {
	baseRTSP := "rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1"
	c := rtsp.Client{}

	req, err := rtsp.NewRequest("OPTIONS", baseRTSP)
	//req.Header["kek"] = "zzz"
	req.Credentials.Username = "admin"
	req.Credentials.Password = "qwerty123"

	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

// sdpSess, err := sdp.ParseString(string(res.Body))
