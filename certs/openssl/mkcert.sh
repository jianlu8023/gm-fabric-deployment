#!/bin/sh
#
# OpenSSL RSA 证书四层结构一键生成脚本
# 证书层级: root -> intermediate -> {http,grpc,docker,applications,libp2p}-authority -> 实体证书
# 使用 openssl x509 -req 方式签发，RSA 密钥 + SHA256 摘要，单证书模式
#
# 依赖: openssl (1.1.x 或 3.x)
# 用法: chmod +x mkcert.sh && ./mkcert.sh
#
# 兼容性: 纯 POSIX sh, 可用 bash / dash / busybox ash 运行
#

set -e

# ============================================================
# 配置变量
# ============================================================

# openssl 命令路径
OPENSSL=${OPENSSL:-openssl}

# 有效期（天）
DAYS_ROOT=3650
DAYS_INTERMEDIATE=1825
DAYS_AUTHORITY=1095
DAYS_ENTITY=365

# ============================================================
# 颜色输出
# ============================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { printf "${BLUE}[INFO]${NC} %s\n" "$1"; }
success() { printf "${GREEN}[OK]${NC}   %s\n" "$1"; }
warn()    { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
error()   { printf "${RED}[ERROR]${NC} %s\n" "$1"; exit 1; }

# ============================================================
# 环境检查
# ============================================================

check_env() {
    info "检查 OpenSSL 环境..."

    if ! command -v "$OPENSSL" >/dev/null 2>&1; then
        error "未找到 $OPENSSL 命令，请确认 OpenSSL 已安装并配置 PATH"
    fi

    version=$("$OPENSSL" version 2>&1 | head -1)
    success "OpenSSL 版本: $version"
}

# ============================================================
# 清理旧文件
# ============================================================

cleanup() {
    info "清理旧的证书文件（保留 .cnf 配置文件）..."
    rm -f *.key *.csr *.crt *.srl *-chain.crt
    success "清理完成"
}

# ============================================================
# 生成 RSA 私钥
# ============================================================

gen_rsa_key() {
    outfile="$1"
    bits="${2:-2048}"
    "$OPENSSL" genrsa -out "$outfile" "$bits"
}

# ============================================================
# 第一层: 根证书 (root)
# ============================================================

gen_root() {
    info "========== 第一层: 生成根证书 (root) =========="

    info "生成根证书私钥 (4096位)..."
    gen_rsa_key root.key 4096

    info "生成自签名根证书 (有效期 ${DAYS_ROOT} 天)..."
    "$OPENSSL" req -new -x509 -nodes -key root.key -sha256 -days "$DAYS_ROOT" \
        -out root.crt -config root.cnf

    success "根证书生成完成: root.crt"
}

# ============================================================
# 第二层: 中间证书 (intermediate)
# ============================================================

gen_intermediate() {
    info "========== 第二层: 生成中间证书 (intermediate) =========="

    info "生成中间证书私钥 (4096位)..."
    gen_rsa_key intermediate.key 4096

    info "生成中间证书 CSR..."
    "$OPENSSL" req -new -key intermediate.key -sha256 -out intermediate.csr -config intermediate.cnf

    info "使用根证书签发中间证书 (有效期 ${DAYS_INTERMEDIATE} 天)..."
    "$OPENSSL" x509 -req -in intermediate.csr -CA root.crt -CAkey root.key \
        -CAcreateserial -out intermediate.crt -days "$DAYS_INTERMEDIATE" -sha256 \
        -extfile intermediate.cnf -extensions v3_intermediate_ca

    success "中间证书生成完成: intermediate.crt"
}

# ============================================================
# 第三层: 应用级 CA
# ============================================================

gen_authority() {
    name="$1"      # authority 名称, 如 http
    display="$2"   # 显示名, 如 HTTP

    info "========== 第三层: 生成 ${display}-authority =========="

    cnf_file="${name}-authority.cnf"
    key_file="${name}-authority.key"
    csr_file="${name}-authority.csr"
    crt_file="${name}-authority.crt"

    info "生成 ${name}-authority 私钥 (4096位)..."
    gen_rsa_key "$key_file" 4096

    info "生成 ${name}-authority CSR..."
    "$OPENSSL" req -new -key "$key_file" -sha256 -out "$csr_file" -config "$cnf_file"

    info "使用中间证书签发 ${name}-authority (有效期 ${DAYS_AUTHORITY} 天)..."
    "$OPENSSL" x509 -req -in "$csr_file" -CA intermediate.crt -CAkey intermediate.key \
        -CAcreateserial -out "$crt_file" -days "$DAYS_AUTHORITY" -sha256 \
        -extfile "$cnf_file" -extensions v3_application_ca

    success "${display}-authority 生成完成: ${crt_file}"
}

# ============================================================
# 第四层: 实体证书 (单证书模式)
# ============================================================

gen_entity_cert() {
    authority_name="$1"   # 签发此证书的 authority 名称, 如 http
    cert_name="$2"        # 证书文件名前缀, 如 hserver
    ext_section="$3"      # 扩展段名, 如 v3_http_cert
    desc="$4"             # 描述, 如 HTTP服务器证书

    cnf_file="${cert_name}.cnf"
    key_file="${cert_name}.key"
    csr_file="${cert_name}.csr"
    crt_file="${cert_name}.crt"

    info "生成 ${desc} 私钥 (2048位)..."
    gen_rsa_key "$key_file" 2048

    info "生成 ${desc} CSR..."
    "$OPENSSL" req -new -key "$key_file" -sha256 -out "$csr_file" -config "$cnf_file"

    info "使用 ${authority_name}-authority 签发 ${desc} (有效期 ${DAYS_ENTITY} 天)..."
    "$OPENSSL" x509 -req -in "$csr_file" \
        -CA "${authority_name}-authority.crt" \
        -CAkey "${authority_name}-authority.key" \
        -CAcreateserial -out "$crt_file" -days "$DAYS_ENTITY" -sha256 \
        -extfile "$cnf_file" -extensions "$ext_section"

    success "${desc} 生成完成: ${crt_file}"
}

# ============================================================
# 生成各 authority 的实体证书
# ============================================================

gen_http_certs() {
    info "========== 第四层: 生成 HTTP 实体证书 =========="

    # HTTP 服务器证书
    gen_entity_cert "http" "hserver" "v3_http_cert" "HTTP服务器证书"

    # HTTP 客户端证书
    gen_entity_cert "http" "hclient" "v3_http_cert" "HTTP客户端证书"

    success "HTTP 实体证书全部生成完成"
}

gen_grpc_certs() {
    info "========== 第四层: 生成 gRPC 实体证书 =========="

    # gRPC 服务器证书
    gen_entity_cert "grpc" "gserver" "v3_grpc_server_cert" "gRPC服务器证书"

    # gRPC 客户端证书
    gen_entity_cert "grpc" "gclient" "v3_grpc_client_cert" "gRPC客户端证书"

    success "gRPC 实体证书全部生成完成"
}

gen_docker_certs() {
    info "========== 第四层: 生成 Docker 实体证书 =========="

    # Docker 服务器证书
    gen_entity_cert "docker" "dserver" "v3_docker_server_cert" "Docker服务器证书"

    # Docker 客户端证书
    gen_entity_cert "docker" "dclient" "v3_docker_client_cert" "Docker客户端证书"

    success "Docker 实体证书全部生成完成"
}

gen_libp2p_certs() {
    info "========== 第四层: 生成 libp2p 实体证书 =========="

    # libp2p 服务器证书
    gen_entity_cert "libp2p" "lserver" "v3_libp2p_server_cert" "libp2p服务器证书"

    # libp2p 客户端证书
    gen_entity_cert "libp2p" "lclient" "v3_libp2p_client_cert" "libp2p客户端证书"

    success "libp2p 实体证书全部生成完成"
}

# ============================================================
# 生成证书链文件
# ============================================================

gen_chains() {
    info "========== 生成证书链文件 =========="

    # root + intermediate 链
    cat root.crt intermediate.crt > root-intermediate-chain.crt
    success "生成: root-intermediate-chain.crt"

    # 每个 authority 的完整链 (root + intermediate + authority)
    for auth in http grpc docker applications libp2p; do
        cat root.crt intermediate.crt "${auth}-authority.crt" > "${auth}-full-chain.crt"
        success "生成: ${auth}-full-chain.crt"
    done
}

# ============================================================
# 验证证书链
# ============================================================

verify_one() {
    cafile="$1"
    cert="$2"
    desc="$3"

    if [ ! -f "$cert" ]; then
        warn "${desc}: 文件不存在，跳过"
        return
    fi

    if "$OPENSSL" verify -CAfile "$cafile" "$cert" >/dev/null 2>&1; then
        success "${desc}: OK"
        passed=$((passed+1))
    else
        warn "${desc}: FAILED"
        failed=$((failed+1))
    fi
}

verify_certs() {
    info "========== 验证证书链 =========="

    passed=0
    failed=0

    # 验证根证书（自签名）
    if "$OPENSSL" verify -CAfile root.crt root.crt >/dev/null 2>&1; then
        success "root (自签名): OK"
        passed=$((passed+1))
    else
        warn "root (自签名): FAILED"
        failed=$((failed+1))
    fi

    # 验证中间证书
    verify_one root.crt intermediate.crt "intermediate (由 root 签发)"

    # 验证各 authority 证书
    for auth in http grpc docker applications libp2p; do
        verify_one root-intermediate-chain.crt "${auth}-authority.crt" "${auth}-authority"
    done

    # 验证各实体证书
    verify_one http-full-chain.crt hserver.crt "hserver"
    verify_one http-full-chain.crt hclient.crt "hclient"
    verify_one grpc-full-chain.crt gserver.crt "gserver"
    verify_one grpc-full-chain.crt gclient.crt "gclient"
    verify_one docker-full-chain.crt dserver.crt "dserver"
    verify_one docker-full-chain.crt dclient.crt "dclient"
    verify_one libp2p-full-chain.crt lserver.crt "lserver"
    verify_one libp2p-full-chain.crt lclient.crt "lclient"

    echo ""
    info "验证结果: 通过 ${passed}, 失败 ${failed}"

    if [ "$failed" -gt 0 ]; then
        warn "有 ${failed} 个证书验证失败，请检查"
    else
        success "所有证书验证通过！"
    fi
}

# ============================================================
# 显示证书过期信息
# ============================================================

show_cert_days() {
    cert="$1"
    desc="$2"

    if [ ! -f "$cert" ]; then
        return
    fi

    enddate=$("$OPENSSL" x509 -in "$cert" -noout -enddate 2>/dev/null | cut -d= -f2)
    if [ -z "$enddate" ]; then
        return
    fi

    end_epoch=$(date -d "$enddate" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$enddate" +%s 2>/dev/null)
    now_epoch=$(date +%s)

    if [ -n "$end_epoch" ]; then
        days_left=$(( (end_epoch - now_epoch) / 86400 ))
        if [ "$days_left" -lt 30 ]; then
            warn "${desc}: ${days_left} 天后过期 (${enddate})"
        elif [ "$days_left" -lt 90 ]; then
            warn "${desc}: ${days_left} 天后过期 (${enddate})"
        else
            success "${desc}: ${days_left} 天后过期 (${enddate})"
        fi
    fi
}

show_expiry_info() {
    info "========== 证书有效期信息 =========="

    show_cert_days root.crt "root"
    show_cert_days intermediate.crt "intermediate"
    show_cert_days http-authority.crt "http-authority"
    show_cert_days hserver.crt "hserver"
    show_cert_days hclient.crt "hclient"
    show_cert_days gserver.crt "gserver"
    show_cert_days gclient.crt "gclient"
    show_cert_days dserver.crt "dserver"
    show_cert_days dclient.crt "dclient"
    show_cert_days lserver.crt "lserver"
    show_cert_days lclient.crt "lclient"
}

# ============================================================
# 统计输出
# ============================================================

show_summary() {
    echo ""
    echo "============================================================"
    echo "  OpenSSL RSA 证书四层结构生成完成"
    echo "============================================================"
    echo ""
    echo "  证书层级结构:"
    echo "    root (根CA, 4096位)"
    echo "    └── intermediate (中间CA, 4096位)"
    echo "        ├── http-authority"
    echo "        │   ├── hserver.crt"
    echo "        │   └── hclient.crt"
    echo "        ├── grpc-authority"
    echo "        │   ├── gserver.crt"
    echo "        │   └── gclient.crt"
    echo "        ├── docker-authority"
    echo "        │   ├── dserver.crt"
    echo "        │   └── dclient.crt"
    echo "        ├── applications-authority"
    echo "        └── libp2p-authority"
    echo "            ├── lserver.crt"
    echo "            └── lclient.crt"
    echo ""

    key_count=$(ls *.key 2>/dev/null | wc -l)
    crt_count=$(ls *.crt 2>/dev/null | wc -l)
    cnf_count=$(ls *.cnf 2>/dev/null | wc -l)

    echo "  文件统计:"
    echo "    配置文件 (.cnf): ${cnf_count} 个"
    echo "    私钥文件 (.key): ${key_count} 个"
    echo "    证书文件 (.crt): ${crt_count} 个"
    echo ""
    echo "  算法: RSA (密钥) + SHA256 (摘要)"
    echo "  模式: 单证书 (签名和加密共用)"
    echo ""
    echo "============================================================"
}

# ============================================================
# 主流程
# ============================================================

main() {
    echo ""
    echo "============================================================"
    echo "  OpenSSL RSA 证书四层结构一键生成脚本"
    echo "============================================================"
    echo ""

    # 1. 环境检查
    check_env

    # 2. 清理旧文件
    cleanup

    # 3. 第一层: 根证书
    gen_root

    # 4. 第二层: 中间证书
    gen_intermediate

    # 5. 第三层: 应用级 CA (5个)
    gen_authority "http" "HTTP"
    gen_authority "grpc" "GRPC"
    gen_authority "docker" "DOCKER"
    gen_authority "applications" "APPLICATIONS"
    gen_authority "libp2p" "LIBP2P"

    # 6. 第四层: 实体证书
    gen_http_certs
    gen_grpc_certs
    gen_docker_certs
    gen_libp2p_certs

    # 7. 生成证书链
    gen_chains

    # 8. 验证证书链
    verify_certs

    # 9. 显示过期信息
    show_expiry_info

    # 10. 输出统计
    show_summary
}

# 执行主流程
main
