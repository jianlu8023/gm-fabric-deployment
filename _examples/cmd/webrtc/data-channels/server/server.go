package server

import (
	sd "_examples/pkg/webrtc/sessiondescription"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func StartServer() {
	// 1. 配置webrtc
	config := webrtc.Configuration{
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
	// 2. 创建peerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()
	// 3. 注册ice candidate 收集完成的回调
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// 4. 注册data channel
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", dataChannel.Label(), dataChannel.ID())

		dataChannel.OnOpen(func() {
			fmt.Printf("Data Channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())

			dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
				fmt.Printf("Server received: '%s'\n", string(msg.Data))
				reply := fmt.Sprintf("Server received: %s", string(msg.Data))
				if err := dataChannel.SendText(reply); err != nil {
					log.Println("Error sending reply:", err)
				}
			})
		})
	})

	// 5. 等待offer 从控制台输入
	fmt.Println("Please paste the offer from the client:")
	offer := webrtc.SessionDescription{}
	sd.Decode(readUntilNewline(), &offer)

	// 6. 设置 RemoteDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Fatal(err)
	}

	// 7. 创建 Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// 8. 设置 LocalDescription
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Fatal(err)
	}

	// 9. 等待 ICE Candidate 收集完成
	<-gatherComplete

	// 10. 打印 Answer 到控制台
	fmt.Println("Please copy this answer and paste it to the client:")
	fmt.Println(sd.Encode(peerConnection.LocalDescription()))

	// 阻塞主线程
	select {}

	// dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// dataChannel.OnOpen(func() {
	// 	fmt.Println("Data channel opened!")
	//
	// 	// 收到客户端消息后回复
	// 	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 		fmt.Printf("Server received: '%s'\n", string(msg.Data))
	// 		reply := fmt.Sprintf("Server received: %s", string(msg.Data))
	// 		err := dataChannel.SendText(reply)
	// 		if err != nil {
	// 			log.Println("Error sending reply:", err)
	// 		}
	// 	})
	// })
	// offer, err := peerConnection.CreateOffer(nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = peerConnection.SetLocalDescription(offer)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// <-gatherComplete
	//
	// sdpHttp(peerConnection)

}

func sdpHttp(peerConnection *webrtc.PeerConnection) {
	http.HandleFunc("/offer", func(w http.ResponseWriter, r *http.Request) {
		offer := peerConnection.LocalDescription()
		json.NewEncoder(w).Encode(offer)
	})

	http.HandleFunc("/answer", func(w http.ResponseWriter, r *http.Request) {
		var answer webrtc.SessionDescription
		err := json.NewDecoder(r.Body).Decode(&answer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = peerConnection.SetRemoteDescription(answer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("Answer received and set successfully.")
	})

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
