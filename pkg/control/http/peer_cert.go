package http

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	gmx509 "github.com/tjfoc/gmsm/x509"
)

// stdlibPeerCertLogger 生成标准 crypto/tls 的 VerifyPeerCertificate 回调。
//
// @description 标准库在完成对端证书链校验后调用本回调（服务端对应客户端证书 mTLS）。
// 链的可信性已由 crypto/tls 保证，这里只做审计留痕：记录对端叶子证书的主题、签发者、
// 证书链长度与到期时间，并在证书即将过期时输出告警。回调始终返回 nil，不干预握手结果。
// @return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error 校验回调
func (c *Control) stdlibPeerCertLogger() func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
			// 对端未出示证书（如单向 TLS 或 VerifyClientCertIfGiven 下客户端未带证书）
			return nil
		}
		leaf := verifiedChains[0][0]
		remaining := time.Until(leaf.NotAfter)
		c.logger.Infof("[http/control] mTLS 对端证书校验通过: subject=%s issuer=%s chain_len=%d not_after=%s remaining=%s",
			leaf.Subject.CommonName, leaf.Issuer.CommonName, len(verifiedChains[0]),
			leaf.NotAfter.Format(time.RFC3339), remaining.Round(time.Hour).String())
		if remaining > 0 && remaining < peerCertExpiringSoonWarn {
			c.logger.Warnf("[http/control] 对端证书即将过期: subject=%s not_after=%s remaining=%s",
				leaf.Subject.CommonName, leaf.NotAfter.Format(time.RFC3339), remaining.Round(time.Hour).String())
		} else if remaining <= 0 {
			c.logger.Warnf("[http/control] 对端证书已过期: subject=%s not_after=%s",
				leaf.Subject.CommonName, leaf.NotAfter.Format(time.RFC3339))
		}
		return nil
	}
}

// gmPeerCertLogger 生成国密 gmtls 的 VerifyPeerCertificate 回调。
//
// @description 国密 TLS 在完成对端证书链校验后调用本回调，verifiedChains 中的证书为
// github.com/tjfoc/gmsm/x509 的国密证书类型。仅做审计留痕，不干预握手结果。
// @return func(rawCerts [][]byte, verifiedChains [][]*gmx509.Certificate) error 校验回调
func (c *Control) gmPeerCertLogger() func(rawCerts [][]byte, verifiedChains [][]*gmx509.Certificate) error {
	return func(_ [][]byte, verifiedChains [][]*gmx509.Certificate) error {
		if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
			return nil
		}
		leaf := verifiedChains[0][0]
		remaining := time.Until(leaf.NotAfter)
		c.logger.Infof("[http/control] GM mTLS 对端证书校验通过: subject=%s issuer=%s chain_len=%d not_after=%s remaining=%s",
			leaf.Subject.CommonName, leaf.Issuer.CommonName, len(verifiedChains[0]),
			leaf.NotAfter.Format(time.RFC3339), remaining.Round(time.Hour).String())
		if remaining > 0 && remaining < peerCertExpiringSoonWarn {
			c.logger.Warnf("[http/control] GM 对端证书即将过期: subject=%s not_after=%s remaining=%s",
				leaf.Subject.CommonName, leaf.NotAfter.Format(time.RFC3339), remaining.Round(time.Hour).String())
		} else if remaining <= 0 {
			c.logger.Warnf("[http/control] GM 对端证书已过期: subject=%s not_after=%s",
				leaf.Subject.CommonName, leaf.NotAfter.Format(time.RFC3339))
		}
		return nil
	}
}

// logStdlibServerCert 打印标准TLS服务端自身证书信息并做临期告警
//
// @description 服务端证书在启动阶段一次性加载并预解析 Leaf，本方法打印证书主体、SAN 与到期时间；
// 剩余有效期低于 peerCertExpiringSoonWarn（含已过期）时输出 WARN，
// 避免证书过期导致客户端握手失败却无从排查
// @param leaf *x509.Certificate 服务端 leaf 证书
// @param certFile string 证书文件路径，用于日志定位
func (c *Control) logStdlibServerCert(leaf *x509.Certificate, certFile string) {
	remaining := time.Until(leaf.NotAfter)
	days := int(remaining / (24 * time.Hour))
	subject := leaf.Subject.CommonName
	if subject == "" {
		subject = leaf.Subject.String()
	}
	names := make([]string, 0, len(leaf.DNSNames)+len(leaf.IPAddresses))
	names = append(names, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		names = append(names, ip.String())
	}
	msg := fmt.Sprintf("TLS服务端证书 file=%s subject=%s SAN=[%s] 有效期至 %s(剩余%d天)",
		certFile, subject, strings.Join(names, ","), leaf.NotAfter.Format("2006-01-02 15:04:05"), days)
	if remaining < peerCertExpiringSoonWarn {
		c.logger.Warnf("[http/control] %s，请及时续期", msg)
		return
	}
	c.logger.Infof("[http/control] %s", msg)
}

// logGMServerCert 打印国密TLS服务端自身证书信息并做临期告警
//
// @description 国密服务端证书在启动阶段一次性加载并预解析 Leaf，本方法打印证书主体、SAN 与到期时间；
// 剩余有效期低于 peerCertExpiringSoonWarn（含已过期）时输出 WARN
// @param leaf *gmx509.Certificate 国密服务端 leaf 证书
// @param certFile string 证书文件路径，用于日志定位
func (c *Control) logGMServerCert(leaf *gmx509.Certificate, certFile string) {
	remaining := time.Until(leaf.NotAfter)
	days := int(remaining / (24 * time.Hour))
	subject := leaf.Subject.CommonName
	if subject == "" {
		subject = leaf.Subject.String()
	}
	names := make([]string, 0, len(leaf.DNSNames)+len(leaf.IPAddresses))
	names = append(names, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		names = append(names, ip.String())
	}
	msg := fmt.Sprintf("GM TLS服务端证书 file=%s subject=%s SAN=[%s] 有效期至 %s(剩余%d天)",
		certFile, subject, strings.Join(names, ","), leaf.NotAfter.Format("2006-01-02 15:04:05"), days)
	if remaining < peerCertExpiringSoonWarn {
		c.logger.Warnf("[http/control] %s，请及时续期", msg)
		return
	}
	c.logger.Infof("[http/control] %s", msg)
}
