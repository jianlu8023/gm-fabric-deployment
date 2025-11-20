package certificate

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"
	
	gmx509 "github.com/tjfoc/gmsm/x509"
)

// formatKeyUsage 格式化密钥用途
func formatKeyUsage(keyUsage x509.KeyUsage) string {
	var usages []string
	if keyUsage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if keyUsage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if keyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if keyUsage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if keyUsage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if keyUsage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if keyUsage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if keyUsage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if keyUsage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	return strings.Join(usages, ", ")
}

// formatExtKeyUsage 格式化扩展密钥用途
func formatExtKeyUsage(extKeyUsage []x509.ExtKeyUsage) []string {
	var usages []string
	for _, usage := range extKeyUsage {
		switch usage {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "EmailProtection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSECEndSystem")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSECTunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSECUser")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "TimeStamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSPSigning")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			usages = append(usages, "MicrosoftServerGatedCrypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			usages = append(usages, "NetscapeServerGatedCrypto")
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			usages = append(usages, "MicrosoftCommercialCodeSigning")
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			usages = append(usages, "MicrosoftKernelCodeSigning")
		default:
			usages = append(usages, fmt.Sprintf("Unknown(%d)", usage))
		}
	}
	return usages
}

// formatGMKeyUsage 格式化国密密钥用途
func formatGMKeyUsage(keyUsage gmx509.KeyUsage) string {
	var usages []string
	if keyUsage&gmx509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if keyUsage&gmx509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if keyUsage&gmx509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if keyUsage&gmx509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if keyUsage&gmx509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if keyUsage&gmx509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if keyUsage&gmx509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if keyUsage&gmx509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if keyUsage&gmx509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	return strings.Join(usages, ", ")
}

// formatGMPublicKeyAlgorithm 格式化国密公钥算法
func formatGMPublicKeyAlgorithm(algo gmx509.PublicKeyAlgorithm) string {
	switch algo {
	case gmx509.RSA:
		return "RSA"
	case gmx509.UnknownPublicKeyAlgorithm:
		return "UnknownPublicKeyAlgorithm"
	case gmx509.DSA:
		return "DSA"
	case gmx509.ECDSA:
		return "ECDSA"
	case gmx509.SM2:
		return "SM2"
	default:
		return fmt.Sprintf("Unknown(%d)", algo)
	}
}

// formatGMExtKeyUsage 格式化国密扩展密钥用途
func formatGMExtKeyUsage(extKeyUsage []gmx509.ExtKeyUsage) []string {
	var usages []string
	for _, usage := range extKeyUsage {
		switch usage {
		case gmx509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case gmx509.ExtKeyUsageServerAuth:
			usages = append(usages, "ServerAuth")
		case gmx509.ExtKeyUsageClientAuth:
			usages = append(usages, "ClientAuth")
		case gmx509.ExtKeyUsageCodeSigning:
			usages = append(usages, "CodeSigning")
		case gmx509.ExtKeyUsageEmailProtection:
			usages = append(usages, "EmailProtection")
		case gmx509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSECEndSystem")
		case gmx509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSECTunnel")
		case gmx509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSECUser")
		case gmx509.ExtKeyUsageTimeStamping:
			usages = append(usages, "TimeStamping")
		case gmx509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSPSigning")
		default:
			usages = append(usages, fmt.Sprintf("Unknown(%d)", usage))
		}
	}
	return usages
}

// calculateFingerprintSHA1 计算SHA1指纹
func calculateFingerprintSHA1(raw []byte) string {
	hash := sha1.Sum(raw)
	return hex.EncodeToString(hash[:])
}

// calculateFingerprintSHA256 计算SHA256指纹
func calculateFingerprintSHA256(raw []byte) string {
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}
