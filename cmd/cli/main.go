package main

import (
	"fmt"
	"rtsp-inspector/clients/rtsp"

	"github.com/pixelbender/go-sdp/sdp"
)

func main() {
	baseRTSP := "rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1"

	c := rtsp.Client{}
	/*
		c.SetCredentials(rtsp.Credentials{
			Username: "admin",
			Password: "qwerty123",
		})
	*/
	req, err := rtsp.NewRequest("OPTIONS", baseRTSP)
	//req.Header["kek"] = "zzz"
	if err != nil {
		panic(err)
	}

	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	req, err = rtsp.NewRequest("DESCRIBE", baseRTSP)
	if err != nil {
		panic(err)
	}

	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}
	sdpSess, err := sdp.ParseString(string(res.Body))
	fmt.Println(sdpSess)

	req, err = rtsp.NewRequest("SETUP", baseRTSP+"/trackID=0")
	if err != nil {
		panic(err)
	}

	req.Header.Add("Transport", "RTP/AVP/TCP;interleaved=0-1")
	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

}

// sdpSess, err := sdp.ParseString(string(res.Body))
