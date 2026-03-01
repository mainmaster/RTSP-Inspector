package main

import (
	"context"
	"fmt"
	"net/url"
	"rtsp-inspector/internal/rtsp_client"
	"rtsp-inspector/types"
	"time"
)

func main() {
	baseRTSP := "rtsp_client://admin:qwerty123@192.168.31.176:554/RVi/1/1"

	c := rtsp_client.Client{}
	u, _ := url.Parse(baseRTSP)
	err := c.Connect(*u)
	if err != nil {
		panic(err)
	}

	req, err := c.NewRequest("OPTIONS", baseRTSP)
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

	for _, t := range res.GetTrackIDs() {
		req, err = c.NewRequest("SETUP", baseRTSP)
		req.SetTrackID(t)

		if err != nil {
			panic(err)
		}

		res, err = c.Do(req)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}

	req, err = c.NewRequest("PLAY", baseRTSP)
	if err != nil {
		panic(err)
	}

	res, err = c.Do(req)
	if err != nil {
		panic(err)
	}

	channels := types.DataChannels{
		VideoCh:     make(chan []byte, 100),
		AudioCh:     make(chan []byte, 100),
		RTCPAudioCh: make(chan []byte, 10),
		RTCPVideoCh: make(chan []byte, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.ProcessStream(ctx, channels)

	for {
		select {
		case audioPack := <-channels.AudioCh:
			fmt.Printf("Audio: %d байт\n", len(audioPack))
		case rtcp := <-channels.RTCPAudioCh:
			fmt.Printf("Служебный RTCP: %d байт\n", len(rtcp))
		case videoPack := <-channels.VideoCh:
			fmt.Printf("Видео: %d байт\n", len(videoPack))
		case rtcp := <-channels.RTCPVideoCh:
			fmt.Printf("Служебный RTCP: %d байт\n", len(rtcp))
		case <-time.After(time.Second * 5):
			fmt.Println("Тишина в эфире более 5 секунд...")
			ctx.Done()
			break
		}
	}

}
