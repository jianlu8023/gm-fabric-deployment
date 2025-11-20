package certificate

import (
	"fmt"
	"math/big"
	"strings"
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

	subject := "/CN=Test CA/O=Test Organization/OU=Test Unit/C=CN/ST=Test State/L=Test City"
	name := parseSubject(subject)

	assert.Equal(t, "Test CA", name.CommonName, "Expected CommonName 'Test CA'")
	assert.Equal(t, []string{"Test Organization"}, name.Organization, "Expected Organization 'Test Organization'")
	assert.Equal(t, []string{"Test Unit"}, name.OrganizationalUnit, "Expected OrganizationalUnit 'Test Unit'")
	assert.Equal(t, []string{"CN"}, name.Country, "Expected Country 'CN'")
	assert.Equal(t, []string{"Test State"}, name.Province, "Expected Province 'Test State'")
	assert.Equal(t, []string{"Test City"}, name.Locality, "Expected Locality 'Test City'")
}

func TestGenerateSerialNumber(t *testing.T) {

	serial, err := generateSerialNumber()
	assert.NoError(t, err, "Failed to generate serial number")
	assert.NotNil(t, serial, "Generated serial number is nil")

	// 验证序列号不为0
	assert.NotEqual(t, 0, serial.Cmp(bigZero), "Generated serial number is zero")
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
	assert.NotNil(t, control, "Certificate control should not be nil")

	// 测试启动
	startupFailed := false
	control.StartUp(func(err error) {
		startupFailed = true
		assert.NoError(t, err, "StartUp should not fail")
	})

	assert.False(t, startupFailed, "StartUp callback should not be called")

	// 验证配置是否正确设置
	assert.Equal(t, cfg, control.config, "Configuration not set correctly")
}

func TestCertificateControl_GenerateRootCertificate(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.CertificateConfig
		algo     string
		certName string
		subject  string
	}{
		{
			name: "RSA-root",
			algo: "RSA",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "RSA",
				RootSubject: "/CN=Test RSA Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			certName: "test_root_rsa",
			subject:  "/CN=Test RSA Root CA/O=Test Organization",
		},
		{
			name: "ECC-root",
			algo: "ECC",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "ECC",
				RootSubject: "/CN=Test ECC Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			certName: "test_root_ecc",
			subject:  "/CN=Test ECC Root CA/O=Test Organization",
		},
		{
			name: "SM2-root",
			algo: "SM2",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "SM2",
				RootSubject: "/CN=Test SM2 Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			certName: "test_root_sm2",
			subject:  "/CN=Test SM2 Root CA/O=Test Organization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := NewCertificateControl(tt.cfg, genLoggerControl())
			assert.NotNil(t, control, "control should not be nil")
			err := control.GenerateRootCertificate(tt.certName, tt.subject, tt.algo, tt.cfg.CertPath)
			assert.NoError(t, err, "generate root certificate should not return error")
		})
	}
}

func TestCertificateControl_GenerateIntermediateCertificate(t *testing.T) {

	tests := []struct {
		name                 string
		cfg                  *config.CertificateConfig
		algo                 string
		intermediateCertName string
		intermediateSubject  string
		rootCertName         string
		rootKeyName          string
		rootName             string
	}{
		{
			name: "RSA-intermediate",
			algo: "RSA",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "RSA",
				RootSubject: "/CN=Test RSA Root CA",
				CertPath:    "./tmp/test_certs",
			},
			intermediateCertName: "test_intermediate_rsa",
			intermediateSubject:  "/CN=Test RSA Intermediate CA",
			rootCertName:         "./tmp/test_certs/test_root_rsa.crt",
			rootKeyName:          "./tmp/test_certs/test_root_rsa.key",
			rootName:             "test_root_rsa",
		},
		{
			name: "ECC-intermediate",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "ECC",
				RootSubject: "/CN=Test ECC Root CA",
				CertPath:    "./tmp/test_certs",
			},
			algo:                 "ECC",
			intermediateCertName: "test_intermediate_ecc",
			intermediateSubject:  "/CN=Test ECC Intermediate CA",
			rootCertName:         "./tmp/test_certs/test_root_ecc.crt",
			rootKeyName:          "./tmp/test_certs/test_root_ecc.key",
			rootName:             "test_root_ecc",
		},
		{
			name: "SM2-intermediate",
			algo: "SM2",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "SM2",
				RootSubject: "/CN=Test SM2 Root CA",
				CertPath:    "./tmp/test_certs",
			},
			intermediateCertName: "test_intermediate_sm2",
			intermediateSubject:  "/CN=Test SM2 Intermediate CA",
			rootCertName:         "./tmp/test_certs/test_root_sm2.crt",
			rootKeyName:          "./tmp/test_certs/test_root_sm2.key",
			rootName:             "test_root_sm2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := NewCertificateControl(tt.cfg, genLoggerControl())
			assert.NotNil(t, control, "control should not be nil")
			err := control.GenerateRootCertificate(tt.rootName, tt.cfg.RootSubject, tt.algo, tt.cfg.CertPath)
			assert.NoError(t, err, "generate root certificate should not return error")
			err = control.GenerateIntermediateCertificate(tt.intermediateCertName, tt.intermediateSubject, tt.algo, tt.rootCertName, tt.rootKeyName, tt.cfg.CertPath)
			assert.NoError(t, err, "generate intermediate certificate should not return error")
		})
	}
}

func TestCertificateControl_GenerateLeafCertificate(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.CertificateConfig
		leafCertName    string
		leafSubject     string
		algo            string
		selfSign        bool
		signCertPath    string
		signKeyPath     string
		sans            string
		expectedSubject string
		expectedIssuer  string
	}{
		{
			name: "RSA-leaf",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "RSA",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_rsa",
			leafSubject:     "/CN=Test Leaf Certificate RSA",
			algo:            "RSA",
			selfSign:        false,
			signCertPath:    "./tmp/test_certs/test_root_rsa.crt",
			signKeyPath:     "./tmp/test_certs/test_root_rsa.key",
			sans:            "",
			expectedSubject: "CN=Test Leaf Certificate RSA",
			expectedIssuer:  "CN=Test Root CA,O=Test Organization",
		},
		{
			name: "ECC-leaf",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "ECC",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_ecc",
			leafSubject:     "/CN=Test Leaf Certificate ECC",
			algo:            "ECC",
			selfSign:        false,
			signCertPath:    "./tmp/test_certs/test_root_ecc.crt",
			signKeyPath:     "./tmp/test_certs/test_root_ecc.key",
			sans:            "",
			expectedSubject: "CN=Test Leaf Certificate ECC",
			expectedIssuer:  "CN=Test Root CA,O=Test Organization",
		},
		{
			name: "SM2-leaf",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "SM2",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_sm2",
			leafSubject:     "/CN=Test Leaf Certificate SM2",
			algo:            "SM2",
			selfSign:        false,
			signCertPath:    "./tmp/test_certs/test_root_sm2.crt",
			signKeyPath:     "./tmp/test_certs/test_root_sm2.key",
			sans:            "",
			expectedSubject: "CN=Test Leaf Certificate SM2",
			expectedIssuer:  "CN=Test Root CA,O=Test Organization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := NewCertificateControl(tt.cfg, genLoggerControl())
			assert.NotNil(t, control, "control should not be nil")

			// 首先生成根证书
			err := control.GenerateRootCertificate("test_root_"+strings.ToLower(tt.algo), tt.cfg.RootSubject, tt.algo, "")
			assert.NoError(t, err, "Failed to generate root certificate")

			// 测试生成叶子证书（非自签）
			err = control.GenerateLeafCertificate(tt.leafCertName, tt.leafSubject, tt.algo, tt.selfSign, tt.signCertPath, tt.signKeyPath, "", tt.sans)
			assert.NoError(t, err, "Failed to generate leaf certificate")

			// 验证叶子证书信息
			certInfo, err := control.GetCertificateInfo(fmt.Sprintf("./tmp/test_certs/%s.crt", tt.leafCertName))
			assert.NoError(t, err, "Failed to get leaf certificate info")
			assert.NotNil(t, certInfo, "Leaf certificate info is nil")
			assert.Equal(t, tt.expectedSubject, certInfo.Subject, "Leaf certificate subject incorrect")
			assert.Equal(t, tt.expectedIssuer, certInfo.Issuer, "Leaf certificate issuer incorrect")
		})
	}
}

// TestCertificateControl_GenerateLeafCertificateWithSAN 测试生成带有SAN的叶子证书
func TestCertificateControl_GenerateLeafCertificateWithSAN(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.CertificateConfig
		rootCertName    string
		rootSubject     string
		leafCertName    string
		leafSubject     string
		signCertPath    string
		signKeyPath     string
		algo            string
		sans            string
		expectedSubject string
		expectedIssuer  string
	}{
		{
			name: "RSA-leaf-with-san",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "RSA",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			rootCertName:    "test_root_san_rsa",
			signCertPath:    "./tmp/test_certs/test_root_san_rsa.crt",
			signKeyPath:     "./tmp/test_certs/test_root_san_rsa.key",
			rootSubject:     "/CN=Test Root CA",
			leafCertName:    "test_leaf_san_rsa",
			leafSubject:     "/CN=example.com",
			algo:            "RSA",
			sans:            "example.com,www.example.com,test.example.com",
			expectedSubject: "CN=example.com",
			expectedIssuer:  "CN=Test Root CA",
		},
		{
			name: "ECC-leaf-with-san",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "ECC",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			rootCertName:    "test_root_san_ecc",
			signCertPath:    "./tmp/test_certs/test_root_san_ecc.crt",
			signKeyPath:     "./tmp/test_certs/test_root_san_ecc.key",
			rootSubject:     "/CN=Test Root CA",
			leafCertName:    "test_leaf_san_ecc",
			leafSubject:     "/CN=example.com",
			algo:            "ECC",
			sans:            "example.com,www.example.com,test.example.com",
			expectedSubject: "CN=example.com",
			expectedIssuer:  "CN=Test Root CA",
		},
		{
			name: "SM2-leaf-with-san",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "SM2",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			rootCertName:    "test_root_san_sm2",
			signCertPath:    "./tmp/test_certs/test_root_san_sm2.crt",
			signKeyPath:     "./tmp/test_certs/test_root_san_sm2.key",
			rootSubject:     "/CN=Test Root CA",
			leafCertName:    "test_leaf_san_sm2",
			leafSubject:     "/CN=example.com",
			algo:            "SM2",
			sans:            "example.com,www.example.com,test.example.com",
			expectedSubject: "CN=example.com",
			expectedIssuer:  "CN=Test Root CA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := NewCertificateControl(tt.cfg, genLoggerControl())
			assert.NotNil(t, control, "Certificate control should not be nil")

			// 首先生成根证书
			err := control.GenerateRootCertificate(tt.rootCertName, tt.rootSubject, tt.algo, "")
			assert.NoError(t, err, "Failed to generate root certificate")

			// 测试生成带有SAN的叶子证书
			err = control.GenerateLeafCertificate(tt.leafCertName, tt.leafSubject, tt.algo, false, tt.signCertPath, tt.signKeyPath, "", tt.sans)
			assert.NoError(t, err, "Failed to generate leaf certificate with SAN")

			// 验证叶子证书信息
			certInfo, err := control.GetCertificateInfo(fmt.Sprintf("./tmp/test_certs/%s.crt", tt.leafCertName))
			assert.NoError(t, err, "Failed to get leaf certificate info with SAN")
			assert.NotNil(t, certInfo, "Leaf certificate info with SAN is nil")
			assert.Equal(t, tt.expectedSubject, certInfo.Subject, "Leaf certificate subject incorrect")
			assert.Equal(t, tt.expectedIssuer, certInfo.Issuer, "Leaf certificate issuer incorrect")
		})
	}
}

func TestCertificateControl_GenerateLeafCertificateSelfSign(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.CertificateConfig
		leafCertName    string
		leafSubject     string
		algo            string
		expectedSubject string
		expectedIssuer  string
	}{
		{
			name: "RSA-leaf-selfsign",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "RSA",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_selfsign_rsa",
			leafSubject:     "/CN=Test Self-Signed Leaf Certificate RSA",
			algo:            "RSA",
			expectedSubject: "CN=Test Self-Signed Leaf Certificate RSA",
			expectedIssuer:  "CN=Test Self-Signed Leaf Certificate RSA",
		},
		{
			name: "ECC-leaf-selfsign",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "ECC",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_selfsign_ecc",
			leafSubject:     "/CN=Test Self-Signed Leaf Certificate ECC",
			algo:            "ECC",
			expectedSubject: "CN=Test Self-Signed Leaf Certificate ECC",
			expectedIssuer:  "CN=Test Self-Signed Leaf Certificate ECC",
		},
		{
			name: "SM2-leaf-selfsign",
			cfg: &config.CertificateConfig{
				Enabled:     true,
				DefaultAlgo: "SM2",
				RootSubject: "/CN=Test Root CA/O=Test Organization",
				CertPath:    "./tmp/test_certs",
			},
			leafCertName:    "test_leaf_selfsign_sm2",
			leafSubject:     "/CN=Test Self-Signed Leaf Certificate SM2",
			algo:            "SM2",
			expectedSubject: "CN=Test Self-Signed Leaf Certificate SM2",
			expectedIssuer:  "CN=Test Self-Signed Leaf Certificate SM2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := NewCertificateControl(tt.cfg, genLoggerControl())
			assert.NotNil(t, control, "Certificate control should not be nil")

			// 测试生成自签叶子证书
			err := control.GenerateLeafCertificate(tt.leafCertName, tt.leafSubject, tt.algo, true, "", "", "", "")
			assert.NoError(t, err, "Failed to generate self-signed leaf certificate")

			// 验证自签证书信息
			certInfo, err := control.GetCertificateInfo(fmt.Sprintf("./tmp/test_certs/%s.crt", tt.leafCertName))
			assert.NoError(t, err, "Failed to get self-signed leaf certificate info")
			assert.NotNil(t, certInfo, "Self-signed leaf certificate info is nil")
			assert.Equal(t, tt.expectedSubject, certInfo.Subject, "Self-signed leaf certificate subject incorrect")
			assert.Equal(t, tt.expectedIssuer, certInfo.Issuer, "Self-signed leaf certificate issuer should be same as subject")
		})
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
	assert.NotNil(t, control, "Certificate control should not be nil")

	// 首先生成根证书
	err := control.GenerateRootCertificate("test_root", "/CN=Test Root CA", "RSA", "")
	if err != nil {
		t.Skipf("Skipping certificate info test due to root certificate generation failure: %v", err)
	}

	// 测试获取证书信息
	certInfo, err := control.GetCertificateInfo("./tmp/test_certs/test_root.crt")
	assert.NoError(t, err, "Failed to get certificate info")
	assert.NotNil(t, certInfo, "Certificate info is nil")
	assert.Equal(t, "RSA", certInfo.Algorithm, "Expected algorithm 'RSA'")
	assert.NotEmpty(t, certInfo.Subject, "Certificate subject should not be empty")
	assert.NotEmpty(t, certInfo.Issuer, "Certificate issuer should not be empty")
	assert.True(t, certInfo.NotBefore.Before(time.Now()), "Certificate notBefore should be in the past")
	assert.True(t, certInfo.NotAfter.After(time.Now()), "Certificate notAfter should be in the future")
	assert.NotNil(t, certInfo.Serial, "Certificate serial should not be nil")
	assert.NotEqual(t, 0, certInfo.Serial.Cmp(bigZero), "Certificate serial should not be zero")
}
