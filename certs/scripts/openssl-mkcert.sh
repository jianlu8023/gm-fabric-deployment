#!/bin/sh
#
# OpenSSL RSA 证书四层结构一键生成脚本 (自包含版)
# 证书层级: root -> intermediate -> {http,grpc,docker,applications,libp2p}-authority -> 实体证书
# 使用 openssl x509 -req 方式签发，RSA 密钥 + SHA256 摘要，单证书模式
#
# 自包含: 全部 cnf 模板内嵌于本脚本, 部署到新环境只需拷贝本文件即可生成
# 全套证书。已有 cnf 默认保留 (便于手改 SAN), 加 --reset-cnf 强制按模板重建。
# (renew.sh 读取的也是磁盘上的同名 cnf, 本脚本运行后会落盘)
#
# 依赖: openssl (1.1.x 或 3.x)
# 用法: chmod +x mkcert.sh && ./mkcert.sh [--reset-cnf] [-h]
#
# 兼容性: 纯 POSIX sh, 可用 bash / dash / busybox ash 运行
#

set -e

# 脚本固定在自身所在目录执行, 产物与 cnf 模板均在此目录
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd) || exit 1
cd "$SCRIPT_DIR" || exit 1

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

# 实体 SAN 模板参数
BASE_DOMAIN="jianlu.site"
SAN_IPS="127.0.0.1 192.168.58.110"

# 是否强制重建实体 cnf (由 --reset-cnf 置位; 默认保留已有 cnf, 便于手改 SAN)
RESET_CNF=0

# 算法: rsa (默认) 或 ecdsa (由 --ecdsa 置位)
ALGO=rsa
# ECDSA 曲线 (由 --curve 覆盖); prime256v1 即 NIST P-256, 兼容性最好
EC_CURVE=prime256v1
# 签名摘要: P-384 及以上自动用 SHA384, 其余 SHA256
compute_md() {
    case "$EC_CURVE" in
        secp384r1|P-384|p384|prime384v1) MD=sha384 ;;
        secp521r1|P-521|p521|prime521v1) MD=sha512 ;;
        *) MD=sha256 ;;
    esac
}
MD=sha256
ALGO_DESC="RSA 4096/2048 (密钥) + SHA256 (签名)"

# 统一密钥生成入口: 按 ALGO 分派 (RSA 2048 实体 / 4096 CA, EC 按曲线)
gen_key() {
    outfile="$1"
    size="$2"   # RSA: bits (2048/4096); EC: 忽略, 用全局曲线
    if [ "$ALGO" = "ecdsa" ]; then
        "$OPENSSL" ecparam -name "$EC_CURVE" -genkey -noout -out "$outfile"
    else
        "$OPENSSL" genrsa -out "$outfile" "$size"
    fi
}

key_desc() {
    if [ "$ALGO" = "ecdsa" ]; then
        echo "EC $EC_CURVE"
    else
        echo "RSA ${1}位"
    fi
}

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
# 内嵌 cnf 模板 (自包含: 新环境无需拷贝任何 .cnf)
# ============================================================

# 应用组注册表: 组名|前缀|显示名|OU基名|服务器CN|客户端CN
# (applications 组只建 authority, 不出实体证书, 与历史行为一致)
app_reg() {
    cat <<'EOF'
http|h|HTTP|HTTP|www.jianlu.site|hclient.jianlu.site
grpc|g|GRPC|Grpc|gserver.jianlu.site|gclient.jianlu.site
docker|d|DOCKER|Docker|dserver.jianlu.site|dclient.jianlu.site
applications|a|APPLICATIONS|Applications|aserver.jianlu.site|aclient.jianlu.site
libp2p|l|LIBP2P|Libp2p|lserver.jianlu.site|lclient.jianlu.site
EOF
}

lookup_group() {
    app_reg | awk -F'|' -v g="$1" '$1==g{print; exit}'
}

# 实体 cnf. SAN 规则: hserver 无自身 FQDN (由 *.jianlu.site 覆盖),
# 其余实体优先放自身 FQDN; 之后是通配/主域/localhost + IP.
# 扩展段名保持历史命名 (v3_http_cert / v3_{group}_{role}_cert), renew.sh 兼容.
write_entity_cnf() {
    ent="$1"      # 实体名, 如 hserver
    group="$2"    # 组名, 如 http
    display="$3"  # 显示名, 如 GRPC (用于 organizationName)
    ou="$4"       # OU 基名, 如 Grpc
    role="$5"     # server 或 client
    cn="$6"       # CommonName
    ext="$7"      # 扩展段名

    cnf="${ent}.cnf"
    if [ -f "$cnf" ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi

    # keyUsage / EKU: server 带 keyEncipherment; client 仅 digitalSignature;
    # http server 沿用历史写法 EKU 含 clientAuth
    ku="digitalSignature"
    eku="clientAuth"
    if [ "$role" = "server" ]; then
        ku="digitalSignature, keyEncipherment"
        eku="serverAuth"
        [ "$group" = "http" ] && eku="serverAuth, clientAuth"
    fi

    # OU: http 组为 "HTTP Certificate", 其余为 "{ou} {Role} Certificate"
    ou_use="${ou} Certificate"
    [ "$group" != "http" ] && {
        case "$role" in
            server) ou_use="${ou} Server Certificate" ;;
            client) ou_use="${ou} Client Certificate" ;;
        esac
    }

    # SAN DNS 列表
    san_dns=""
    [ "$ent" != "hserver" ] && san_dns="${cn}
"
    san_dns="${san_dns}*.${BASE_DOMAIN}
${BASE_DOMAIN}
localhost"

    {
        printf '# %s  (由 mkcert.sh 内嵌模板生成, 可手改后重签)\n' "$cnf"
        cat <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
countryName = CN
stateOrProvinceName = Xinjiang
localityName = Urumqi
organizationName = ${display} Certificate Authority
organizationalUnitName = ${ou_use}
commonName = ${cn}

[v3_req]
subjectAltName = @alt_names

[alt_names]
EOF
        n=1
        for d in $san_dns; do
            printf 'DNS.%s = %s\n' "$n" "$d"; n=$((n+1))
        done
        n=1
        for ip in $SAN_IPS; do
            printf 'IP.%s = %s\n' "$n" "$ip"; n=$((n+1))
        done
        cat <<EOF

[${ext}]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = critical, ${ku}
extendedKeyUsage = ${eku}
subjectAltName = @alt_names
EOF
    } > "$cnf"
    success "生成配置文件: $cnf"
}

write_root_cnf() {
    [ -f root.cnf ] && [ "$RESET_CNF" -ne 1 ] && return 0
    cat > root.cnf <<'EOF'
# openssl-root.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_ca

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed Certificate Authority
CN = Self-Signed Root Certificate Authority

[ v3_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE, pathlen:2
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
EOF
    success "生成配置文件: root.cnf"
}

write_intermediate_cnf() {
    [ -f intermediate.cnf ] && [ "$RESET_CNF" -ne 1 ] && return 0
    cat > intermediate.cnf <<'EOF'
# openssl-intermediate.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_intermediate_ca

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed Intermediate Certificate Authority
OU = Intermediate Certificate Authority
CN = Self-Signed Intermediate Certificate Authority

[ v3_intermediate_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE, pathlen:1
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
EOF
    success "生成配置文件: intermediate.cnf"
}

write_authority_cnf() {
    name="$1"      # 组名, 如 http
    display="$2"   # 显示名, 如 HTTP

    cnf="${name}-authority.cnf"
    if [ -f "$cnf" ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi
    cat > "$cnf" <<EOF
# openssl-${name}-authority.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_application_ca

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed ${display} Certificate Authority
OU = ${display} Certificate Authority
CN = Self-Signed ${display} Certificate Authority

[ v3_application_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
extendedKeyUsage = serverAuth, clientAuth
EOF
    success "生成配置文件: $cnf"
}

# 落盘全部 cnf (已有默认保留, --reset-cnf 强制重建)
ensure_cnfs() {
    info "检查/生成 cnf 配置模板..."
    write_root_cnf
    write_intermediate_cnf
    for line in $(app_reg); do
        IFS='|' read -r gname prefix display oubase srv_cn cli_cn <<EOF
$line
EOF
        write_authority_cnf "$gname" "$display"
    done
    # 实体 cnf: http/grpc/docker/libp2p 四组 (applications 仅 CA, 无实体)
    # 参数: 实体名|组名|显示名|OU基名|角色|CN|扩展段名
    write_entity_cnf hserver http HTTP HTTP server "www.jianlu.site" v3_http_cert
    write_entity_cnf hclient http HTTP HTTP client hclient.jianlu.site v3_http_cert
    write_entity_cnf gserver grpc GRPC Grpc server gserver.jianlu.site v3_grpc_server_cert
    write_entity_cnf gclient grpc GRPC Grpc client gclient.jianlu.site v3_grpc_client_cert
    write_entity_cnf dserver docker DOCKER Docker server dserver.jianlu.site v3_docker_server_cert
    write_entity_cnf dclient docker DOCKER Docker client dclient.jianlu.site v3_docker_client_cert
    write_entity_cnf lserver libp2p LIBP2P Libp2p server lserver.jianlu.site v3_libp2p_server_cert
    write_entity_cnf lclient libp2p LIBP2P Libp2p client lclient.jianlu.site v3_libp2p_client_cert
}

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
# 第一层: 根证书 (root)
# ============================================================

gen_root() {
    info "========== 第一层: 生成根证书 (root) =========="

    info "生成根证书私钥 ($(key_desc 4096))..."
    gen_key root.key 4096

    info "生成自签名根证书 (有效期 ${DAYS_ROOT} 天)..."
    "$OPENSSL" req -new -x509 -nodes -key root.key "-$MD" -days "$DAYS_ROOT" \
        -out root.crt -config root.cnf

    success "根证书生成完成: root.crt"
}

# ============================================================
# 第二层: 中间证书 (intermediate)
# ============================================================

gen_intermediate() {
    info "========== 第二层: 生成中间证书 (intermediate) =========="

    info "生成中间证书私钥 ($(key_desc 4096))..."
    gen_key intermediate.key 4096

    info "生成中间证书 CSR..."
    "$OPENSSL" req -new -key intermediate.key "-$MD" -out intermediate.csr -config intermediate.cnf

    info "使用根证书签发中间证书 (有效期 ${DAYS_INTERMEDIATE} 天)..."
    "$OPENSSL" x509 -req -in intermediate.csr -CA root.crt -CAkey root.key \
        -CAcreateserial -out intermediate.crt -days "$DAYS_INTERMEDIATE" "-$MD" \
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

    info "生成 ${name}-authority 私钥 ($(key_desc 4096))..."
    gen_key "$key_file" 4096

    info "生成 ${name}-authority CSR..."
    "$OPENSSL" req -new -key "$key_file" "-$MD" -out "$csr_file" -config "$cnf_file"

    info "使用中间证书签发 ${name}-authority (有效期 ${DAYS_AUTHORITY} 天)..."
    "$OPENSSL" x509 -req -in "$csr_file" -CA intermediate.crt -CAkey intermediate.key \
        -CAcreateserial -out "$crt_file" -days "$DAYS_AUTHORITY" "-$MD" \
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

    info "生成 ${desc} 私钥 ($(key_desc 2048))..."
    gen_key "$key_file" 2048

    info "生成 ${desc} CSR..."
    "$OPENSSL" req -new -key "$key_file" "-$MD" -out "$csr_file" -config "$cnf_file"

    info "使用 ${authority_name}-authority 签发 ${desc} (有效期 ${DAYS_ENTITY} 天)..."
    "$OPENSSL" x509 -req -in "$csr_file" \
        -CA "${authority_name}-authority.crt" \
        -CAkey "${authority_name}-authority.key" \
        -CAcreateserial -out "$crt_file" -days "$DAYS_ENTITY" "-$MD" \
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
    echo "  OpenSSL ${ALGO} 证书四层结构生成完成"
    echo "============================================================"
    echo ""
    echo "  证书层级结构:"
    echo "    root (根CA, $(key_desc 4096))"
    echo "    └── intermediate (中间CA, $(key_desc 4096))"
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
    echo "  算法: ${ALGO_DESC}"
    echo "  模式: 单证书 (签名和加密共用)"
    echo ""
    echo "============================================================"
}

# ============================================================
# 帮助
# ============================================================

usage() {
    cat <<'EOF'
OpenSSL 证书四层结构一键生成脚本 (自包含版, cnf 模板内嵌, RSA/ECDSA 双算法)

用法:
  ./mkcert.sh [选项]

选项:
  (无)              全新生成: CA 层 + 四组实体证书 (默认 RSA 算法)
  --ecdsa           切换为 ECDSA 椭圆曲线算法 (默认曲线 prime256v1/P-256)
  --curve <name>    指定 EC 曲线: prime256v1(默认)/secp384r1/secp521r1;
                    P-384/P-521 自动配 SHA384/SHA512 签名
  --cnf-only        只落盘/重建 cnf 模板, 不签发任何证书 (供手改 SAN)
  --reset-cnf       强制按内嵌模板重建全部 cnf (手改过的 SAN 会丢失)
  -h, --help        显示本帮助

算法说明:
  - 文件名/cnf/扩展段在两种算法下完全相同, 同一目录只能存在一套产物;
    如需同时保留 RSA 和 ECDSA 两套, 把脚本拷到不同目录分别运行
  - ECDSA 密钥更短、签名更快, 是当前 TLS 证书的主流选择;
    RSA 兼容性最广 (老旧客户端), 按需选择
  - 续签: renew.sh 沿用已有私钥, 无需感知算法; 只有全新生成时才选算法

说明:
  - 新环境只需拷贝本文件, 运行时自动生成全部所需 .cnf 到脚本所在目录
  - 已有 .cnf 默认保留 (便于手改 SAN), 可手改后重跑本脚本重签
  - openssl genrsa/ecparam 输出到 stderr 属正常现象
EOF
}

# ============================================================
# 主流程
# ============================================================

main() {
    CNF_ONLY=0
    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help) usage; exit 0 ;;
            --ecdsa) ALGO=ecdsa ;;
            --rsa) ALGO=rsa ;;
            --curve)
                shift
                [ $# -eq 0 ] && error "--curve 需要一个参数 (如 prime256v1)"
                EC_CURVE="$1"
                ;;
            --curve=*) EC_CURVE=${1#--curve=} ;;
            --reset-cnf) RESET_CNF=1 ;;
            --cnf-only) CNF_ONLY=1 ;;
            *) error "未知参数: $1 (用 -h 查看帮助)" ;;
        esac
        shift
    done

    if [ "$ALGO" = "ecdsa" ]; then
        compute_md
        ALGO_DESC="EC ${EC_CURVE} (密钥) + ${MD} (签名)"
        # 校验曲线是否受支持
        if ! "$OPENSSL" ecparam -name "$EC_CURVE" >/dev/null 2>&1; then
            error "未知曲线: $EC_CURVE (可选 prime256v1 / secp384r1 / secp521r1)"
        fi
    fi

    echo ""
    echo "============================================================"
    echo "  OpenSSL 证书四层结构一键生成脚本 (${ALGO})"
    echo "============================================================"
    echo ""

    # 1. 落盘 cnf 模板 (缺失时自动生成, 已有默认保留)
    ensure_cnfs
    if [ "$CNF_ONLY" -eq 1 ]; then
        success "cnf-only 模式: 配置模板已就绪, 未签发任何证书"
        exit 0
    fi

    # 2. 环境检查
    check_env

    # 3. 清理旧文件
    cleanup

    # 4. 第一层: 根证书
    gen_root

    # 5. 第二层: 中间证书
    gen_intermediate

    # 6. 第三层: 应用级 CA (5个)
    gen_authority "http" "HTTP"
    gen_authority "grpc" "GRPC"
    gen_authority "docker" "DOCKER"
    gen_authority "applications" "APPLICATIONS"
    gen_authority "libp2p" "LIBP2P"

    # 7. 第四层: 实体证书
    gen_http_certs
    gen_grpc_certs
    gen_docker_certs
    gen_libp2p_certs

    # 8. 生成证书链
    gen_chains

    # 9. 验证证书链
    verify_certs

    # 10. 显示过期信息
    show_expiry_info

    # 11. 输出统计
    show_summary
}

# 执行主流程
main "$@"
