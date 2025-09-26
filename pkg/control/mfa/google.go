package mfa

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"time"
)

type googleAuth struct {
	Secret       string
	ExpireSecond uint64
	Digits       int
}

func (g *googleAuth) Totp() string {
	count := uint64(time.Now().Unix()) / g.ExpireSecond
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(g.Secret)
	if err != nil {
		fmt.Printf("generate key failed: %v\n", err)
		return ""
	}
	codeInt := g.totp(key, count, g.Digits)
	intFormat := fmt.Sprintf("%%0%dd", g.Digits)
	code := fmt.Sprintf(intFormat, codeInt)
	fmt.Printf("code: %s\n", code)
	return code
}
func (g *googleAuth) Qr(label, issuer string) string {
	issuer = url.QueryEscape(label)
	// 规范文档 https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	// otpauth://totp/ACME%20Co:john.doe@email.com?secret=HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ&issuer=ACME%20Co&algorithm=SHA1&digits=6&period=30
	return fmt.Sprintf(`otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d`, label, g.Secret, issuer, g.Digits, g.ExpireSecond)
}

func (g *googleAuth) totp(key []byte, count uint64, digits int) int {
	hash := hmac.New(sha1.New, key)
	binary.Write(hash, binary.BigEndian, count)
	sum := hash.Sum(nil)
	// 取sha1的最后4byte
	// 0x7FFFFFFF 是long int的最大值
	// math.MaxUint32 == 2^32-1
	// & 0x7FFFFFFF == 2^31  Set the first bit of truncatedHash to zero  //remove the most significant bit
	// len(sum)-1]&0x0F 最后 像登陆 (bytes.len-4)
	// 取sha1 bytes的最后4byte 转换成 uint32
	v := binary.BigEndian.Uint32(sum[sum[len(sum)-1]&0x0F:]) & 0x7FFFFFFF
	d := uint32(1)
	// 取十进制的余数
	for i := 0; i < digits && i < 8; i++ {
		d *= 10
	}
	return int(v % d)
}

// GoogleProvider Google认证提供商实现
// @struct GoogleProvider
// @implements Provider
type GoogleProvider struct {
	issuer string // 颁发者名称
}

// GenerateSecret 生成Google认证器的MFA密钥
// @param userID string 用户ID
// @return string 密钥
// @return string 二维码URL
// @return error 生成过程中的错误
func (g *GoogleProvider) GenerateSecret(userID string) (string, string, error) {
	if userID == "" {
		return "", "", ErrInvalidUserID
	}
	// https://zh.mojotv.cn/go/golang-2fa
	// // 生成TOTP密钥
	// key, err := totp.Generate(totp.GenerateOpts{
	// 	Issuer:      g.issuer,
	// 	AccountName: userID,
	// 	Period:      30,
	// 	SecretSize:  20,
	// 	Algorithm:   otp.AlgorithmSHA1,
	// 	Digits:      otp.DigitsSix,
	// })
	// if err != nil {
	// 	g.mfaControl.logger.Errorf("[control] failed to generate Google TOTP key: %v", err)
	// 	return "", "", ErrGenerateSecretFailed
	// }
	//
	// // 存储密钥
	// g.mfaControl.secretStore[userID] = key.Secret()
	// g.mfaControl.logger.Debugf("[control] Google MFA secret generated for user: %s", userID)
	//
	// // 返回密钥和二维码URL
	// return key.Secret(), key.URL(), nil
	return "", "", nil

}

// VerifyCode 验证Google认证器的MFA代码
// @param userID string 用户ID
// @param code string MFA代码
// @return bool 验证结果
func (g *GoogleProvider) VerifyCode(secret string, code string) bool {

	// // 验证代码
	// valid, err := totp.Validate(code, secret)
	// if err != nil {
	// 	g.mfaControl.logger.Errorf("[control] failed to validate Google MFA code: %v", err)
	// 	return false
	// }
	//
	// g.mfaControl.logger.Debugf("[control] Google MFA code validation for user: %s, valid: %v", userID, valid)
	// return valid
	return false
}
