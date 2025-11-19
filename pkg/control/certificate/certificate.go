package certificate

import (
	"math/big"
	"time"
)

// Certificate 证书信息
type Certificate struct {
	Algorithm string    // 算法类型 RSA/ECC/SM2
	Subject   string    // 主题信息
	Issuer    string    // 颁发者
	NotBefore time.Time // 有效期开始时间
	NotAfter  time.Time // 有效期结束时间
	Serial    *big.Int  // 序列号
}
