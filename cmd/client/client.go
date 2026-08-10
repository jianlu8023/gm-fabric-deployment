package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/jianlu8023/go-tools/v2/pkg/http"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/json/sonic"
	"github.com/jianlu8023/go-tools/v2/pkg/pidfile"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jianlu8023/golang-example/pkg/control/job"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/server"
	"github.com/jianlu8023/golang-example/version"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tjfoc/gmsm/gmtls"
	gmx509 "github.com/tjfoc/gmsm/x509"
	"math/rand/v2"
	gohttp "net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {

	serverControl, err := server.NewServerControlFromFile()
	if err != nil {
		fmt.Printf("generate server control failed: %v\n", err)
		return
	}

	mainLogger := serverControl.GetLoggerControl().GenLogger("main")

	mainLogger.Infof("start server version %v", version.Version)

	// pidfile
	{
		runOS := runtime.GOOS
		switch runOS {
		case "windows":
			mainLogger.Infof("running on Windows")
		case "linux":
			mainLogger.Infof("running on Linux")
			if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
				mainLogger.Errorf("generate pid file failed: %v", err)
				return
			}
			defer func() {
				pidfile.ReleasePID()
			}()
		case "darwin": // macOS
			mainLogger.Infof("running on MacOS")
			if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
				mainLogger.Errorf("generate pid file failed: %v", err)
				return
			}
			defer func() {
				pidfile.ReleasePID()
			}()
		default:
			mainLogger.Warnf("running on an unknown operating system: %s", runOS)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverControl.StartUp(func(err error) {
		mainLogger.Errorf("start server failed: %v", err)
		quit <- os.Interrupt
	})
	defer func(serverControl *server.Control) {
		if err := serverControl.Shutdown(); err != nil {
			mainLogger.Errorf("shutdown server failed: %v", err)
		}
	}(serverControl)

	{

		if serverControl.GetJobControl() != nil {
			serverControl.GetJobControl().RegisterJob(&job.Job{
				Name: "grpc-ping-message",
				Task: func() {
					response, err := serverControl.GetGrpcControl().Call(&pb.BaseRequest{
						MessageType: libp2p.MsgBasePing,
					})
					if err != nil {
						mainLogger.Errorf("send ping message err: %v", err)
					} else {
						mainLogger.Debugf("send ping message success %v", response)
					}
				},
				Interval: time.Second * time.Duration(rand.IntN(5-3)+3),
			})

			serverControl.GetJobControl().RegisterJob(&job.Job{
				Name:     "libp2p-ping-message",
				Interval: time.Duration(rand.IntN(10-5)+5) * time.Second,
				Task: func() {
					pingMsg := &libp2p.Message{
						Type:    libp2p.MsgBasePing,
						Content: []byte("ping"),
						From:    serverControl.GetLibp2pControl().GetLocalhostPeerID(),
					}
					if err := serverControl.GetLibp2pControl().BroadcastMessage(pingMsg); err != nil {
						mainLogger.Errorf("broadcast ping message failed: %v", err)
					}
				},
			})

			serverControl.GetJobControl().RegisterJob(&job.Job{
				Name:     "http-request",
				Interval: time.Duration(rand.IntN(10-5)+5) * time.Second,
				Task: func() {

					var client *http.Client
					if serverControl.GetConfigControl().GetWebConfig().TlsGM {
						certPool := gmx509.NewCertPool()
						rootPem, err := os.ReadFile("certs/tongsuo/subca.crt")
						if err != nil {
							mainLogger.Errorf("read root.crt err: %v", err)
							return
						}
						if ok := certPool.AppendCertsFromPEM(rootPem); !ok {
							mainLogger.Errorf("append root.crt err")
							return
						}

						keyPair1, err := gmtls.LoadX509KeyPair("certs/tongsuo/client_sign.crt", "certs/tongsuo/client_sign.key")
						if err != nil {
							mainLogger.Errorf("load key pair err: %v", err)
							return
						}

						keyPair2, err := gmtls.LoadX509KeyPair("certs/tongsuo/client_enc.crt", "certs/tongsuo/client_enc.key")
						if err != nil {
							mainLogger.Errorf("load key pair err: %v", err)
							return
						}
						tlsConfig := &gmtls.Config{
							GMSupport: &gmtls.GMSupport{
								WorkMode: gmtls.ModeGMSSLOnly,
							},
							Certificates:       []gmtls.Certificate{keyPair1, keyPair2},
							RootCAs:            certPool,
							InsecureSkipVerify: false,
							ServerName:         "grpc",
						}
						client = http.NewClientWithGMTls(tlsConfig)
					} else {
						certPool := x509.NewCertPool()
						rootPem, err := os.ReadFile("certs/openssl/root.crt")
						if err != nil {
							mainLogger.Errorf("read root.crt err: %v", err)
							return
						}
						if ok := certPool.AppendCertsFromPEM(rootPem); !ok {
							mainLogger.Errorf("append root.crt err")
							return
						}

						keyPair, err := tls.LoadX509KeyPair("certs/openssl/hclient-chain.crt", "certs/openssl/hclient.key")
						if err != nil {
							mainLogger.Errorf("load key pair err: %v", err)
							return
						}

						client = http.NewClient().SetTLSClientConfig(&tls.Config{
							InsecureSkipVerify: false,
							RootCAs:            certPool,
							Certificates:       []tls.Certificate{keyPair},
						})
					}

					if client == nil {
						return
					}

					var objJson commonhttp.BaseResponse
					code, err := client.
						GetJSON(
							"https://127.0.0.1:8080/example/endpoints",
							map[string]interface{}{},
							&objJson,
						)
					if err != nil {
						mainLogger.Errorf("get router failed: %v", err)
						return
					}
					if code != gohttp.StatusOK {
						if code == gohttp.StatusTooManyRequests {
							mainLogger.Errorf("get router failed code: %v body: %v", code, objJson)
						} else {
							mainLogger.Errorf("get router failed: %v", code)
						}
					}
					pretty, err := sonic.NewStandardSonic().MarshalString(objJson)
					if err != nil {
						mainLogger.Errorf("marshal json failed: %v", err)
						return
					}
					mainLogger.Infof("response: %v ", pretty)

					time.Sleep(time.Duration(rand.IntN(1)) * time.Second)
					body, code, err := client.GET(
						"https://127.0.0.1:8080/example/ping",
						map[string]interface{}{},
					)
					if err != nil {
						mainLogger.Errorf("get ping failed: %v", err)
						return
					}
					if code != gohttp.StatusOK {
						if code == gohttp.StatusTooManyRequests {
							mainLogger.Errorf("get ping failed code: %v body: %v", code, string(body))
						} else {
							mainLogger.Errorf("get ping failed: %v", code)
						}
					}
					mainLogger.Debugf("response: %v ", string(body))

					time.Sleep(time.Duration(rand.IntN(1)) * time.Second)
					body, code, err = client.GET(
						"https://127.0.0.1:8080/example/libp2p/list",
						map[string]interface{}{
							"isPage":   true,
							"pageNo":   1,
							"pageSize": 10,
						},
					)
					if err != nil {
						mainLogger.Errorf("get libp2p list failed: %v", err)
						return
					}
					if code != gohttp.StatusOK {
						if code == gohttp.StatusTooManyRequests {
							mainLogger.Errorf("get libp2p list failed code: %v body: %v", code, string(body))
						} else {
							mainLogger.Errorf("get libp2p list failed: %v", code)
						}
					}
					mainLogger.Debugf("response: %v ", string(body))

					time.Sleep(time.Duration(rand.IntN(1)) * time.Second)
					body, code, err = client.POST(
						"https://127.0.0.1:8080/example/ping",
						nil,
					)
					if err != nil {
						mainLogger.Errorf("post ping failed: %v", err)
						return
					}
					if code != gohttp.StatusOK {
						if code == gohttp.StatusTooManyRequests {
							mainLogger.Errorf("post ping failed code: %v body: %v", code, string(body))
						} else {
							mainLogger.Errorf("post ping failed: %v", code)
						}
					}
					mainLogger.Debugf("response: %v ", string(body))
				},
			})
		}
	}

	// libp2p
	{
		serverControl.GetLibp2pControl().RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %v messageType from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionNode, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Infof("received %v protocol %v message from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
			nodeInfoMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Content: []byte("success"),
				Type:    libp2p.MsgLibp2pNode,
			}
			if err := serverControl.GetLibp2pControl().SendMessageToPeer(nodeInfoMsg.To, nodeInfoMsg); err != nil {
				mainLogger.Errorf("send node info to peer %v failed: %v", nodeInfoMsg.To, err)
			}
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionDockerNetworks, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %v message from %v", protocolID, msg.Type, msg.From)
			if serverControl.GetDockerControl() == nil {
				return
			}
			listNetworks, err := serverControl.GetDockerControl().ListNetworks()
			if err != nil {
				mainLogger.Errorf("list docker networks failed: %v", err)
				return
			}
			bytes, err := json.Marshal(listNetworks)
			if err != nil {
				mainLogger.Errorf("marshal docker networks failed: %v", err)
				return
			}

			dockerNetworkMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Content: bytes,
				Type:    libp2p.MsgDockerNetworks,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerNetworkMsg.To, dockerNetworkMsg); err != nil {
				mainLogger.Errorf("send docker networks to peer %v failed: %v", dockerNetworkMsg.To, err)
			}
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionDockerImages, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %s message from %v", protocolID, msg.Type, msg.From)

			if serverControl.GetDockerControl() == nil {
				return
			}

			listImages, err := serverControl.GetDockerControl().ListImages()
			if err != nil {
				mainLogger.Errorf("list docker images failed: %v", err)
				return
			}
			content, err := json.Marshal(listImages)
			if err != nil {
				mainLogger.Errorf("marshal docker images failed: %v", err)
				return
			}

			dockerImageMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Type:    libp2p.MsgDockerImages,
				Content: content,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerImageMsg.To, dockerImageMsg); err != nil {
				mainLogger.Errorf("send docker images to peer %v failed: %v", dockerImageMsg.To, err)
			}

		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgDockerImagePull, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %s message from %v", protocolID, msg.Type, msg.From)

			if serverControl.GetDockerControl() == nil {
				return
			}

			var dockerPullContent libp2p.DockerImagePullContent
			if err := json.Unmarshal(msg.Content, &dockerPullContent); err != nil {
				mainLogger.Errorf("unmarshal docker image pull content failed: %v", err)
				return
			}

			if err := serverControl.GetDockerControl().PullImage(dockerPullContent.ImageName,
				docker.WithImagePullPlatform(dockerPullContent.Platform),
				docker.WithImagePullRegistryAuth(dockerPullContent.RegistryAuth),
			); err != nil {
				mainLogger.Errorf("pull docker image failed: %v", err)
				return
			}

			listImages, err := serverControl.GetDockerControl().ListImages()
			if err != nil {
				mainLogger.Errorf("list docker images failed: %v", err)
				return
			}
			content, err := json.Marshal(listImages)
			if err != nil {
				mainLogger.Errorf("marshal docker images failed: %v", err)
				return
			}

			dockerImageMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Type:    libp2p.MsgDockerImages,
				Content: content,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerImageMsg.To, dockerImageMsg); err != nil {
				mainLogger.Errorf("send docker images to peer %v failed: %v", dockerImageMsg.To, err)
			}

		})

	}

	<-quit
	mainLogger.Infof("received shutdown signal...")

	time.Sleep(5 * time.Second)
}
