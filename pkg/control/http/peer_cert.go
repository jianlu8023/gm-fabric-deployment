package http

import (
	"crypto/x509"
	"time"

	gmx509 "github.com/tjfoc/gmsm/x509"
)

// peerCertExpiringSoonWarn 对端证书剩余有效期小于该阈值时输出告警，便于运维提前续期。
const peerCertExpiringSoonWarn = 30 * 24 * time.Hour

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
