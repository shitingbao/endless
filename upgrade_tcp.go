// 这里对于长连接
// 总思路，开启新进程，继承老进程的tcp服务
// 老进程等待所有连接关闭后退出
// 新的进程监听新的连接，老进程由于被继承不会继续监听，相当于把端口让出给新进程
package endless

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

var (
	// 设置一个重启的参数，用于区分正常开启还是升级
	argReload      = flag.Bool("reload", false, "listen on fd open 3 (internal use only)")
	defaultAddress = ":8080"
)

type conflag map[string]net.Conn

type EndlessTcp struct {
	address    string
	listen     net.Listener
	wg         *sync.WaitGroup // 该标识标记了父进程的退出逻辑，在进程 listen 的时候 add，并且在信号接收的地方 wait ，在连接全部断开的时候 done，这样连接全部断开的时候，就自动退出了父进程
	readLength int
	conflags   conflag
	// 注意 conflags 这个升级过后的内容，两个进程是不共用的，比如原进程连接有五个，升级过后的有10个连接，输出长度分别是5和10.而不是都是15
}

// default adress is ":8080"
func New(ads ...string) *EndlessTcp {
	e := &EndlessTcp{
		address:    defaultAddress,
		wg:         &sync.WaitGroup{},
		readLength: 256,
		conflags:   make(map[string]net.Conn),
	}
	if len(ads) > 0 {
		e.address = ads[0]
	}
	return e
}

// SetReadLength 设置每次读取的长度
func (e *EndlessTcp) SetReadLength(readLength int) {
	e.readLength = readLength
}

// EndlessTcpListen监听入口
func (e *EndlessTcp) EndlessTcpRegisterAndListen(u UpgradeRead) error {
	flag.Parse()
	add, err := net.ResolveTCPAddr("tcp4", e.address)
	if err != nil {
		return err
	}
	if *argReload {
		// 获取到cmd中的ExtraFiles内的文件信息，以它为内容启动监听
		// ExtraFiles的内容在reload方法中放入
		log.Println("EndlessTcpRegisterAndListen reload:", *argReload)
		f := os.NewFile(3, "")
		l, err := net.FileListener(f)
		if err != nil {
			return err
		}
		e.listen = l
	} else {
		l, err := net.ListenTCP("tcp", add)
		if err != nil {
			return err
		}
		e.listen = l
	}
	go e.listenAccept(u)
	e.signalHandler()
	return nil
}

// 注意不能使用代理的情况连接，可能会出现RemoteAddr相同的情况，导致con连接对象覆盖
func (e *EndlessTcp) listenAccept(u UpgradeRead) {
	log.Println("start listen ", e.address)
	for {
		con, err := e.listen.Accept()
		if err != nil {
			log.Println("Accept:", err)
			return
		}
		e.conflags[con.RemoteAddr().String()] = con
		e.wg.Add(1)
		e.handle(con, u)
	}
}

// read write 方法待定
func (e *EndlessTcp) handle(con net.Conn, u UpgradeRead) {
	go e.read(con, u)
	// go e.write(con)
}

func (e *EndlessTcp) read(con net.Conn, u UpgradeRead) {
	for {
		result := make([]byte, e.readLength)
		n, err := con.Read(result)
		if err != nil {
			e.wg.Done()
			delete(e.conflags, con.RemoteAddr().String())
			log.Println("断开 con，当前：", len(e.conflags))
			return
		}
		u.ReadMessage(&ReadMes{
			N:   n,
			Mes: result,
		})
	}
}

func (e *EndlessTcp) Write(v interface{}) (int, error) {
	for _, con := range e.GetCons() {
		b, err := json.Marshal(v)
		if err != nil {
			return -1, err
		}
		n, err := con.Write(b)
		if err != nil {
			return n, err
		}
	}
	return 0, nil
}

func (e *EndlessTcp) GetCons() conflag {
	return e.conflags
}

// 信号处理
func (e *EndlessTcp) signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(ch)
			log.Printf("stop listen")
			return
		case syscall.SIGUSR2:
			if err := e.reload(); err != nil {
				log.Fatalf("restart error: %v", err)
			}
			// go e.reload()
			log.Println("start wait")
			e.wg.Wait()
			log.Println("stop wait")
			return
		}
	}
}

// 重启方法，这里放入进程中的输入，输出和错误
// 以及最重要的listen对象（放入ExtraFiles中），以文件句柄的形式继承
// 相当于有了所有父进程的属性，然后重新执行该可执行文件
func (e *EndlessTcp) reload() error {
	defer e.listen.Close()
	// 待定 可能需要 close 父进程的 listen 不然父进程和子进程一起接受连接，如果父进程一直有接受连接就无法直接退出
	// 但是 close 时，子进程如果还没开始监听，就会丢失连接
	// 已经连接不会受影响
	f, err := e.listen.(*net.TCPListener).File()
	if err != nil {
		log.Println("reload", err)
		return err
	}
	cmd := exec.Command(os.Args[0], "-reload")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.ExtraFiles = append(cmd.ExtraFiles, f)
	return cmd.Start() // 注意这里要用 Start，不能 Run，不然就在新进程中卡住，因为执行可执行程序 Run 中会 wait 等待
}
