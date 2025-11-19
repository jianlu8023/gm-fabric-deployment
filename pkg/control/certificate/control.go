package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	gmsm2 "github.com/tjfoc/gmsm/sm2"
	gmx509 "github.com/tjfoc/gmsm/x509"
	"go.uber.org/zap"
)

// Control 证书控制器
type Control struct {
	config *config.CertificateConfig
	logger *zap.SugaredLogger
	mutex  sync.RWMutex
	once   sync.Once
}

// NewCertificateControl 新建证书控制器
// @return *Control 证书控制器
func NewCertificateControl(certificateConfig *config.CertificateConfig, loggerControl *logger.Control) *Control {
	if certificateConfig == nil {
		certificateConfig = getDefaultConfig()
	}
	if !certificateConfig.Enabled {
		return nil
	}

	if loggerControl == nil {
		return nil
	}

	certificateLogger := loggerControl.GenLogger("[Cert]")

	control := &Control{
		config: certificateConfig,
		logger: certificateLogger,
	}

	return control
}

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

// StartUp 启动证书控制器
// @param failedFunc func(err error) 启动失败回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			// 如果启用了证书功能，则自动生成根证书
			if err := c.ensureRootCertificate(); err != nil {
				c.logger.Errorf("[control] failed to ensure root certificate: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
		}
	})
}

// Shutdown 关闭证书控制器
// @return error 关闭过程中的错误
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutdown certificate control...")
	return nil
}

// ensureRootCertificate 确保根证书存在
// @return error 错误信息
func (c *Control) ensureRootCertificate() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 检查根证书是否存在
	rootCertPath := filepath.Join(c.config.CertPath, "root.crt")
	rootKeyPath := filepath.Join(c.config.CertPath, "root.key")

	// 如果根证书和私钥都存在，则不需要重新生成
	if exists, err := path.FileExists(rootCertPath); err == nil && exists {
		if exists, err := path.FileExists(rootKeyPath); err == nil && exists {
			c.logger.Infof("[control] root certificate already exists...")
			return nil
		}
	}

	// 创建证书和私钥目录
	if _, err := path.CreateDir(c.config.CertPath); err != nil {
		return fmt.Errorf("failed to create cert path: %w", err)
	}

	// 生成根证书
	c.logger.Debugf("[control] starting generate root certificate...")
	return c.GenerateRootCertificate("root", c.config.RootSubject, c.config.DefaultAlgo, "")
}

// GenerateRootCertificate 生成根证书
// @param name string 证书名称
// @param subject string 主题信息
// @param algorithm string 算法类型 RSA/ECC/SM2
// @return error 错误信息
func (c *Control) GenerateRootCertificate(name, subject, algorithm string, certPath string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	switch strings.ToUpper(algorithm) {
	case "RSA":
		return c.generateRSARootCertificate(name, subject, certPath)
	case "ECC":
		return c.generateECCRootCertificate(name, subject, certPath)
	case "SM2":
		return c.generateSM2RootCertificate(name, subject, certPath)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// generateRSARootCertificate 生成RSA根证书
// @param name string 证书名称
// @param subject string 主题信息
// @return error 错误信息
func (c *Control) generateRSARootCertificate(name, subject string, certPath string) error {
	// 生成RSA私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		MaxPathLenZero:        false,
	}

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")
	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Debugf("[control] RSA root certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateECCRootCertificate 生成ECC根证书
// @param name string 证书名称
// @param subject string 主题信息
// @return error 错误信息
func (c *Control) generateECCRootCertificate(name, subject string, certPath string) error {
	// 生成ECC私钥 ecc 默认使用 p384曲线
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ECC private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		MaxPathLenZero:        false,
	}

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}
	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}
	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Debugf("[control] ECC root certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateSM2RootCertificate 生成SM2根证书
// @param name string 证书名称
// @param subject string 主题信息
// @return error 错误信息
func (c *Control) generateSM2RootCertificate(name, subject string, certPath string) error {
	// 生成SM2私钥
	privateKey, err := gmsm2.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate SM2 private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := gmx509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		KeyUsage:              gmx509.KeyUsageCertSign | gmx509.KeyUsageCRLSign,
		ExtKeyUsage:           []gmx509.ExtKeyUsage{gmx509.ExtKeyUsageServerAuth, gmx509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		MaxPathLenZero:        false,
	}

	// 自签名证书
	publicKey := &privateKey.PublicKey
	derBytes, err := gmx509.CreateCertificate(&template, &template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyPath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	// 使用国密库的私钥序列化方法
	privateKeyBytes, err := gmx509.WritePrivateKeyToPem(privateKey, nil)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Debugf("[Control] SM2 root certificate generated: %s, key: %s", certFilePath, keyPath)
	return nil
}

// GenerateIntermediateCertificate 生成中间证书
// @param name string 证书名称
// @param subject string 主题信息
// @param algorithm string 算法类型 RSA/ECC/SM2
// @return error 错误信息
func (c *Control) GenerateIntermediateCertificate(name, subject, algorithm string, signCertPath, signKeyPath, certPath string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if stringer.IsBlank(signCertPath) {
		signCertPath = filepath.Join(c.config.CertPath, "root.crt")
	}

	// 检查根证书是否存在
	if exists, err := path.FileExists(signCertPath); err != nil || !exists {
		return fmt.Errorf("root certificate not found: %s", signCertPath)
	}

	if stringer.IsBlank(signKeyPath) {
		signKeyPath = filepath.Join(c.config.CertPath, "root.key")
	}

	if exists, err := path.FileExists(signKeyPath); err != nil || !exists {
		return fmt.Errorf("root key not found: %s", signKeyPath)
	}

	switch strings.ToUpper(algorithm) {
	case "RSA":
		return c.generateRSAIntermediateCertificate(name, subject, signCertPath, signKeyPath, certPath)
	case "ECC":
		return c.generateECCIntermediateCertificate(name, subject, signCertPath, signKeyPath, certPath)
	case "SM2":
		return c.generateSM2IntermediateCertificate(name, subject, signCertPath, signKeyPath, certPath)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// generateRSAIntermediateCertificate 生成RSA中间证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateRSAIntermediateCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string) error {
	// 读取根证书
	rootCertBytes, err := os.ReadFile(rootCertPath)
	if err != nil {
		return fmt.Errorf("failed to read root certificate: %w", err)
	}

	rootCertBlock, _ := pem.Decode(rootCertBytes)
	if rootCertBlock == nil {
		return fmt.Errorf("failed to decode root certificate")
	}

	rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse root certificate: %w", err)
	}

	// 读取根私钥
	rootKeyBytes, err := os.ReadFile(rootKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read root key: %w", err)
	}

	rootKeyBlock, _ := pem.Decode(rootKeyBytes)
	if rootKeyBlock == nil {
		return fmt.Errorf("failed to decode root key")
	}

	var rootPrivateKey *rsa.PrivateKey
	if rootKeyBlock.Type == "RSA PRIVATE KEY" {
		rootPrivateKey, err = x509.ParsePKCS1PrivateKey(rootKeyBlock.Bytes)
	} else if rootKeyBlock.Type == "PRIVATE KEY" {
		parsedKey, err := x509.ParsePKCS8PrivateKey(rootKeyBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse root private key: %w", err)
		}
		var ok bool
		rootPrivateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("root key is not RSA private key")
		}
	} else {
		return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to parse root private key: %w", err)
	}

	// 生成中间证书私钥
	intermediateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate intermediate private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0), // 5年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
	}

	// 签发证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCert, &intermediateKey.PublicKey, rootPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create intermediate certificate: %w", err)
	}

	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(intermediateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] RSA intermediate certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateECCIntermediateCertificate 生成ECC中间证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateECCIntermediateCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string) error {
	// 读取根证书
	rootCertBytes, err := os.ReadFile(rootCertPath)
	if err != nil {
		return fmt.Errorf("failed to read root certificate: %w", err)
	}

	rootCertBlock, _ := pem.Decode(rootCertBytes)
	if rootCertBlock == nil {
		return fmt.Errorf("failed to decode root certificate")
	}

	rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse root certificate: %w", err)
	}

	// 读取根私钥
	rootKeyBytes, err := os.ReadFile(rootKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read root key: %w", err)
	}

	rootKeyBlock, _ := pem.Decode(rootKeyBytes)
	if rootKeyBlock == nil {
		return fmt.Errorf("failed to decode root key")
	}

	// 解析根私钥
	var rootPrivateKey *ecdsa.PrivateKey
	if rootKeyBlock.Type == "EC PRIVATE KEY" {
		rootPrivateKey, err = x509.ParseECPrivateKey(rootKeyBlock.Bytes)
	} else if rootKeyBlock.Type == "PRIVATE KEY" {
		parsedKey, err := x509.ParsePKCS8PrivateKey(rootKeyBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse root private key: %w", err)
		}
		var ok bool
		rootPrivateKey, ok = parsedKey.(*ecdsa.PrivateKey)
		if !ok {
			return fmt.Errorf("root key is not ECC private key")
		}
	} else {
		return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to parse root private key: %w", err)
	}

	// 生成中间证书私钥
	intermediateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate intermediate private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0), // 5年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
	}

	// 签发证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCert, &intermediateKey.PublicKey, rootPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create intermediate certificate: %w", err)
	}

	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func() {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func(keyOut *os.File) {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}(keyOut)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(intermediateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] ECC intermediate certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateSM2IntermediateCertificate 生成SM2中间证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateSM2IntermediateCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string) error {
	// 读取根证书
	rootCertBytes, err := os.ReadFile(rootCertPath)
	if err != nil {
		return fmt.Errorf("failed to read root certificate: %w", err)
	}

	rootCertBlock, _ := pem.Decode(rootCertBytes)
	if rootCertBlock == nil {
		return fmt.Errorf("failed to decode root certificate")
	}

	rootCert, err := gmx509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse root certificate: %w", err)
	}

	// 读取根私钥
	rootKeyBytes, err := os.ReadFile(rootKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read root key: %w", err)
	}

	rootKeyBlock, _ := pem.Decode(rootKeyBytes)
	if rootKeyBlock == nil {
		return fmt.Errorf("failed to decode root key")
	}

	// 解析根私钥
	var rootPrivateKey *gmsm2.PrivateKey
	if rootKeyBlock.Type == "PRIVATE KEY" {
		rootPrivateKey, err = gmx509.ReadPrivateKeyFromPem(rootKeyBytes, nil)
		if err != nil {
			return fmt.Errorf("failed to parse root private key: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
	}

	// 生成中间证书私钥
	intermediateKey, err := gmsm2.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate intermediate private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := gmx509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0), // 5年有效期
		KeyUsage:              gmx509.KeyUsageCertSign | gmx509.KeyUsageCRLSign,
		ExtKeyUsage:           []gmx509.ExtKeyUsage{gmx509.ExtKeyUsageServerAuth, gmx509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
	}

	// 签发证书
	publicKey := &intermediateKey.PublicKey
	derBytes, err := gmx509.CreateCertificate(&template, rootCert, publicKey, rootPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create intermediate certificate: %w", err)
	}

	if stringer.IsBlank(certPath) {
		c.logger.Warnf("[control] certPath is empty, use default path...")
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func(certOut *os.File) {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}(certOut)

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")
	

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func(keyOut *os.File) {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}(keyOut)

	// 使用国密库的私钥序列化方法
	intermediateKeyBytes, err := gmx509.WritePrivateKeyToPem(intermediateKey, nil)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: intermediateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] SM2 intermediate certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// GenerateLeafCertificate 生成叶子证书
// @param name string 证书名称
// @param subject string 主题信息
// @param algorithm string 算法类型 RSA/ECC/SM2
// @return error 错误信息
func (c *Control) GenerateLeafCertificate(name, subject, algorithm string, selfSign bool, signCertPath, signKeyPath, certPath string) error {

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if stringer.IsBlank(signCertPath) {
		signCertPath = filepath.Join(c.config.CertPath, "root.crt")
	}

	// 检查根证书是否存在
	if exists, err := path.FileExists(signCertPath); err != nil || !exists {
		return fmt.Errorf("root certificate not found: %s", signCertPath)
	}
	if stringer.IsBlank(signKeyPath) {
		signKeyPath = filepath.Join(c.config.CertPath, "root.key")
	}
	if exists, err := path.FileExists(signKeyPath); err != nil || !exists {
		return fmt.Errorf("root key not found: %s", signKeyPath)
	}

	switch strings.ToUpper(algorithm) {
	case "RSA":
		return c.generateRSALeafCertificate(name, subject, signCertPath, signKeyPath, certPath, selfSign)
	case "ECC":
		return c.generateECCLeafCertificate(name, subject, signCertPath, signKeyPath, certPath, selfSign)
	case "SM2":
		return c.generateSM2LeafCertificate(name, subject, signCertPath, signKeyPath, certPath, selfSign)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// generateRSALeafCertificate 生成RSA叶子证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateRSALeafCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string, selfSign bool) error {
	// 如果 selfSign 为 true，则使用自签名方式生成证书

	// 生成叶子证书私钥
	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate leaf private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	var (
		derBytes []byte
		// cert     *x509.Certificate
	)

	if selfSign {
		// 自签名证书
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &leafKey.PublicKey, leafKey)
		if err != nil {
			return fmt.Errorf("failed to create self-signed leaf certificate: %w", err)
		}
		// cert = &template
	} else {
		// 使用根证书签发
		// 读取根证书
		rootCertBytes, err := os.ReadFile(rootCertPath)
		if err != nil {
			return fmt.Errorf("failed to read root certificate: %w", err)
		}

		rootCertBlock, _ := pem.Decode(rootCertBytes)
		if rootCertBlock == nil {
			return fmt.Errorf("failed to decode root certificate")
		}

		rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse root certificate: %w", err)
		}

		// 读取根私钥
		rootKeyBytes, err := os.ReadFile(rootKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read root key: %w", err)
		}

		rootKeyBlock, _ := pem.Decode(rootKeyBytes)
		if rootKeyBlock == nil {
			return fmt.Errorf("failed to decode root key")
		}

		var rootPrivateKey *rsa.PrivateKey
		if rootKeyBlock.Type == "RSA PRIVATE KEY" {
			rootPrivateKey, err = x509.ParsePKCS1PrivateKey(rootKeyBlock.Bytes)
		} else if rootKeyBlock.Type == "PRIVATE KEY" {
			parsedKey, err := x509.ParsePKCS8PrivateKey(rootKeyBlock.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse root private key: %w", err)
			}
			var ok bool
			rootPrivateKey, ok = parsedKey.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("root key is not RSA private key")
			}
		} else {
			return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
		}

		if err != nil {
			return fmt.Errorf("failed to parse root private key: %w", err)
		}

		// 签发证书
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, rootCert, &leafKey.PublicKey, rootPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to create leaf certificate: %w", err)
		}
		// cert = rootCert
	}

	if stringer.IsBlank(certPath) {
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func(certOut *os.File) {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}(certOut)

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")
	

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func(keyOut *os.File) {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}(keyOut)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] RSA leaf certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateECCLeafCertificate 生成ECC叶子证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateECCLeafCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string, selfSign bool) error {
	// 如果 selfSign 为 true，则使用自签名方式生成证书

	// 生成叶子证书私钥
	leafKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate leaf private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	var (
		derBytes []byte
		// cert     *x509.Certificate
	)

	if selfSign {
		// 自签名证书
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &leafKey.PublicKey, leafKey)
		if err != nil {
			return fmt.Errorf("failed to create self-signed leaf certificate: %w", err)
		}
		// cert = &template
	} else {
		// 使用根证书签发
		// 读取根证书
		rootCertBytes, err := os.ReadFile(rootCertPath)
		if err != nil {
			return fmt.Errorf("failed to read root certificate: %w", err)
		}

		rootCertBlock, _ := pem.Decode(rootCertBytes)
		if rootCertBlock == nil {
			return fmt.Errorf("failed to decode root certificate")
		}

		rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse root certificate: %w", err)
		}

		// 读取根私钥
		rootKeyBytes, err := os.ReadFile(rootKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read root key: %w", err)
		}

		rootKeyBlock, _ := pem.Decode(rootKeyBytes)
		if rootKeyBlock == nil {
			return fmt.Errorf("failed to decode root key")
		}

		// 解析根私钥
		var rootPrivateKey *ecdsa.PrivateKey
		if rootKeyBlock.Type == "EC PRIVATE KEY" {
			rootPrivateKey, err = x509.ParseECPrivateKey(rootKeyBlock.Bytes)
		} else if rootKeyBlock.Type == "PRIVATE KEY" {
			parsedKey, err := x509.ParsePKCS8PrivateKey(rootKeyBlock.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse root private key: %w", err)
			}
			var ok bool
			rootPrivateKey, ok = parsedKey.(*ecdsa.PrivateKey)
			if !ok {
				return fmt.Errorf("root key is not ECC private key")
			}
		} else {
			return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
		}

		if err != nil {
			return fmt.Errorf("failed to parse root private key: %w", err)
		}

		// 签发证书
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, rootCert, &leafKey.PublicKey, rootPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to create leaf certificate: %w", err)
		}
		// cert = rootCert
	}

	if stringer.IsBlank(certPath) {
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func(certOut *os.File) {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}(certOut)

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")
	

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func(keyOut *os.File) {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}(keyOut)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] ECC leaf certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// generateSM2LeafCertificate 生成SM2叶子证书
// @param name string 证书名称
// @param subject string 主题信息
// @param rootCertPath string 根证书路径
// @param rootKeyPath string 根私钥路径
// @return error 错误信息
func (c *Control) generateSM2LeafCertificate(name, subject, rootCertPath, rootKeyPath string, certPath string, selfSign bool) error {
	// 如果 selfSign 为 true，则使用自签名方式生成证书

	// 生成叶子证书私钥
	leafKey, err := gmsm2.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate leaf private key: %w", err)
	}

	// 解析主题信息
	subjectName := parseSubject(subject)

	// 生成序列号
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// 创建证书模板
	template := gmx509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              gmx509.KeyUsageDigitalSignature | gmx509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []gmx509.ExtKeyUsage{gmx509.ExtKeyUsageServerAuth, gmx509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	var (
		derBytes []byte
		// cert     *gmx509.Certificate
	)

	if selfSign {
		// 自签名证书
		publicKey := &leafKey.PublicKey
		derBytes, err = gmx509.CreateCertificate(&template, &template, publicKey, leafKey)
		if err != nil {
			return fmt.Errorf("failed to create self-signed leaf certificate: %w", err)
		}
		// cert = &template
	} else {
		// 使用根证书签发
		// 读取根证书
		rootCertBytes, err := os.ReadFile(rootCertPath)
		if err != nil {
			return fmt.Errorf("failed to read root certificate: %w", err)
		}

		rootCertBlock, _ := pem.Decode(rootCertBytes)
		if rootCertBlock == nil {
			return fmt.Errorf("failed to decode root certificate")
		}

		rootCert, err := gmx509.ParseCertificate(rootCertBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse root certificate: %w", err)
		}

		// 读取根私钥
		rootKeyBytes, err := os.ReadFile(rootKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read root key: %w", err)
		}

		rootKeyBlock, _ := pem.Decode(rootKeyBytes)
		if rootKeyBlock == nil {
			return fmt.Errorf("failed to decode root key")
		}

		// 解析根私钥
		var rootPrivateKey *gmsm2.PrivateKey
		if rootKeyBlock.Type == "PRIVATE KEY" {
			rootPrivateKey, err = gmx509.ReadPrivateKeyFromPem(rootKeyBytes, nil)
			if err != nil {
				return fmt.Errorf("failed to parse root private key: %w", err)
			}
		} else {
			return fmt.Errorf("unsupported root key type: %s", rootKeyBlock.Type)
		}

		// 签发证书
		publicKey := &leafKey.PublicKey
		derBytes, err = gmx509.CreateCertificate(&template, rootCert, publicKey, rootPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to create leaf certificate: %w", err)
		}
		// cert = rootCert
	}

	if stringer.IsBlank(certPath) {
		certPath = c.config.CertPath
	}

	// 保存证书
	certFilePath := filepath.Join(certPath, name+".crt")
	

	certOut, err := os.Create(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer func(certOut *os.File) {
		if err := certOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} certOut file failed: %v", name, err)
		}
	}(certOut)

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// 保存私钥
	keyFilePath := filepath.Join(certPath, name+".key")
	

	keyOut, err := os.Create(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func(keyOut *os.File) {
		if err := keyOut.Close(); err != nil {
			c.logger.Errorf("[control] close certificate {%v} keyOut file failed: %v", name, err)
		}
	}(keyOut)

	// 使用国密库的私钥序列化方法
	leafKeyBytes, err := gmx509.WritePrivateKeyToPem(leafKey, nil)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: leafKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	c.logger.Infof("[control] SM2 leaf certificate generated: %s, key: %s", certFilePath, keyFilePath)
	return nil
}

// GetCertificateInfo 获取证书信息
// @param certPath string 证书路径
// @return *Certificate 证书信息
// @return error 错误信息
func (c *Control) GetCertificateInfo(certPath string) (*Certificate, error) {
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
			Algorithm: "RSA/ECC",
			Subject:   cert.Subject.String(),
			Issuer:    cert.Issuer.String(),
			NotBefore: cert.NotBefore,
			NotAfter:  cert.NotAfter,
			Serial:    cert.SerialNumber,
		}
	} else {
		// 尝试解析为国密证书
		if cert, err := gmx509.ParseCertificate(certBlock.Bytes); err == nil {
			certInfo = &Certificate{
				Algorithm: "SM2",
				Subject:   cert.Subject.String(),
				Issuer:    cert.Issuer.String(),
				NotBefore: cert.NotBefore,
				NotAfter:  cert.NotAfter,
				Serial:    cert.SerialNumber,
			}
		} else {
			return nil, fmt.Errorf("failed to parse certificate")
		}
	}

	return certInfo, nil
}
