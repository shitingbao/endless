# endless_tcp

可进行无缝升级的 tcp 连接  
基本实现过程解释：  
1.正常启动 server  
2.收到重启信号后，将 server 的 listen 对象转化为 file 对象，放入 cmd 的 ExtraFiles 属性中（同时将输入，输出，err 等信息也带上）  
3.使用 cmd 继承进程命令，重新执行项目文件（附带上述属性内容），并获取 cmd 中的 ExtraFiles 中的 file 内容，将其转为 server 服务对象，而不是新开一个（这时候，是开了一个子进程，包含父进程以前信息，会继续接收新连接，而不关闭以前的连接，因为是进程继承，而不是重启），等待旧进程连接都关闭后，关闭旧进程即可。

## 前言

偶然发现虽然有无缝升级的 http 代码，但是没有 tcp 连接处理，心血来潮就写了一个，有发现问题，欢迎指正和提出！！

## Example 介绍，详细代码请看 example 文件下的 server 和 client 方法

#### 只要三步

1.定义一个实现了 ReadMessage(b \*ReadMes) 的方法

```go
type tcpModel struct {
}

func (t *tcpModel) ReadMessage(b *tcp.ReadMes) {
	log.Println(b.N, ":", string(b.Mes))
}
```

2.new 一个 EndlessTcp 对象，并将上述对象放入，开始监听

```go
	e := tcp.New(":8080")
	t := &tcpModel{}
	if err := e.EndlessTcpRegisterAndListen(t); err != nil {
		return
	}
```

3.修改代码并发送信号（以 test 为例子）

```
$ go build //修改代码后，重新构建出 test
$ ps -ef|grep test // 找到正在运行的 test 对应进程号，比如是进程号是： 1234
$ kill -SIGUSR2 1234 // 向 1234 进程发送一个 SIGUSR2 信号，如果有连接，就会显示如下：
  501 31021 27220   0  3:09下午 ttys001    0:00.09 ./test
  501 31088 31021   0  3:10下午 ttys001    0:00.05 ./test -reload

第一个是第一次执行的进程，第二个带有 -reload 参数是升级后的进程，这时候，新的连接就会被新进程接受，
旧进程将不会接受新连接，不过会继续为还没有断开的连接提供服务，直到所有旧连接都断开，然后结束旧连接。如下，只剩一个：

 501 31088     1   0  3:10下午 ??         0:00.96 ./test -reload
```

#### Example 使用方法

1.启动 server  
2.执行 client  
3.修改 server 代码（比如写入的内容修改）  
4.执行上述升级操作  
5.新连接一个 client（这时候发现两个连接，接受到的内容是不一样的）

---

## 注意事项以及问题

1.由于内部结构使用了 map 来保存 con，key 使用的是远程 ip，所有当使用代理时（比如 nginx 代理），注意 ip 重复可能导致 con 被覆盖

2.当多个连接的时候，还是需要自己处理一下特定消息的发送和接受，自己处理包的内容（write 方法和 read 方法，现在 write 方法是对所有连接进行发送消息）

3.不要改变监听的端口号，这会出现意外的错误

4.当频繁升级时，可能会滞留一些老连接，这时候需要注意升级的次数和当前的连接

5.对于 os.NewFile 的第一个参数为什么是 ‘ 3 ’ ，  
0，1，2，分别代表了，标准输入，标准输出，和标准错误（就是控制台的输入，显示，还有程序的报错），所以手动打开的就是从‘3’开始，linux 系统都是遵循从小到大的原则，就是打开文件，就从小到大分配这个值，具体解释请参考：https://blog.csdn.net/u011244839/article/details/72865493

## 额外的扩展

grpc 也适用于这个方法，因为他也是基于 tcp 的方式来监听的，说到底也是一个 tcp 的 listen ，所以也可以这么无缝重启
