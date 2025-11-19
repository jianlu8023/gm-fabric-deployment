package certificate

import (
	"math/big"
	"testing"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

var bigZero = big.NewInt(0)

func genLoggerControl() *logger.Control {
	return logger.NewLoggerControl(nil)
}

func TestParseSubject(t *testing.T) {
	subject := "/CN=Test CA/O=Test Organization/OU=Test Unit/C=CN/ST=Test State/L=Test City"
	name := parseSubject(subject)

	if name.CommonName != "Test CA" {
		t.Errorf("Expected CommonName 'Test CA', got '%s'", name.CommonName)
	}

	if len(name.Organization) != 1 || name.Organization[0] != "Test Organization" {
		t.Errorf("Expected Organization 'Test Organization', got '%v'", name.Organization)
	}

	if len(name.OrganizationalUnit) != 1 || name.OrganizationalUnit[0] != "Test Unit" {
		t.Errorf("Expected OrganizationalUnit 'Test Unit', got '%v'", name.OrganizationalUnit)
	}

	if len(name.Country) != 1 || name.Country[0] != "CN" {
		t.Errorf("Expected Country 'CN', got '%v'", name.Country)
	}

	if len(name.Province) != 1 || name.Province[0] != "Test State" {
		t.Errorf("Expected Province 'Test State', got '%v'", name.Province)
	}

	if len(name.Locality) != 1 || name.Locality[0] != "Test City" {
		t.Errorf("Expected Locality 'Test City', got '%v'", name.Locality)
	}
}

func TestGenerateSerialNumber(t *testing.T) {
	serial, err := generateSerialNumber()
	if err != nil {
		t.Errorf("Failed to generate serial number: %v", err)
	}

	if serial == nil {
		t.Error("Generated serial number is nil")
	}

	// 验证序列号不为0
	if serial.Cmp(bigZero) == 0 {
		t.Error("Generated serial number is zero")
	}
}

func TestCertificateControl_StartUp(t *testing.T) {

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}
	control := NewCertificateControl(cfg, genLoggerControl())

	// 测试启动
	control.StartUp(func(err error) {
		t.Errorf("StartUp failed: %v", err)
	})

	// 验证配置是否正确设置
	if control.config != cfg {
		t.Error("Configuration not set correctly")
	}
}

func TestCertificateControl_GenerateRootCertificate(t *testing.T) {

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}
	control := NewCertificateControl(cfg, genLoggerControl())

	// 测试生成RSA根证书
	err := control.GenerateRootCertificate("test_root_rsa", "/CN=Test RSA Root CA", "RSA", "")
	if err != nil {
		t.Errorf("Failed to generate RSA root certificate: %v", err)
	}

	// 测试生成ECC根证书
	err = control.GenerateRootCertificate("test_root_ecc", "/CN=Test ECC Root CA", "ECC", "")
	if err != nil {
		t.Errorf("Failed to generate ECC root certificate: %v", err)
	}

	// 测试生成SM2根证书
	err = control.GenerateRootCertificate("test_root_sm2", "/CN=Test SM2 Root CA", "SM2", "")
	if err != nil {
		t.Errorf("Failed to generate SM2 root certificate: %v", err)
	}
}

func TestCertificateControl_GenerateIntermediateCertificate(t *testing.T) {

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping intermediate certificate test due to root certificate generation failure: %v", err)
	}

	// 测试生成中间证书
	err = control.GenerateIntermediateCertificate("test_intermediate", "/CN=Test Intermediate CA", "RSA",
		"", "", "")
	if err != nil {
		t.Errorf("Failed to generate intermediate certificate: %v", err)
	}
}

func TestCertificateControl_GenerateLeafCertificate(t *testing.T) {

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping leaf certificate test due to root certificate generation failure: %v", err)
	}

	// 测试生成叶子证书
	err = control.GenerateLeafCertificate("test_leaf", "/CN=Test Leaf Certificate", "RSA", false, "", "", "")
	if err != nil {
		t.Errorf("Failed to generate leaf certificate: %v", err)
	}
}

func TestCertificateControl_GetCertificateInfo(t *testing.T) {

	// 创建测试配置
	cfg := &config.CertificateConfig{
		Enabled:     true,
		DefaultAlgo: "RSA",
		RootSubject: "/CN=Test Root CA/O=Test Organization",
		CertPath:    "./tmp/test_certs",
	}

	control := NewCertificateControl(cfg, genLoggerControl())

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping certificate info test due to root certificate generation failure: %v", err)
	}

	// 测试获取证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_root.crt")
	if err != nil {
		t.Errorf("Failed to get certificate info: %v", err)
	}

	if certInfo == nil {
		t.Error("Certificate info is nil")
	}

	if certInfo.Algorithm != "RSA/ECC" {
		t.Errorf("Expected algorithm 'RSA/ECC', got '%s'", certInfo.Algorithm)
	}

	if certInfo.Subject == "" {
		t.Error("Certificate subject is empty")
	}

	if certInfo.Issuer == "" {
		t.Error("Certificate issuer is empty")
	}

	if certInfo.Serial == nil {
		t.Error("Certificate serial is nil")
	}

	// 验证有效期
	if certInfo.NotBefore.After(time.Now()) {
		t.Error("Certificate notBefore is in the future")
	}

	if certInfo.NotAfter.Before(time.Now()) {
		t.Error("Certificate notAfter is in the past")
	}
}
