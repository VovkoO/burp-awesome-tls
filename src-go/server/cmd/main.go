package main

import "C"

import (
	"flag"
	"fmt"
	"log"

	"server"
)

func main() {
	interceptAddr := flag.String("intercept", server.DefaultInterceptProxyAddress, "Intercept proxy address to listen on ([ip:]port)")
	burpAddr := flag.String("burp", server.DefaultBurpProxyAddress, "Burp proxy address to listen on ([ip:]port)")
	spoofAddr := flag.String("spoof", server.DefaultSpoofProxyAddress, "Spoof proxy address to listen on ([ip:]port)")
	flag.Parse()

	if err := server.StartProxy(*interceptAddr, *burpAddr); err != nil {
		log.Fatalln(err)
	}

	log.Fatalln(server.StartServer(*spoofAddr))
}

//export StartServer
func StartServer(spoofAddr *C.char) *C.char {
	if err := server.StartServer(C.GoString(spoofAddr)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export StopServer
func StopServer() *C.char {
	if err := server.StopServer(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export SaveSettings
func SaveSettings(configJson *C.char) *C.char {
	if err := server.SaveSettings(C.GoString(configJson)); err != nil {
		return C.CString(err.Error())
	}

	return C.CString("")
}

//export SmokeTest
func SmokeTest() {
	fmt.Println("smoke test success")
}
