# golimiter

Go Limiter 令牌限速桶 上下行速度计算

缺点：开始令牌桶是满的，速度可能会比较大，限速精度存在上下波动

## 使用示例

服务端限速 512KB/s(1024 * 512)

客户端测试发送 10M 数据(1024 * 1024 * 10)

完成时间 20s

```go
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/wenjy/golimiter"
	"github.com/wenjy/helper"
)

type server struct {
	up   *golimiter.Limiter
	seed *golimiter.Speed
}

var s1 *server

// 服务端限速 512KB/s
// 客户端测试发送 10M 数据 1024 * 1024 * 10
// 完成时间 20s
func main() {

	s1 = &server{
		up:   golimiter.NewLimiter(1024 * 512),
		seed: golimiter.NewSpeed(),
	}

	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:8083")
		if err != nil {
			// handle error
			fmt.Println(err)
			return
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
				fmt.Println(err)
				return
			}
			go handleConnection(conn)
		}
	}()

	conn, err := net.Dial("tcp", "127.0.0.1:8083")

	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}

	buf := make([]byte, 1024*1024) // 1MB
	count := 0
	num := 10
	start := time.Now().Unix()
	for i := 0; i < num; i++ {
		n, err := conn.Write(buf)
		if err != nil {
			// handle error
			fmt.Println(err)
			return
		}
		count += n
		fmt.Println("client count", count, time.Now().Unix()-start)
	}

	for {

	}
}

func handleConnection(c net.Conn) {
	var count int
	buf := make([]byte, 2048)
	start := time.Now().Unix()
	for {
		s1.up.WaitToken(len(buf))
		rn, err := c.Read(buf)
		s1.seed.IncrUpBytes(rn)
		count += rn
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("服务端读取[%v]字节 使用时间[%d]s", count, time.Now().Unix()-start)
		u, _ := s1.seed.UpDownSpeed()
		fmt.Println("上行速度", helper.HumanizeSize(uint64(u)))
	}
}

```

输出：`服务端读取[10485760]字节 使用时间[19]s上行速度 514 KiB`