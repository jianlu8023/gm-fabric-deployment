package certificate

import (
	"math/big"
	"time"
)

// Certificate 证书信息
type Certificate struct {
	Algorithm          string    // 算法类型 RSA/ECC/SM2
	PublicKeyAlgorithm string    // 公钥算法
	SignatureAlgorithm string    // 签名算法
	Subject            string    // 主题信息
	Issuer             string    // 颁发者
	NotBefore          time.Time // 有效期开始时间
	NotAfter           time.Time // 有效期结束时间
	Serial             *big.Int  // 序列号
	Version            int       // 证书版本
	IsCA               bool      // 是否是CA证书
	DNSNames           []string  // 主题备用名称
	KeyUsage           string    // 密钥用途
	ExtKeyUsage        []string  // 扩展密钥用途
	PublicKeyLength    int       // 公钥长度（位）
	FingerprintSHA1    string    // SHA1指纹
	FingerprintSHA256  string    // SHA256指纹
}
