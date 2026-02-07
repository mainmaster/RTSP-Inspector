package main

import (
	"fmt"
	"rtsp-inspector/clients/rtsp"
)

func main() {
	baseRTSP := "rtsp://admin:qwerty123@192.168.31.176:554/RVi/1/1"
	c, err := rtsp.NewClient(baseRTSP)
	if err != nil {
		panic(err)
	}
	res, err := c.Options()
	fmt.Println(res, err)
	res, err = c.Describe()
	fmt.Println(res, err)
}
