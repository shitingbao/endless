package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8080")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	go func() {
		for {
			s := "hello"
			_, err = conn.Write([]byte(s))
			checkError(err)
			time.Sleep(time.Second * 2)
		}
	}()
	go func() {
		for {
			result := make([]byte, 256)
			_, err = conn.Read(result)
			checkError(err)
			mes := ""
			json.Unmarshal(result, &mes)
			log.Println("read:", mes)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Printf("stop")
			signal.Stop(ch)
			log.Printf("graceful shutdown")
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
