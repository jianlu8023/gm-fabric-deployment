#!/bin/sh
#
# 铜锁国密双证书四层结构一键生成脚本
# 证书层级: root -> intermediate -> {http,grpc,docker,applications,libp2p}-authority -> 实体双证书
# 使用 tongsuossl x509 -req 方式签发
#
# 依赖: tongsuossl (Tongsuo 8.x)
# 用法: chmod +x mkcert.sh && ./mkcert.sh
#
# 兼容性: 纯 POSIX sh, 可用 bash / dash / busybox ash 运行
#

set -e

# ============================================================
# 配置变量
# ============================================================

# tongsuossl 命令路径（如果不在 PATH 中，请修改为完整路径）
TONGSUOSSL=${TONGSUOSSL:-tongsuo}

# DN 基础信息
C="CN"
ST="Xinjiang"
L="Urumqi"
O_ROOT="The Self-Signed Certificate Authority"
O_INTERMEDIATE="The Self-Signed Intermediate Certificate Authority"
OU_INTERMEDIATE="Intermediate Certificate Authority"

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
NC='\033[0m' # No Color

info()    { printf "${BLUE}[INFO]${NC} %s\n" "$1"; }
success() { printf "${GREEN}[OK]${NC}   %s\n" "$1"; }
warn()    { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
error()   { printf "${RED}[ERROR]${NC} %s\n" "$1"; exit 1; }

# ============================================================
# 环境检查
# ============================================================

check_env() {
    info "检查铜锁环境..."

    if ! command -v "$TONGSUOSSL" >/dev/null 2>&1; then
        error "未找到 $TONGSUOSSL 命令，请确认铜锁已安装并配置 PATH"
    fi

    version=$("$TONGSUOSSL" version 2>&1 | head -1)
    success "铜锁版本: $version"

    # 测试 SM2 密钥生成
    info "测试 SM2 算法支持..."
    "$TONGSUOSSL" genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out /tmp/tongsuo_test.key 2>/dev/null \
        || error "SM2 算法不支持，请检查铜锁安装"
    rm -f /tmp/tongsuo_test.key
    success "SM2 算法支持正常"
}

# ============================================================
# 清理旧文件
# ============================================================

cleanup() {
    info "清理旧的证书文件..."
    rm -f *.key *.csr *.crt *.srl *-chain.crt
    success "清理完成"
}

# ============================================================
# 生成 SM2 私钥
# ============================================================

gen_sm2_key() {
    outfile="$1"
    "$TONGSUOSSL" genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out "$outfile"
}

# ============================================================
# 第一层: 根证书 (root)
# ============================================================

gen_root() {
    info "========== 第一层: 生成根证书 (root) =========="

    cn="Self-Signed Root Certificate Authority"
    subj="/C=${C}/ST=${ST}/L=${L}/O=${O_ROOT}/CN=${cn}"

    info "生成根证书私钥..."
    gen_sm2_key root.key

    info "生成自签名根证书 (有效期 ${DAYS_ROOT} 天)..."
    "$TONGSUOSSL" req -new -x509 -nodes -key root.key -sm3 -days "$DAYS_ROOT" \
        -out root.crt -config root.cnf -subj "$subj"

    success "根证书生成完成: root.crt"
}

# ============================================================
# 第二层: 中间证书 (intermediate)
# ============================================================

gen_intermediate() {
    info "========== 第二层: 生成中间证书 (intermediate) =========="

    cn="Self-Signed Intermediate Certificate Authority"
    subj="/C=${C}/ST=${ST}/L=${L}/O=${O_INTERMEDIATE}/OU=${OU_INTERMEDIATE}/CN=${cn}"

    info "生成中间证书私钥..."
    gen_sm2_key intermediate.key

    info "生成中间证书 CSR..."
    "$TONGSUOSSL" req -config intermediate.cnf -new -key intermediate.key \
        -out intermediate.csr -sm3 -nodes -subj "$subj"

    info "使用根证书签发中间证书 (有效期 ${DAYS_INTERMEDIATE} 天)..."
    "$TONGSUOSSL" x509 -req -in intermediate.csr -CA root.crt -CAkey root.key \
        -CAcreateserial -out intermediate.crt -days "$DAYS_INTERMEDIATE" -sm3 \
        -extfile intermediate.cnf -extensions v3_intermediate_ca

    success "中间证书生成完成: intermediate.crt"
}

# ============================================================
# 第三层: 应用级 CA
# ============================================================

gen_authority() {
    name="$1"      # authority 名称, 如 http
    prefix="$2"    # 前缀, 如 h (用于实体证书命名)
    display="$3"   # 显示名, 如 HTTP

    info "========== 第三层: 生成 ${display}-authority =========="

    cn="Self-Signed ${display} Certificate Authority"
    o="The Self-Signed ${display} Certificate Authority"
    ou="${display} Certificate Authority"
    subj="/C=${C}/ST=${ST}/L=${L}/O=${o}/OU=${ou}/CN=${cn}"

    cnf_file="${name}-authority.cnf"
    key_file="${name}-authority.key"
    csr_file="${name}-authority.csr"
    crt_file="${name}-authority.crt"

    info "生成 ${name}-authority 私钥..."
    gen_sm2_key "$key_file"

    info "生成 ${name}-authority CSR..."
    "$TONGSUOSSL" req -config "$cnf_file" -new -key "$key_file" \
        -out "$csr_file" -sm3 -nodes -subj "$subj"

    info "使用中间证书签发 ${name}-authority (有效期 ${DAYS_AUTHORITY} 天)..."
    "$TONGSUOSSL" x509 -req -in "$csr_file" -CA intermediate.crt -CAkey intermediate.key \
        -CAcreateserial -out "$crt_file" -days "$DAYS_AUTHORITY" -sm3 \
        -extfile "$cnf_file" -extensions v3_application_ca

    success "${display}-authority 生成完成: ${crt_file}"
}

# ============================================================
# 第四层: 实体双证书
# ============================================================

gen_entity_cert() {
    authority_name="$1"   # 签发此证书的 authority 名称, 如 http
    entity_prefix="$2"    # 实体证书前缀, 如 hserver
    cert_type="$3"        # sign 或 enc
    role="$4"             # server 或 client
    cn="$5"               # CommonName
    o="$6"                # Organization
    ou="$7"               # OrganizationalUnit

    cnf_file="${entity_prefix}_${cert_type}.cnf"
    key_file="${entity_prefix}_${cert_type}.key"
    csr_file="${entity_prefix}_${cert_type}.csr"
    crt_file="${entity_prefix}_${cert_type}.crt"
    ext_section="${role}_${cert_type}_req"

    subj="/C=${C}/ST=${ST}/L=${L}/O=${o}/OU=${ou}/CN=${cn}"

    info "生成 ${entity_prefix}_${cert_type} 私钥..."
    gen_sm2_key "$key_file"

    info "生成 ${entity_prefix}_${cert_type} CSR..."
    "$TONGSUOSSL" req -config "$cnf_file" -new -key "$key_file" \
        -out "$csr_file" -sm3 -nodes -subj "$subj"

    info "使用 ${authority_name}-authority 签发 ${entity_prefix}_${cert_type} (有效期 ${DAYS_ENTITY} 天)..."
    "$TONGSUOSSL" x509 -req -in "$csr_file" \
        -CA "${authority_name}-authority.crt" \
        -CAkey "${authority_name}-authority.key" \
        -CAcreateserial -out "$crt_file" -days "$DAYS_ENTITY" -sm3 \
        -extfile "$cnf_file" -extensions "$ext_section"

    success "${entity_prefix}_${cert_type} 证书生成完成: ${crt_file}"
}

gen_all_entity_certs() {
    auth_name="$1"    # authority 名称, 如 http
    prefix="$2"       # 前缀, 如 h
    display="$3"      # 显示名, 如 HTTP
    server_cn="$4"    # 服务器证书 CN
    client_cn="$5"    # 客户端证书 CN
    o="$6"            # Organization
    ou="$7"           # OrganizationalUnit

    info "========== 第四层: 生成 ${display} 实体双证书 =========="

    # 服务器签名证书
    gen_entity_cert "$auth_name" "${prefix}server" "sign" "server" "$server_cn" "$o" "$ou"

    # 服务器加密证书
    gen_entity_cert "$auth_name" "${prefix}server" "enc" "server" "$server_cn" "$o" "$ou"

    # 客户端签名证书
    gen_entity_cert "$auth_name" "${prefix}client" "sign" "client" "$client_cn" "$o" "$ou"

    # 客户端加密证书
    gen_entity_cert "$auth_name" "${prefix}client" "enc" "client" "$client_cn" "$o" "$ou"

    success "${display} 实体双证书全部生成完成"
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

    if "$TONGSUOSSL" verify -CAfile "$cafile" "$cert" >/dev/null 2>&1; then
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

    # 验证中间证书
    verify_one root.crt intermediate.crt "intermediate (由 root 签发)"

    # 验证各 authority 证书
    for auth in http grpc docker applications libp2p; do
        verify_one root-intermediate-chain.crt "${auth}-authority.crt" "${auth}-authority"
    done

    # 验证各实体证书 (auth:prefix 配对循环, 兼容无关联数组的 POSIX sh)
    for ap in http:h grpc:g docker:d applications:a libp2p:l; do
        auth=${ap%%:*}
        prefix=${ap#*:}
        for role in server client; do
            for type in sign enc; do
                verify_one "${auth}-full-chain.crt" "${prefix}${role}_${type}.crt" "${prefix}${role}_${type}"
            done
        done
    done

    echo ""
    info "验证结果: 通过 ${passed}, 失败 ${failed}"

    if [ "$failed" -gt 0 ]; then
        error "有 ${failed} 个证书验证失败，请检查"
    else
        success "所有证书验证通过！"
    fi
}

# ============================================================
# 统计输出
# ============================================================

show_summary() {
    echo ""
    echo "============================================================"
    echo "  铜锁国密双证书生成完成"
    echo "============================================================"
    echo ""
    echo "  证书层级结构:"
    echo "    root (根CA)"
    echo "    └── intermediate (中间CA)"
    echo "        ├── http-authority"
    echo "        │   ├── hserver_sign.crt / hserver_enc.crt"
    echo "        │   └── hclient_sign.crt / hclient_enc.crt"
    echo "        ├── grpc-authority"
    echo "        │   ├── gserver_sign.crt / gserver_enc.crt"
    echo "        │   └── gclient_sign.crt / gclient_enc.crt"
    echo "        ├── docker-authority"
    echo "        │   ├── dserver_sign.crt / dserver_enc.crt"
    echo "        │   └── dclient_sign.crt / dclient_enc.crt"
    echo "        ├── applications-authority"
    echo "        │   ├── aserver_sign.crt / aserver_enc.crt"
    echo "        │   └── aclient_sign.crt / aclient_enc.crt"
    echo "        └── libp2p-authority"
    echo "            ├── lserver_sign.crt / lserver_enc.crt"
    echo "            └── lclient_sign.crt / lclient_enc.crt"
    echo ""

    key_count=$(ls *.key 2>/dev/null | wc -l)
    crt_count=$(ls *.crt 2>/dev/null | wc -l)
    cnf_count=$(ls *.cnf 2>/dev/null | wc -l)

    echo "  文件统计:"
    echo "    配置文件 (.cnf): ${cnf_count} 个"
    echo "    私钥文件 (.key): ${key_count} 个"
    echo "    证书文件 (.crt): ${crt_count} 个"
    echo ""
    echo "  算法: SM2 (密钥) + SM3 (摘要)"
    echo "  模式: 双证书 (签名证书 + 加密证书)"
    echo ""
    echo "============================================================"
}

# ============================================================
# 主流程
# ============================================================

main() {
    echo ""
    echo "============================================================"
    echo "  铜锁国密双证书四层结构一键生成脚本"
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
    gen_authority "http" "h" "HTTP"
    gen_authority "grpc" "g" "GRPC"
    gen_authority "docker" "d" "DOCKER"
    gen_authority "applications" "a" "APPLICATIONS"
    gen_authority "libp2p" "l" "LIBP2P"

    # 6. 第四层: 实体双证书
    gen_all_entity_certs "http" "h" "HTTP" \
        "www.jianlu.site" "hclient.jianlu.site" \
        "HTTP Certificate Authority" "HTTP Certificate"

    gen_all_entity_certs "grpc" "g" "GRPC" \
        "gserver.jianlu.site" "gclient.jianlu.site" \
        "GRPC Certificate Authority" "GRPC Certificate"

    gen_all_entity_certs "docker" "d" "DOCKER" \
        "dserver.jianlu.site" "dclient.jianlu.site" \
        "DOCKER Certificate Authority" "DOCKER Certificate"

    gen_all_entity_certs "applications" "a" "APPLICATIONS" \
        "aserver.jianlu.site" "aclient.jianlu.site" \
        "Applications Certificate Authority" "Applications Certificate"

    gen_all_entity_certs "libp2p" "l" "LIBP2P" \
        "lserver.jianlu.site" "lclient.jianlu.site" \
        "LIBP2P Certificate Authority" "LIBP2P Certificate"

    # 7. 生成证书链
    gen_chains

    # 8. 验证证书链
    verify_certs

    # 9. 输出统计
    show_summary
}

# 执行主流程
main
