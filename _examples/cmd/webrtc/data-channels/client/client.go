package client

import (
	sd "_examples/pkg/webrtc/sessiondescription"
	"bufio"
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func StartClient() {
	// 1.  配置 WebRTC
	config := webrtc.Configuration{
		ICETransportPolicy: webrtc.ICETransportPolicyAll,

		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"}, // 作为备选
			},
			{
				URLs:       []string{"turn:192.168.58.110:3478?transport=udp", "turn:192.168.58.110:3478?transport=tcp"},
				Username:   "abc",
				Credential: "abcpw",
			},
		},
	}
	// 2. 创建 PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()
	// 3. 注册 ICE Candidate 收集完成的回调
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// 4. 创建 DataChannel
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Fatal(err)
	}
	// 5. 注册 DataChannel 的 OnOpen 回调
	dataChannel.OnOpen(func() {
		fmt.Println("Data channel opened!")

		// 定时发送消息
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			message := "Hello from client!"
			fmt.Printf("Client sending: '%s'\n", message)
			err := dataChannel.SendText(message)
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	})
	// 6. 注册 DataChannel 的 OnMessage 回调
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Client received: '%s'\n", string(msg.Data))
	})
	// 7. 创建 Offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}
	// 8. 设置 LocalDescription
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		log.Fatal(err)
	}
	// 9. 等待 ICE Candidate 收集完成
	<-gatherComplete

	// 10. 打印 Offer 到控制台
	fmt.Println("Please copy this offer and paste it to the server:")
	fmt.Println(sd.Encode(peerConnection.LocalDescription()))

	// 11. 等待 Answer 从控制台输入
	fmt.Println("Please paste the answer from the server:")
	answer := webrtc.SessionDescription{}
	sd.Decode(readUntilNewline(), &answer)

	// 12. 设置 RemoteDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		log.Fatal(err)
	}

	// 阻塞主线程
	select {}

	// var dataChannel *webrtc.DataChannel
	// peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
	// 	dataChannel = dc
	// 	dataChannel.OnOpen(func() {
	// 		fmt.Println("Data channel opened!")
	//
	// 		// 定时发送消息
	// 		ticker := time.NewTicker(3 * time.Second)
	// 		defer ticker.Stop()
	//
	// 		for range ticker.C {
	// 			message := "Hello from client!"
	// 			fmt.Printf("Client sending: '%s'\n", message)
	// 			err := dataChannel.SendText(message)
	// 			if err != nil {
	// 				log.Println("Error sending message:", err)
	// 				return
	// 			}
	// 		}
	// 	})
	//
	// 	// 注册 DataChannel 的 OnMessage 回调
	// 	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 		fmt.Printf("Client received: '%s'\n", string(msg.Data))
	// 	})
	// })
	//
	// resp, err := http.Get("http://localhost:8080/offer")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer resp.Body.Close()
	//
	// offer := webrtc.SessionDescription{}
	// err = json.NewDecoder(resp.Body).Decode(&offer)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // 6. 设置 RemoteDescription
	// err = peerConnection.SetRemoteDescription(offer)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // 7. 创建 Answer
	// answer, err := peerConnection.CreateAnswer(nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // 8. 设置 LocalDescription
	// err = peerConnection.SetLocalDescription(answer)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // 9. 等待 ICE Candidate 收集完成
	// <-gatherComplete
	//
	// // 10. 将 Answer 发送给 Server
	// answerBytes, err := json.Marshal(peerConnection.LocalDescription())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// _, err = http.Post("http://localhost:8080/answer", "application/json", bytes.NewBuffer(answerBytes))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Answer sent to server successfully.")
	//
	// // 阻塞主线程
	// select {}
}

// Read from stdin until we get a newline.
func readUntilNewline() (in string) {
	var err error

	r := bufio.NewReader(os.Stdin)
	for {
		in, err = r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}

		if in = strings.TrimSpace(in); len(in) > 0 {
			break
		}
	}

	fmt.Println("")

	return
}
