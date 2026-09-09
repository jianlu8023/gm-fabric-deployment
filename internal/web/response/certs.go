package response

import (
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// PeerCertInfo 对端（客户端）证书信息
//
// @description 描述 mTLS 握手中对端出示的叶子证书关键信息
// @struct
type PeerCertInfo struct {
	Subject           string    `json:"subject"`             // 证书主题
	Issuer            string    `json:"issuer"`              // 签发者
	NotBefore         time.Time `json:"not_before"`          // 有效期开始时间
	NotAfter          time.Time `json:"not_after"`           // 有效期结束时间
	DNSNames          []string  `json:"dns_names,omitempty"` // 主题备用名称
	Serial            string    `json:"serial,omitempty"`    // 序列号
	FingerprintSHA256 string    `json:"fingerprint_sha256"`  // SHA256指纹
}

// CertsTlsResponse 当前TLS连接信息响应
//
// @description 返回当前 HTTPS 连接的 TLS/mTLS 协商结果，包括协议版本、加密套件、ALPN 与对端证书
// @struct
type CertsTlsResponse struct {
	TlsEnabled        bool           `json:"tls_enabled"`                // 服务端是否启用TLS
	GmTls             bool           `json:"gm_tls"`                     // 是否为国密TLS
	Version           string         `json:"version,omitempty"`          // 协商的TLS版本
	CipherSuite       string         `json:"cipher_suite,omitempty"`     // 协商的加密套件
	NegotiatedProto   string         `json:"negotiated_proto,omitempty"` // ALPN 协商协议（h2/http/1.1）
	ServerName        string         `json:"server_name,omitempty"`      // SNI 服务器名
	HandshakeComplete bool           `json:"handshake_complete"`         // 握手是否完成
	ClientCertPresent bool           `json:"client_cert_present"`        // 对端是否出示客户端证书
	PeerCerts         []PeerCertInfo `json:"peer_certs,omitempty"`       // 对端证书链（叶子证书在前）
	Note              string         `json:"note,omitempty"`             // 补充说明（如国密连接下标准库无法读取连接状态）
}

// String 将 CertsTlsResponse 转换为字符串
//
// @description 将TLS连接信息响应序列化为 JSON 字符串
// @return string JSON 格式字符串
func (resp CertsTlsResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}
