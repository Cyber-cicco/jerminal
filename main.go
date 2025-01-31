package main

import "github.com/Cyber-cicco/jerminal/server"

func main() {
	s := server.New(8002)
    s.TestListen()
}
