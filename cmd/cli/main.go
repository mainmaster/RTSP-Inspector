package main

import (
	"fmt"
	"net/url"
	"rtsp-inspector/clients/rtsp"
	"rtsp-inspector/types"
	"time"

	"github.com/pixelbender/go-sdp/sdp"
)

func main() {
	baseRTSP := "rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1"

	c := rtsp.Client{}
	u, _ := url.Parse(baseRTSP)
	err := c.Connect(*u)
	if err != nil {
		panic(err)
	}
	/*
		c.SetCredentials(rtsp.Credentials{
			Username: "admin",
			Password: "qwerty123",
		})
	*/
	req, err := c.NewRequest("OPTIONS", baseRTSP)
	//req.Header["kek"] = "zzz"
	if err != nil {
		panic(err)
	}

	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	req, err = c.NewRequest("DESCRIBE", baseRTSP)
	if err != nil {
		panic(err)
	}

	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}
	sdpSess, err := sdp.ParseString(string(res.Body))
	fmt.Println(sdpSess)

	req, err = c.NewRequest("SETUP", baseRTSP+"/trackID=0")
	if err != nil {
		panic(err)
	}

	req.Header.Add("Transport", "RTP/AVP/TCP;interleaved=0-1")
	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	c.SetSessionID(res.Header.Get("Session"))

	req, err = c.NewRequest("PLAY", baseRTSP)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Session", c.GetSessionID())
	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}

	channels := types.DataChannels{
		VideoCh: make(chan []byte, 100),
		AudioCh: make(chan []byte, 100),
		RTCPCh:  make(chan []byte, 10),
		ErrCh:   make(chan error, 1),
	}

	go c.ProcessStream(channels)

	for {
		select {
		case audioPack := <-channels.AudioCh:
			fmt.Printf("Audio: %d байт\n", len(audioPack))
		case videoPack := <-channels.VideoCh:
			fmt.Printf("Видео: %d байт\n", len(videoPack))
		case rtcp := <-channels.RTCPCh:
			fmt.Printf("Служебный RTCP: %d байт\n", len(rtcp))
		case err := <-channels.ErrCh:
			fmt.Printf("Ошибка: %v\n", err)
			return // Выходим при ошибке
		case <-time.After(time.Second * 5):
			fmt.Println("Тишина в эфире более 5 секунд...")
		}
	}

}

// sdpSess, err := sdp.ParseString(string(res.Body))
