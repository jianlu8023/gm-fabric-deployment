package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// 取消这个credential

type customCredential struct{}

func (c *customCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer from context")
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("failed to get tls info from peer")
	}
	if len(tlsInfo.State.PeerCertificates) > 0 {
		cert := tlsInfo.State.PeerCertificates[0]
		subject := cert.Subject.String()
		fmt.Printf("客户端证书 Subject: %s\n", subject)
		// 可以在这里将客户端证书信息添加到 metadata 中
		return map[string]string{
			"client-subject": subject,
		}, nil
	}
	return nil, fmt.Errorf("没有客户端证书")
}

func (c *customCredential) RequireTransportSecurity() bool {
	return false
}
