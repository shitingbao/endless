package main

import (
	"log"
	"time"

	tcp "github.com/shitingbao/endless_tcp"
)

type tcpModel struct {
}

func (t *tcpModel) ReadMessage(b *tcp.ReadMes) {
	log.Println(b.N, ":", string(b.Mes))
}
func main() {
	e := tcp.New("")
	t := &tcpModel{}
	go func() {
		for {
			e.Write("aaaaaa")
			time.Sleep(time.Second * 2)
		}
	}()
	if err := e.EndlessTcpRegisterAndListen(t); err != nil {
		log.Println(err)
	}
}
