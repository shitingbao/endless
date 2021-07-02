package main

import (
	"log"
	"time"

	tcp "github.com/shitingbao/endless_tcp"
)

type tcpModel struct {
}

// ReadMessage 会接受到 read 中的数据，自己处理沾包等问题
func (t *tcpModel) ReadMessage(b *tcp.ReadMes) {
	log.Println(b.N, ":", string(b.Mes))
}

func main() {
	e := tcp.New()       // 默认使用 :8080 端口
	e.SetReadLength(256) // 设置每次读取长度，不设置默认使用 256 长度
	t := &tcpModel{}
	go func() { // 这里 Write 只是做一个例子，实际的使用需要自己定义逻辑，或者使用 GetCons 获取所有连接后筛选写入
		for {
			if _, err := e.Write("aaaaaa"); err != nil {
				return
			}
			time.Sleep(time.Second * 2)
		}
	}()
	if err := e.EndlessTcpRegisterAndListen(t); err != nil {
		log.Println(err)
	}
}
