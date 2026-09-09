package service

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/certificate"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// errInvalidCertPath 用户指定的证书路径非法或越出受信任目录
var errInvalidCertPath = errors.New("非法的证书路径：仅允许访问受信任证书目录下的证书文件")

// 允许用户通过 path/dir 参数访问的证书文件扩展名。
var certsAllowedExts = map[string]bool{
	".crt": true,
	".cer": true,
	".pem": true,
	".der": true,
}

// CertsService 证书服务接口
//
// @description 定义证书只读查询服务的方法，包括证书详情、证书目录列表与当前TLS连接信息
// @interface
type CertsService interface {
	// GetCertsTls 获取当前TLS连接信息
	//
	// @description 返回当前 HTTPS 连接的 TLS/mTLS 协商信息与对端证书信息
	// @param ctx *gin.Context Gin上下文
	// @param req *request.CertsTlsRequest TLS连接信息请求参数
	GetCertsTls(ctx *gin.Context, req *request.CertsTlsRequest)
}

// certsServiceImpl 证书服务结构体
//
// @description 提供证书只读查询功能，复用证书控制器与HTTP控制器的配置
// @struct
type certsServiceImpl struct {
	*Service                         // Service 基础服务，提供日志功能
	certControl *certificate.Control // certControl 证书控制器，可能为 nil（证书功能未启用）
	httpControl *http.Control        // httpControl HTTP控制器，用于读取TLS配置与连接状态
}

// NewCertsService 创建证书服务实例
//
// @description 创建并返回一个新的证书服务实例
// @param service *Service 基础服务
// @param certControl *certificate.Control 证书控制器（可为 nil）
// @param httpControl *http.Control HTTP控制器
// @return CertsService 证书服务实例
func NewCertsService(service *Service, certControl *certificate.Control, httpControl *http.Control) CertsService {
	return &certsServiceImpl{
		Service:     service,
		certControl: certControl,
		httpControl: httpControl,
	}
}

// GetCertsTls 获取当前TLS连接信息
//
// @description 返回当前 HTTPS 连接的 TLS/mTLS 协商信息与对端证书信息
// @param ctx *gin.Context Gin上下文
// @param req *request.CertsTlsRequest TLS连接信息请求参数
func (s *certsServiceImpl) GetCertsTls(ctx *gin.Context, req *request.CertsTlsRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "certsServiceImpl", "getCertsTls",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received certs tls request...")

	cfg := s.getServerConfig()
	resp := response.CertsTlsResponse{}
	if cfg != nil {
		resp.TlsEnabled = cfg.TlsEnabled
		resp.GmTls = cfg.TlsGM
	}

	// 标准库 net/http 仅在标准 crypto/tls 连接时填充 Request.TLS；国密 gmtls 连接下为 nil
	state := ctx.Request.TLS
	if state != nil {
		resp.Version = tlsVersionName(state.Version)
		resp.CipherSuite = tls.CipherSuiteName(state.CipherSuite)
		resp.NegotiatedProto = state.NegotiatedProtocol
		resp.ServerName = state.ServerName
		resp.HandshakeComplete = state.HandshakeComplete
		resp.ClientCertPresent = len(state.PeerCertificates) > 0
		peers := make([]response.PeerCertInfo, 0, len(state.PeerCertificates))
		for _, cert := range state.PeerCertificates {
			peers = append(peers, response.PeerCertInfo{
				Subject:           cert.Subject.String(),
				Issuer:            cert.Issuer.String(),
				NotBefore:         cert.NotBefore,
				NotAfter:          cert.NotAfter,
				DNSNames:          cert.DNSNames,
				Serial:            cert.SerialNumber.String(),
				FingerprintSHA256: certFingerprintSHA256(cert.Raw),
			})
		}
		resp.PeerCerts = peers
	} else {
		switch {
		case cfg != nil && cfg.TlsEnabled && cfg.TlsGM:
			resp.Note = "国密TLS连接的握手状态无法通过标准库 net/http 读取，GM mTLS 对端证书信息请查看服务端日志"
		case cfg == nil || !cfg.TlsEnabled:
			resp.Note = "当前为明文HTTP连接，无TLS信息"
		default:
			resp.Note = "无法读取当前连接的TLS状态"
		}
	}

	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}

// getServerConfig 获取HTTP服务器配置，httpControl 为 nil 时返回 nil
//
// @description 安全获取HTTP服务器配置，避免空指针
// @return *config.HttpServerConfig 服务器配置
func (s *certsServiceImpl) getServerConfig() *config.HttpServerConfig {
	if s.httpControl == nil {
		return nil
	}
	return s.httpControl.GetServerConfig()
}

// allowedCertRoots 返回允许访问的证书根目录集合（证书控制器目录 + 服务端TLS证书所在目录）
//
// @description 汇总所有受信任的证书目录，用于收敛用户通过 path/dir 参数访问的路径
// @return []string 允许访问的根目录绝对路径列表
func (s *certsServiceImpl) allowedCertRoots() []string {
	roots := make([]string, 0, 4)
	if s.certControl != nil {
		if cp := strings.TrimSpace(s.certControl.GetCertPath()); cp != "" {
			if abs, err := filepath.Abs(cp); err == nil {
				roots = append(roots, abs)
			}
		}
	}
	if cfg := s.getServerConfig(); cfg != nil {
		for _, cf := range cfg.TlsCertFile {
			if strings.TrimSpace(cf) != "" {
				if abs, err := filepath.Abs(filepath.Dir(cf)); err == nil {
					roots = append(roots, abs)
				}
			}
		}
		if strings.TrimSpace(cfg.TlsRCACertFile) != "" {
			if abs, err := filepath.Abs(filepath.Dir(cfg.TlsRCACertFile)); err == nil {
				roots = append(roots, abs)
			}
		}
	}
	return roots
}

// resolveCertPath 校验并收敛用户指定的证书文件路径
//
// @description 校验扩展名并确保目标文件位于受信任的证书目录内，返回绝对路径
// @param p string 用户输入的证书文件路径
// @return string 校验通过后的绝对路径
// @return error 扩展名非法或路径越界时的错误
func (s *certsServiceImpl) resolveCertPath(p string) (string, error) {
	abs, err := filepath.Abs(strings.TrimSpace(p))
	if err != nil {
		return "", err
	}
	if !certsAllowedExts[strings.ToLower(filepath.Ext(abs))] {
		return "", errInvalidCertPath
	}
	for _, root := range s.allowedCertRoots() {
		if isPathWithin(abs, root) {
			return abs, nil
		}
	}
	return "", errInvalidCertPath
}

// resolveCertDir 校验并收敛用户指定的证书目录
//
// @description 确保目标目录位于受信任的证书目录内（或为受信任目录本身），返回绝对路径
// @param d string 用户输入的证书目录
// @return string 校验通过后的绝对路径
// @return error 路径越界时的错误
func (s *certsServiceImpl) resolveCertDir(d string) (string, error) {
	abs, err := filepath.Abs(strings.TrimSpace(d))
	if err != nil {
		return "", err
	}
	for _, root := range s.allowedCertRoots() {
		if isPathWithin(abs, root) || abs == root {
			return abs, nil
		}
	}
	return "", errInvalidCertPath
}

// isPathWithin 判断目标路径是否位于根目录之内
//
// @description 通过计算相对路径判断 target 是否在 root 目录下，避免目录穿越
// @param target string 目标绝对路径
// @param root string 根目录绝对路径
// @return bool 是否在根目录内
func isPathWithin(target, root string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// tlsVersionName 将TLS版本常量转换为可读名称
//
// @description 把 crypto/tls 的版本号转换为 TLS1.2/TLS1.3 等可读字符串
// @param v uint16 TLS版本常量
// @return string 可读版本名
func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS1.3"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS10:
		return "TLS1.0"
	default:
		return "unknown"
	}
}

// certFingerprintSHA256 计算证书 DER 的 SHA-256 指纹（冒号分隔，便于与 openssl 对比）
//
// @description 对证书原始 DER 字节计算 SHA-256 并格式化为冒号分隔的十六进制串
// @param raw []byte 证书 DER 字节
// @return string 冒号分隔的 SHA-256 指纹
func certFingerprintSHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	h := strings.ToUpper(hex.EncodeToString(sum[:]))
	pairs := make([]string, len(h)/2)
	for i := range pairs {
		pairs[i] = h[2*i : 2*i+2]
	}
	return strings.Join(pairs, ":")
}
