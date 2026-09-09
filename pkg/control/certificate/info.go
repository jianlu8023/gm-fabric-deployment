package certificate

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gmsm2 "github.com/tjfoc/gmsm/sm2"
	gmx509 "github.com/tjfoc/gmsm/x509"
)

// certFileExts 认定为证书文件的扩展名。
var certFileExts = map[string]bool{
	".crt": true,
	".cer": true,
	".pem": true,
	".der": true,
}

// keyFileExts 认定为私钥文件的扩展名。
var keyFileExts = map[string]bool{
	".key": true,
}

// CertFileInfo 证书目录下的文件摘要信息
//
// @description 用于 /certs/list 接口返回证书目录中的文件元数据（不含私钥内容）
// @struct
type CertFileInfo struct {
	Name    string    `json:"name"`              // 文件名
	Path    string    `json:"path"`              // 文件绝对路径
	Size    int64     `json:"size"`              // 文件大小（字节）
	ModTime time.Time `json:"mod_time"`          // 最后修改时间
	Kind    string    `json:"kind"`              // 文件类型：cert（证书）| key（私钥）| other
	IsCA    bool      `json:"is_ca,omitempty"`   // 证书文件是否为 CA 证书（仅 kind=cert 时尝试判定）
	Subject string    `json:"subject,omitempty"` // 证书主题（仅 kind=cert 且可解析时填充）
}

// ParseCertificateFile 解析证书文件并返回证书信息
//
// @description 读取 PEM 格式证书文件，优先按标准 x509 解析，失败后再尝试按国密 x509 解析；
// 解析成功返回填充完整的 Certificate 信息。该函数为包级函数，不依赖 Control 实例，
// 便于 web 层在证书控制器未启用时复用。
// @param certPath string 证书文件路径
// @return *Certificate 证书信息
// @return error 读取或解析失败时的错误信息
func ParseCertificateFile(certPath string) (*Certificate, error) {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate")
	}

	var certInfo *Certificate

	// 尝试解析为标准x509证书
	if cert, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
		certInfo = &Certificate{
			Subject:     cert.Subject.String(),
			Issuer:      cert.Issuer.String(),
			NotBefore:   cert.NotBefore,
			NotAfter:    cert.NotAfter,
			Serial:      cert.SerialNumber,
			Version:     cert.Version,
			IsCA:        cert.IsCA,
			DNSNames:    cert.DNSNames,
			KeyUsage:    formatKeyUsage(cert.KeyUsage),
			ExtKeyUsage: formatExtKeyUsage(cert.ExtKeyUsage),
		}

		// 设置公钥算法和长度
		certInfo.PublicKeyAlgorithm = cert.PublicKeyAlgorithm.String()
		switch pubKey := cert.PublicKey.(type) {
		case *rsa.PublicKey:
			certInfo.Algorithm = "RSA"
			certInfo.PublicKeyLength = pubKey.N.BitLen()
		case *ecdsa.PublicKey:
			certInfo.Algorithm = "ECC"
			certInfo.PublicKeyLength = pubKey.Curve.Params().BitSize
		}

		// 设置签名算法
		certInfo.SignatureAlgorithm = cert.SignatureAlgorithm.String()

		// 计算指纹
		certInfo.FingerprintSHA1 = calculateFingerprintSHA1(cert.Raw)
		certInfo.FingerprintSHA256 = calculateFingerprintSHA256(cert.Raw)
	} else {
		// 尝试解析为国密证书
		if cert, err := gmx509.ParseCertificate(certBlock.Bytes); err == nil {
			certInfo = &Certificate{
				Algorithm:   "SM2",
				Subject:     cert.Subject.String(),
				Issuer:      cert.Issuer.String(),
				NotBefore:   cert.NotBefore,
				NotAfter:    cert.NotAfter,
				Serial:      cert.SerialNumber,
				Version:     cert.Version,
				IsCA:        cert.IsCA,
				DNSNames:    cert.DNSNames,
				KeyUsage:    formatGMKeyUsage(cert.KeyUsage),
				ExtKeyUsage: formatGMExtKeyUsage(cert.ExtKeyUsage),
			}

			// 设置公钥算法和长度
			certInfo.PublicKeyAlgorithm = formatGMPublicKeyAlgorithm(cert.PublicKeyAlgorithm)
			if _, ok := cert.PublicKey.(*gmsm2.PublicKey); ok {
				// SM2公钥长度通常是256位
				certInfo.PublicKeyLength = 256
			}

			// 设置签名算法
			certInfo.SignatureAlgorithm = cert.SignatureAlgorithm.String()

			// 计算指纹
			certInfo.FingerprintSHA1 = calculateFingerprintSHA1(cert.Raw)
			certInfo.FingerprintSHA256 = calculateFingerprintSHA256(cert.Raw)
		} else {
			return nil, fmt.Errorf("failed to parse certificate")
		}
	}

	return certInfo, nil
}

// ListCertificateFiles 列举证书目录下的证书与私钥文件
//
// @description 遍历指定目录（不递归），返回其中证书/私钥文件的元数据列表，按文件名排序；
// 对可解析的证书文件额外填充 IsCA 与 Subject 字段，私钥文件不解析内容以避免泄露敏感信息。
// @param dir string 证书目录路径
// @return []CertFileInfo 文件摘要列表
// @return error 目录不存在或读取失败时的错误信息
func ListCertificateFiles(dir string) ([]CertFileInfo, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("cert dir is empty")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert dir %s: %w", dir, err)
	}

	files := make([]CertFileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		kind := "other"
		if certFileExts[ext] {
			kind = "cert"
		} else if keyFileExts[ext] {
			kind = "key"
		} else {
			// 仅收集证书与私钥文件，跳过其他文件
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		fullPath := filepath.Join(dir, name)
		item := CertFileInfo{
			Name:    name,
			Path:    fullPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Kind:    kind,
		}
		// 对证书文件尝试解析，补充 CA 标记与主题（解析失败不影响列表返回）
		if kind == "cert" {
			if cert, err := ParseCertificateFile(fullPath); err == nil {
				item.IsCA = cert.IsCA
				item.Subject = cert.Subject
			}
		}
		files = append(files, item)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return files, nil
}
