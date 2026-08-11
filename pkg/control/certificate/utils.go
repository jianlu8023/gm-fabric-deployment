package certificate

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	gmsm2 "github.com/tjfoc/gmsm/sm2"
	gmx509 "github.com/tjfoc/gmsm/x509"
	"go.uber.org/zap"
)

// parseSubject 解析主题信息
// @param subject string 主题信息，使用/分割(/OU=/O=/CN等)
// @return pkix.Name 解析后的主题信息
func parseSubject(subject string) pkix.Name {
	name := pkix.Name{}
	parts := strings.Split(subject, "/")
	for _, part := range parts {
		if stringer.IsBlank(part) {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]
		switch key {
		case "CN":
			name.CommonName = value
		case "O":
			name.Organization = append(name.Organization, value)
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, value)
		case "C":
			name.Country = append(name.Country, value)
		case "ST":
			name.Province = append(name.Province, value)
		case "L":
			name.Locality = append(name.Locality, value)
		}
	}
	return name
}

// generateSerialNumber 生成序列号
// @return *big.Int 序列号
func generateSerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
}

// parseSANs 解析SAN信息
// @param sans string 逗号分隔的SAN列表
// @return []string 解析后的SAN列表
func parseSANs(sans string) []string {
	var dnsNames []string
	if !stringer.IsBlank(sans) {
		sanList := strings.Split(sans, ",")
		for _, san := range sanList {
			san = strings.TrimSpace(san)
			if !stringer.IsBlank(san) {
				dnsNames = append(dnsNames, san)
			}
		}
	}
	return dnsNames
}

// saveCertificate 保存证书到文件
// @param certPath string 证书保存路径
// @param name string 证书名称
// @param derBytes []byte 证书DER数据
// @param logger *zap.SugaredLogger 日志对象
// @return string 保存的证书文件路径
// @return error 错误信息
func saveCertificate(certPath, name string, derBytes []byte, logger *zap.SugaredLogger) (string, error) {
	certFilePath := filepath.Join(certPath, name+".crt")

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create certificate file: %w", err)
	}

	defer func() {
		if err := certOut.Close(); err != nil {
			logger.Errorf("[certificate/control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", fmt.Errorf("failed to write certificate: %w", err)
	}

	return certFilePath, nil
}

// saveRSAPrivateKey 保存RSA私钥到文件
// @param certPath string 私钥保存路径
// @param name string 私钥名称
// @param privateKey *rsa.PrivateKey RSA私钥
// @param logger *zap.SugaredLogger 日志对象
// @return string 保存的私钥文件路径
// @return error 错误信息
func saveRSAPrivateKey(certPath, name string, privateKey *rsa.PrivateKey, logger *zap.SugaredLogger) (string, error) {
	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create key file: %w", err)
	}

	defer func() {
		if err := keyOut.Close(); err != nil {
			logger.Errorf("[certificate/control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return "", fmt.Errorf("failed to write private key: %w", err)
	}

	return keyFilePath, nil
}

// saveECCPrivateKey 保存ECC私钥到文件
// @param certPath string 私钥保存路径
// @param name string 私钥名称
// @param privateKey *ecdsa.PrivateKey ECC私钥
// @param logger *zap.SugaredLogger 日志对象
// @return string 保存的私钥文件路径
// @return error 错误信息
func saveECCPrivateKey(certPath, name string, privateKey *ecdsa.PrivateKey, logger *zap.SugaredLogger) (string, error) {
	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create key file: %w", err)
	}

	defer func() {
		if err := keyOut.Close(); err != nil {
			logger.Errorf("[certificate/control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return "", fmt.Errorf("failed to write private key: %w", err)
	}

	return keyFilePath, nil
}

// saveSM2PrivateKey 保存SM2私钥到文件
// @param certPath string 私钥保存路径
// @param name string 私钥名称
// @param privateKey *gmsm2.PrivateKey SM2私钥
// @param logger *zap.SugaredLogger 日志对象
// @return string 保存的私钥文件路径
// @return error 错误信息
func saveSM2PrivateKey(certPath, name string, privateKey *gmsm2.PrivateKey, logger *zap.SugaredLogger) (string, error) {
	// 使用国密库的私钥序列化方法
	privateKeyBytes, err := gmx509.WritePrivateKeyToPem(privateKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create key file: %w", err)
	}

	defer func() {
		if err := keyOut.Close(); err != nil {
			logger.Errorf("[certificate/control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	// WritePrivateKeyToPem已经返回了完整的PEM格式数据，直接写入文件即可
	if _, err := keyOut.Write(privateKeyBytes); err != nil {
		return "", fmt.Errorf("failed to write private key: %w", err)
	}

	return keyFilePath, nil
}

// GenerateCertChain 生成证书链文件
// @param certPath string 证书链保存路径
// @param certChainName string 证书链名称
// @param certFilePaths ...string 变长参数，证书文件路径列表，按照顺序添加到证书链中
// @return string 保存的证书链文件路径
// @return error 错误信息
func GenerateCertChain(certPath, certChainName string, certFilePaths ...string) (string, error) {
	// 检查证书链名称是否已包含任意后缀，如果没有则添加.crt后缀
	fileName := certChainName
	// 如果文件名不包含点号（没有后缀），则添加.crt后缀
	if !strings.Contains(fileName, ".") {
		fileName += ".crt"
	}

	// 构造证书链文件路径
	chainFilePath := filepath.Join(certPath, fileName)

	// 创建证书链文件
	chainOut, err := os.Create(chainFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create certificate chain file: %w", err)
	}

	defer func() {
		if err := chainOut.Close(); err != nil {
			fmt.Printf("close certificate chain {%v} file failed: %v", certChainName, err)
		}
	}()

	// 遍历所有证书文件路径
	for _, certFilePath := range certFilePaths {
		// 检查证书文件是否存在
		fileInfo, err := os.Stat(certFilePath)
		if err != nil {
			return "", fmt.Errorf("certificate file %s does not exist: %w", certFilePath, err)
		}

		// 检查是否是文件而非目录
		if fileInfo.IsDir() {
			return "", fmt.Errorf("%s is a directory, not a certificate file", certFilePath)
		}

		// 读取证书文件内容
		certData, err := os.ReadFile(certFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read certificate file %s: %w", certFilePath, err)
		}

		// 直接将证书内容写入证书链文件
		if _, err := chainOut.Write(certData); err != nil {
			return "", fmt.Errorf("failed to write certificate content to chain file: %w", err)
		}
	}

	return chainFilePath, nil
}
