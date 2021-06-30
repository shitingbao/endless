# endless_tcp
可进行无缝升级的tcp连接

## 前言
偶然发现虽然有无缝升级的http代码，但是没有tcp连接处理，心血来潮就写了一个，有发现问题，欢迎指正和提出！！

## Example 介绍，详细代码请看example文件下的server和client方法
#### 只要三步  
1.定义一个实现了 ReadMessage(b *ReadMes)  的方法
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

3.修改代码并发送信号（以test为例子）
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
#### Example使用方法
1.启动server  
2.执行client  
3.修改server代码（比如写入的内容修改）  
4.执行上述升级操作  
5.新连接一个client（这时候发现两个连接，接受到的内容是不一样的）  

----------------------
## 注意事项以及问题

1.由于内部结构使用了 map 来保存con，key使用的是远程ip，所有当使用代理时（比如nginx代理），注意ip重复可能导致con被覆盖

2.当多个连接的时候，还是需要自己处理一下特定消息的发送和接受，自己处理包的内容（write方法和read方法，现在write方法是对所有连接进行发送消息）

3.不要改变监听的端口号，这会出现意外的错误

4.当频繁升级时，可能会滞留一些老连接，这时候需要注意升级的次数和当前的连接

5.对于os.NewFile 的第一个参数为什么是 ‘ 3 ’ ，  
0，1，2，分别代表了，标准输入，标准输出，和标准错误（就是控制台的输入，显示，还有程序的报错），所以手动打开的就是从‘3’开始，linux系统都是遵循从小到大的原则，就是打开文件，就从小到大分配这个值，具体解释请参考：https://blog.csdn.net/u011244839/article/details/72865493