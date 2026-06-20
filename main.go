package main

import (
	"log"
	"runtime"
)

const (
	typeData     = 0x00
	typeRegister = 0x01
	typeRegReply = 0x02
	mtu          = 1400
)

func main() {
	log.SetFlags(log.Ltime)
	log.Printf("ОС: %s — режим: %s", runtime.GOOS, modeName)
	run()
}
