package certificate

import (
	"math/big"
	"testing"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/stretchr/testify/assert"
)

var bigZero = big.NewInt(0)

func genLoggerControl() *logger.Control {
	return logger.NewLoggerControl(nil)
}

func TestParseSubject(t *testing.T) {
	assert := assert.New(t)

	subject := "/CN=Test CA/O=Test Organization/OU=Test Unit/C=CN/ST=Test State/L=Test City"
	name := parseSubject(subject)

	assert.Equal("Test CA", name.CommonName, "Expected CommonName 'Test CA'")
	assert.Equal([]string{"Test Organization"}, name.Organization, "Expected Organization 'Test Organization'")
	assert.Equal([]string{"Test Unit"}, name.OrganizationalUnit, "Expected OrganizationalUnit 'Test Unit'")
	assert.Equal([]string{"CN"}, name.Country, "Expected Country 'CN'")
	assert.Equal([]string{"Test State"}, name.Province, "Expected Province 'Test State'")
	assert.Equal([]string{"Test City"}, name.Locality, "Expected Locality 'Test City'")
}

func TestGenerateSerialNumber(t *testing.T) {
	assert := assert.New(t)

	serial, err := generateSerialNumber()
	assert.NoError(err, "Failed to generate serial number")
	assert.NotNil(serial, "Generated serial number is nil")

	// 验证序列号不为0
	assert.NotEqual(0, serial.Cmp(bigZero), "Generated serial number is zero")
}

func TestCertificateControl_StartUp(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}
	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 测试启动
	startupFailed := false
	control.StartUp(func(err error) {
		startupFailed = true
		assert.NoError(err, "StartUp should not fail")
	})

	assert.False(startupFailed, "StartUp callback should not be called")

	// 验证配置是否正确设置
	assert.Equal(cfg, control.config, "Configuration not set correctly")
}

func TestCertificateControl_GenerateRootCertificate(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}
	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 测试生成RSA根证书
	err := control.GenerateRootCertificate("test_root_rsa", "/CN=Test RSA Root CA", "RSA", "")
	assert.NoError(err, "Failed to generate RSA root certificate")

	// 测试生成ECC根证书
	err = control.GenerateRootCertificate("test_root_ecc", "/CN=Test ECC Root CA", "ECC", "")
	assert.NoError(err, "Failed to generate ECC root certificate")

	// 测试生成SM2根证书
	err = control.GenerateRootCertificate("test_root_sm2", "/CN=Test SM2 Root CA", "SM2", "")
	assert.NoError(err, "Failed to generate SM2 root certificate")
}

func TestCertificateControl_GenerateIntermediateCertificate(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping intermediate certificate test due to root certificate generation failure: %v", err)
	}

	// 测试生成中间证书
	err = control.GenerateIntermediateCertificate("test_intermediate", "/CN=Test Intermediate CA", "RSA",
		"", "", "")
	assert.NoError(err, "Failed to generate intermediate certificate")
}

func TestCertificateControl_GenerateLeafCertificate(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	assert.NoError(err, "Failed to generate root certificate")

	// 测试生成叶子证书（非自签）
	err = control.GenerateLeafCertificate("test_leaf", "/CN=Test Leaf Certificate", "RSA", false, "", "", "", "")
	assert.NoError(err, "Failed to generate leaf certificate")

	// 验证叶子证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_leaf.crt")
	assert.NoError(err, "Failed to get leaf certificate info")
	assert.NotNil(certInfo, "Leaf certificate info is nil")
	assert.Equal("CN=Test Leaf Certificate", certInfo.Subject, "Leaf certificate subject incorrect")
	assert.Equal("CN=Test Root CA,O=Test Organization", certInfo.Issuer, "Leaf certificate issuer incorrect")
}

// TestCertificateControl_GenerateLeafCertificateWithSAN 测试生成带有SAN的叶子证书
func TestCertificateControl_GenerateLeafCertificateWithSAN(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root_san", "/CN=Test Root CA", "RSA", "")
	assert.NoError(err, "Failed to generate root certificate")

	// 测试生成带有SAN的叶子证书
	// SAN格式：逗号分隔的域名列表
	sans := "example.com,www.example.com,test.example.com"
	err = control.GenerateLeafCertificate("test_leaf_san", "/CN=example.com", "RSA", false, "", "", "", sans)
	assert.NoError(err, "Failed to generate leaf certificate with SAN")

	// 验证叶子证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_leaf_san.crt")
	assert.NoError(err, "Failed to get leaf certificate info with SAN")
	assert.NotNil(certInfo, "Leaf certificate info with SAN is nil")
	assert.Equal("CN=example.com", certInfo.Subject, "Leaf certificate subject incorrect")
	assert.Equal("CN=Test Root CA,O=Test Organization", certInfo.Issuer, "Leaf certificate issuer incorrect")
}

func TestCertificateControl_GenerateLeafCertificateSelfSign(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 测试生成自签叶子证书
	err := control.GenerateLeafCertificate("test_leaf_selfsign", "/CN=Test Self-Signed Leaf Certificate", "RSA", true, "", "", "", "")
	assert.NoError(err, "Failed to generate self-signed leaf certificate")

	// 验证自签证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_leaf_selfsign.crt")
	assert.NoError(err, "Failed to get self-signed leaf certificate info")
	assert.NotNil(certInfo, "Self-signed leaf certificate info is nil")
	assert.Equal("CN=Test Self-Signed Leaf Certificate", certInfo.Subject, "Self-signed leaf certificate subject incorrect")
	assert.Equal("CN=Test Self-Signed Leaf Certificate", certInfo.Issuer, "Self-signed leaf certificate issuer should be same as subject")
}

func TestCertificateControl_GetCertificateInfo(t *testing.T) {
	assert := assert.New(t)

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())
	assert.NotNil(control, "Certificate control should not be nil")

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping certificate info test due to root certificate generation failure: %v", err)
	}

	// 测试获取证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_root.crt")
	assert.NoError(err, "Failed to get certificate info")
	assert.NotNil(certInfo, "Certificate info is nil")
	assert.Equal("RSA", certInfo.Algorithm, "Expected algorithm 'RSA'")
	assert.NotEmpty(certInfo.Subject, "Certificate subject should not be empty")
	assert.NotEmpty(certInfo.Issuer, "Certificate issuer should not be empty")
	assert.True(certInfo.NotBefore.Before(time.Now()), "Certificate notBefore should be in the past")
	assert.True(certInfo.NotAfter.After(time.Now()), "Certificate notAfter should be in the future")
	assert.NotNil(certInfo.Serial, "Certificate serial should not be nil")
	assert.NotEqual(0, certInfo.Serial.Cmp(bigZero), "Certificate serial should not be zero")
}
