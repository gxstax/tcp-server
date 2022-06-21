package main

import (
	"fmt"
	"github.com/gxstax/tcp-server/frame"
	"github.com/gxstax/tcp-server/packet"
	"github.com/lucasepe/codename"
	"net"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var num int = 5

	wg.Add(5)

	for i := 0; i < num; i++ {
		go func(i int) {
			defer wg.Done()
			startClient(i)
		}(i + 1)
	}
	wg.Wait()
}

func startClient(i int) {
	quit := make(chan struct{})
	done := make(chan struct{})
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer conn.Close()
	fmt.Printf("[client %d]: dial ok\n", i)

	// 生成 payload
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}

	frameCodec := frame.NewMyFrameCodec()
	var counter int

	go func() {
		// handle ack
		for {
			select {
			case <-quit:
				done <- struct{}{}
				return
			default:
				// 避免阻塞在 <-quit
			}

			conn.SetReadDeadline(time.Now().Add(time.Second * 1))
			ackFramePayLoad, err := frameCodec.Decode(conn) // 解码 []byte -> FramePayLoad
			if err != nil {
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
						continue
					}
				}
				panic(err)
			}

			p, err := packet.Decode(ackFramePayLoad) // 解码: []byte -> Packet
			submitAck, ok := p.(*packet.SubmitAck)
			if !ok {
				panic("not submitAck")
			}
			fmt.Printf("[client %d]: the result of submit ack[%s] is %d\n", i, submitAck.ID, submitAck.Result)
		}
	}()

	for {
		// send submit
		counter++
		id := fmt.Sprintf("%08d", counter)
		payload := codename.Generate(rng, 4)
		s := &packet.Submit{
			ID:			id,
			Payload: 	[]byte(payload),
		}

		framePayload, err := packet.Encode(s) // 编码: Packet -> []byte
		if err != nil {
			panic(err)
		}

		fmt.Printf("[client %d]: send submit id = %s, payload=%s, frame length = %d\n", i, s.ID, s.Payload, len(framePayload) + 4)

		err = frameCodec.Encode(conn, framePayload) // 编码: FramePayload -> []byte
		if err != nil {
			panic(err)
		}

		time.Sleep( 1 * time.Second)
		if counter >= 10 {
			quit <- struct{}{}
			<-done
			fmt.Printf("[client %d]: exit ok\n", i)
			return
		}
	}
}
