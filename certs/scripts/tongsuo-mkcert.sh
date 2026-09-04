#!/bin/sh
#
# 铜锁国密证书一键生成脚本 (合并版): GM/T0024 双证书 + RFC8998 单证书 (POSIX sh)
# 证书层级: root -> intermediate -> {http,grpc,docker,applications,libp2p}-authority
#           -> 实体 (每个实体 3 张叶子: *_sign + *_enc 双证书 + 无后缀 RFC8998 单证书)
# 使用 tongsuossl x509 -req 方式签发
#
# SM2 标识强制 (重要):
#   所有 SM2 签名操作显式强制 sm2_id=1234567812345678 (GM/T 与 Go gmsm / 原生
#   OpenSSL 的标准默认值), 保证整条证书链可被 Go gmsm / 原生 OpenSSL 校验。
#   铜锁 CLI (req/x509 -req) 的默认 distid 并非该值, 必须显式传参。
#
# 一套 CA 两种叶子 (合并生成):
#   CA 层只生成一份, 每个实体叶子三种形态:
#     - GM/T0024 双证书: {prefix}server_sign.crt + {prefix}server_enc.crt
#       (NTLS/双向认证场景, 签名与加密密钥分离)
#     - RFC8998 单证书:  {prefix}server.crt (sign+enc 合一)
#       (TLS1.3 + TLS_SM4_GCM_SM3/CCM_SM3 场景, 密钥协商走 ECDHE key_share)
#
# 依赖: tongsuossl (Tongsuo 8.x)
# 用法: chmod +x mkcert.sh && ./mkcert.sh
#   ./mkcert.sh                 全新生成 (CA + 全部实体双证书与单证书)
#   ./mkcert.sh --keep-ca-keys  保留 CA 私钥, 仅重签 CA 证书 + 全部实体
#   ./mkcert.sh --resign-ca     保留 CA 私钥, 仅重签 CA 层证书 (不动实体)
#   ./mkcert.sh --reset-cnf     强制按模板重建全部实体 cnf (手改过的内容会丢失)
#   ./mkcert.sh --cnf-only      只落盘/重建 cnf 模板, 不签发任何证书 (供手改 SAN)
#
# 自包含: 全部 cnf 模板 (CA 层 + 双证书 + 单证书) 内嵌于本脚本, 部署到新环境
# 只需拷贝本文件即可生成全套证书; 已有 cnf 默认保留 (便于手改 SAN)。
# (renew.sh 读取的也是磁盘上的同名 cnf, 本脚本运行后会落盘)
#
# 已部署的旧叶子证书用非标准 ID 签名, 重签 CA 后需续签才能被标准 ID 校验
#   (./renew.sh renew-all)
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

# tongsuossl 命令路径（如果不在 PATH 中，请修改为完整路径）
TONGSUOSSL=${TONGSUOSSL:-tongsuo}

# SM2 签名者标识: GM/T 标准默认值, 所有签名操作显式强制
SM2_ID=${SM2_ID:-1234567812345678}
# SM2_SIGOPT: auto(默认, tongsuo 强制 / 原生 openssl 关闭) | 1(强制) | 0(关闭)
SM2_SIGOPT=${SM2_SIGOPT:-auto}

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

# 实体单证书 (RFC8998) SAN 模板参数
BASE_DOMAIN="jianlu.site"
SAN_IPS="127.0.0.1 192.168.58.110"

# 是否强制重建实体 cnf (由 --reset-cnf 置位; 默认保留已有 cnf, 便于手改 SAN)
RESET_CNF=0

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
# 内嵌 cnf 模板 (自包含: 新环境无需拷贝任何 .cnf)
# ============================================================

# 实体双证书 cnf (GM/T0024). SAN 规则: hserver_sign/hserver_enc 无自身 FQDN
# (由 *.jianlu.site 覆盖), 其余实体优先放自身 FQDN.
# 扩展段名 {role}_{type}_req 与 gen_entity_cert 的签发参数保持一致.
write_double_cnf() {
    ent="$1"      # 实体名, 如 hserver
    ctype="$2"    # sign 或 enc
    role="$3"     # server 或 client
    cn="$4"       # CommonName

    cnf_file="${ent}_${ctype}.cnf"
    if [ -f "$cnf_file" ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi

    ext_section="${role}_${ctype}_req"
    # sign: 签名用途; enc: 密钥交换/加密用途 (GM/T0024 分工)
    ku="nonRepudiation, digitalSignature"
    [ "$ctype" = "enc" ] && ku="keyAgreement, keyEncipherment, dataEncipherment"
    eku="serverAuth"
    [ "$role" = "client" ] && eku="clientAuth"

    san_dns=""
    [ "$ent" != "hserver" ] && san_dns="${cn}
"
    san_dns="${san_dns}*.${BASE_DOMAIN}
${BASE_DOMAIN}
localhost"

    {
        printf '# %s  (由 mkcert.sh 内嵌模板生成, 可手改后重签)\n' "$cnf_file"
        cat <<'EOF'
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only
default_md          = sha256
req_extensions      = v3_req

[ req_distinguished_name ]
countryName                     = optional
stateOrProvinceName             = optional
localityName                    = optional
0.organizationName              = optional
organizationalUnitName          = optional
commonName                      = optional
emailAddress                    = optional

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[ alt_names ]
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

[ $ext_section ]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = $ku
subjectAltName = @alt_names
extendedKeyUsage = $eku
EOF
    } > "$cnf_file"
    success "生成配置文件: $cnf_file"
}

write_root_cnf() {
    if [ -f root.cnf ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi
    cat > root.cnf <<'EOF'
# tongsuo-root.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits        = 2048
default_md          = sha256
distinguished_name  = req_distinguished_name
string_mask         = utf8only
x509_extensions     = v3_ca

[ req_distinguished_name ]
countryName                     = optional
stateOrProvinceName             = optional
localityName                    = optional
0.organizationName              = optional
organizationalUnitName          = optional
commonName                      = optional
emailAddress                    = optional

[ v3_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:2
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
EOF
    success "生成配置文件: root.cnf"
}

write_intermediate_cnf() {
    if [ -f intermediate.cnf ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi
    cat > intermediate.cnf <<'EOF'
# tongsuo-intermediate.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits        = 2048
default_md          = sha256
distinguished_name  = req_distinguished_name
string_mask         = utf8only

[ req_distinguished_name ]
countryName                     = optional
stateOrProvinceName             = optional
localityName                    = optional
0.organizationName              = optional
organizationalUnitName          = optional
commonName                      = optional
emailAddress                    = optional

[ v3_intermediate_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:1
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
# tongsuo-${name}-authority.cnf (由 mkcert.sh 内嵌模板生成)
[ req ]
default_bits        = 2048
default_md          = sha256
distinguished_name  = req_distinguished_name
string_mask         = utf8only

[ req_distinguished_name ]
countryName                     = optional
stateOrProvinceName             = optional
localityName                    = optional
0.organizationName              = optional
organizationalUnitName          = optional
commonName                      = optional
emailAddress                    = optional

[ v3_application_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
extendedKeyUsage = serverAuth, clientAuth
EOF
    success "生成配置文件: $cnf"
}

# 落盘全部 cnf (CA 层 + 双证书实体 + RFC8998 单证书; 已有默认保留, --reset-cnf 强制重建)
ensure_cnfs() {
    info "检查/生成 cnf 配置模板..."
    write_root_cnf
    write_intermediate_cnf
    for ap in http:h:HTTP grpc:g:GRPC docker:d:DOCKER applications:a:APPLICATIONS libp2p:l:LIBP2P; do
        auth=${ap%%:*}
        rest=${ap#*:}
        prefix=${rest%%:*}
        display=${rest#*:}
        write_authority_cnf "$auth" "$display"
    done
    # 双证书实体 cnf
    write_double_cnf hserver sign server "www.jianlu.site"
    write_double_cnf hserver enc server "www.jianlu.site"
    write_double_cnf hclient sign client "hclient.jianlu.site"
    write_double_cnf hclient enc client "hclient.jianlu.site"
    write_double_cnf gserver sign server "gserver.jianlu.site"
    write_double_cnf gserver enc server "gserver.jianlu.site"
    write_double_cnf gclient sign client "gclient.jianlu.site"
    write_double_cnf gclient enc client "gclient.jianlu.site"
    write_double_cnf dserver sign server "dserver.jianlu.site"
    write_double_cnf dserver enc server "dserver.jianlu.site"
    write_double_cnf dclient sign client "dclient.jianlu.site"
    write_double_cnf dclient enc client "dclient.jianlu.site"
    write_double_cnf aserver sign server "aserver.jianlu.site"
    write_double_cnf aserver enc server "aserver.jianlu.site"
    write_double_cnf aclient sign client "aclient.jianlu.site"
    write_double_cnf aclient enc client "aclient.jianlu.site"
    write_double_cnf lserver sign server "lserver.jianlu.site"
    write_double_cnf lserver enc server "lserver.jianlu.site"
    write_double_cnf lclient sign client "lclient.jianlu.site"
    write_double_cnf lclient enc client "lclient.jianlu.site"
    # RFC8998 单证书 cnf (write_single_cnf 定义在下方)
    write_single_cnf hserver server "www.jianlu.site"
    write_single_cnf hclient client "hclient.jianlu.site"
    write_single_cnf gserver server "gserver.jianlu.site"
    write_single_cnf gclient client "gclient.jianlu.site"
    write_single_cnf dserver server "dserver.jianlu.site"
    write_single_cnf dclient client "dclient.jianlu.site"
    write_single_cnf aserver server "aserver.jianlu.site"
    write_single_cnf aclient client "aclient.jianlu.site"
    write_single_cnf lserver server "lserver.jianlu.site"
    write_single_cnf lclient client "lclient.jianlu.site"
}

# ============================================================
# SM2 标识强制参数 (使用处需不加引号展开以词分割)
# ============================================================

sm2_id_enabled() {
    case "$SM2_SIGOPT" in
        0) return 1 ;;
        1) return 0 ;;
        *) case "$TONGSUOSSL" in
               *openssl*) return 1 ;;
               *) return 0 ;;
           esac ;;
    esac
}

SIGOPT_ARGS=""
SM2ID_X509_ARGS=""
if sm2_id_enabled; then
    case "$TONGSUOSSL" in
        *openssl*)
            # 原生 openssl 参数名是 distid; 其 x509 -req 带 sigopt 有 CSR 自签校验缺陷, 默认不强制
            SIGOPT_ARGS="-sigopt distid:${SM2_ID}"
            SM2ID_X509_ARGS=""
            warn "原生 openssl: x509 -req 带 sigopt 有已知缺陷, 建议 SM2_SIGOPT=0"
            ;;
        *)
            SIGOPT_ARGS="-sigopt sm2_id:${SM2_ID}"
            SM2ID_X509_ARGS="-sm2-id ${SM2_ID}"
            info "SM2 标识强制: ${SM2_ID} (sigopt sm2_id + x509 -sm2-id)"
            ;;
    esac
else
    info "SM2 标识: 使用工具默认值 (${TONGSUOSSL} 对原生 openssl 默认即 ${SM2_ID})"
fi

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
    if [ "$RESIGN_CA_ONLY" -eq 1 ]; then
        info "resign-ca 模式: 保留全部现有文件, 仅重签 CA 层证书"
        return 0
    fi
    if [ "$KEEP_CA_KEYS" -eq 1 ]; then
        info "清理旧的证书文件 (保留 CA 私钥)..."
        for kf in *.key; do
            case "$kf" in
                root.key|intermediate.key|*-authority.key) ;;
                *) rm -f "$kf" ;;
            esac
        done
        rm -f *.csr *.crt *.srl *-chain.crt
    else
        info "清理旧的证书文件..."
        rm -f *.key *.csr *.crt *.srl *-chain.crt
    fi
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

    if [ "$KEEP_CA_KEYS" -eq 1 ] && [ -f root.key ]; then
        info "保留现有根证书私钥 root.key"
    else
        info "生成根证书私钥..."
        gen_sm2_key root.key
    fi

    info "生成自签名根证书 (有效期 ${DAYS_ROOT} 天, 强制 SM2 ID ${SM2_ID})..."
    "$TONGSUOSSL" req -new -x509 -nodes -key root.key -sm3 -days "$DAYS_ROOT" \
        -out root.crt -config root.cnf -subj "$subj" $SIGOPT_ARGS

    success "根证书生成完成: root.crt"
}

# ============================================================
# 第二层: 中间证书 (intermediate)
# ============================================================

gen_intermediate() {
    info "========== 第二层: 生成中间证书 (intermediate) =========="

    cn="Self-Signed Intermediate Certificate Authority"
    subj="/C=${C}/ST=${ST}/L=${L}/O=${O_INTERMEDIATE}/OU=${OU_INTERMEDIATE}/CN=${cn}"

    if [ "$KEEP_CA_KEYS" -eq 1 ] && [ -f intermediate.key ]; then
        info "保留现有中间证书私钥 intermediate.key"
    else
        info "生成中间证书私钥..."
        gen_sm2_key intermediate.key
    fi

    info "生成中间证书 CSR..."
    "$TONGSUOSSL" req -config intermediate.cnf -new -key intermediate.key \
        -out intermediate.csr -sm3 -nodes -subj "$subj" $SIGOPT_ARGS

    info "使用根证书签发中间证书 (有效期 ${DAYS_INTERMEDIATE} 天)..."
    "$TONGSUOSSL" x509 -req -in intermediate.csr -CA root.crt -CAkey root.key \
        -CAcreateserial -out intermediate.crt -days "$DAYS_INTERMEDIATE" -sm3 \
        -extfile intermediate.cnf -extensions v3_intermediate_ca \
        $SIGOPT_ARGS $SM2ID_X509_ARGS

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

    if [ "$KEEP_CA_KEYS" -eq 1 ] && [ -f "$key_file" ]; then
        info "保留现有 ${name}-authority 私钥"
    else
        info "生成 ${name}-authority 私钥..."
        gen_sm2_key "$key_file"
    fi

    info "生成 ${name}-authority CSR..."
    "$TONGSUOSSL" req -config "$cnf_file" -new -key "$key_file" \
        -out "$csr_file" -sm3 -nodes -subj "$subj" $SIGOPT_ARGS

    info "使用中间证书签发 ${name}-authority (有效期 ${DAYS_AUTHORITY} 天)..."
    "$TONGSUOSSL" x509 -req -in "$csr_file" -CA intermediate.crt -CAkey intermediate.key \
        -CAcreateserial -out "$crt_file" -days "$DAYS_AUTHORITY" -sm3 \
        -extfile "$cnf_file" -extensions v3_application_ca \
        $SIGOPT_ARGS $SM2ID_X509_ARGS

    success "${display}-authority 生成完成: ${crt_file}"
}

# ============================================================
# 第四层: 实体双证书 (GM/T0024: sign + enc 分离)
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
        -out "$csr_file" -sm3 -nodes -subj "$subj" $SIGOPT_ARGS

    info "使用 ${authority_name}-authority 签发 ${entity_prefix}_${cert_type} (有效期 ${DAYS_ENTITY} 天)..."
    "$TONGSUOSSL" x509 -req -in "$csr_file" \
        -CA "${authority_name}-authority.crt" \
        -CAkey "${authority_name}-authority.key" \
        -CAcreateserial -out "$crt_file" -days "$DAYS_ENTITY" -sm3 \
        -extfile "$cnf_file" -extensions "$ext_section" \
        $SIGOPT_ARGS $SM2ID_X509_ARGS

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

    # RFC8998 单证书 (server + client)
    gen_single_entity "$auth_name" "${prefix}server" "server" "$server_cn" "$o" "$ou"
    gen_single_entity "$auth_name" "${prefix}client" "client" "$client_cn" "$o" "$ou"

    success "${display} 实体证书全部生成完成 (双证书 + RFC8998 单证书)"
}

# ============================================================
# 第五层: 实体单证书 (RFC8998 / GM TLS 1.3: sign+enc 合一)
# ============================================================

# 单证书 cnf 模板. SAN 规则同现有 sign cnf: hserver 无自身 FQDN (由
# *.jianlu.site 覆盖), 其余实体优先放自身 FQDN; 之后是通配/主域/localhost + IP.
# 已有 cnf 默认保留 (便于手改 SAN), 加 --reset-cnf 强制按模板重建.
write_single_cnf() {
    ent="$1"     # 实体名, 如 hserver
    role="$2"    # server 或 client
    cn="$3"      # CommonName

    cnf_file="${ent}.cnf"
    if [ -f "$cnf_file" ] && [ "$RESET_CNF" -ne 1 ]; then
        return 0
    fi

    ext_section="server_ext"; eku="serverAuth, clientAuth"
    [ "$role" = "client" ] && { ext_section="client_ext"; eku="clientAuth"; }

    # SAN DNS 列表
    san_dns=""
    case "$role:$ent" in
        server:hserver) ;;                      # http server 无自身 FQDN
        *) san_dns="${cn}
" ;;
    esac
    san_dns="${san_dns}*.${BASE_DOMAIN}
${BASE_DOMAIN}
localhost"

    {
        printf '# %s  (RFC8998 / GM TLS 1.3 单证书配置, SM2+SM3, 与双证书共用 CA)\n' "$cnf_file"
        cat <<'EOF'
[ req ]
default_bits        = 2048
default_md          = sha256
distinguished_name  = req_distinguished_name
string_mask         = utf8only
req_extensions      = v3_req

[ req_distinguished_name ]
countryName                     = optional
stateOrProvinceName             = optional
localityName                    = optional
0.organizationName              = optional
organizationalUnitName          = optional
commonName                      = optional
emailAddress                    = optional

[ v3_req ]
basicConstraints = CA:FALSE
subjectAltName = @alt_names

[ alt_names ]
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

[ $ext_section ]
# RFC8998 §3.3.3: 证书必须具备数字签名能力 (digitalSignature)
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = $eku
subjectAltName = @alt_names
EOF
    } > "$cnf_file"
    success "生成配置文件: $cnf_file"
}

gen_single_entity() {
    authority_name="$1"   # 签发此证书的 authority 名称, 如 http
    entity_prefix="$2"    # 实体证书名, 如 hserver (无 sign/enc 后缀)
    role="$3"             # server 或 client
    cn="$4"               # CommonName
    o="$5"                # Organization
    ou="$6"               # OrganizationalUnit

    cnf_file="${entity_prefix}.cnf"
    key_file="${entity_prefix}.key"
    csr_file="${entity_prefix}.csr"
    crt_file="${entity_prefix}.crt"
    ext_section="server_ext"
    [ "$role" = "client" ] && ext_section="client_ext"

    subj="/C=${C}/ST=${ST}/L=${L}/O=${o}/OU=${ou}/CN=${cn}"

    info "生成 ${entity_prefix} (RFC8998 单证书, ${role}, CN=${cn}) 私钥..."
    gen_sm2_key "$key_file"

    write_single_cnf "$entity_prefix" "$role" "$cn"

    info "生成 ${entity_prefix} CSR..."
    "$TONGSUOSSL" req -config "$cnf_file" -new -key "$key_file" \
        -out "$csr_file" -sm3 -nodes -subj "$subj" $SIGOPT_ARGS

    info "使用 ${authority_name}-authority 签发 ${entity_prefix} (有效期 ${DAYS_ENTITY} 天)..."
    "$TONGSUOSSL" x509 -req -in "$csr_file" \
        -CA "${authority_name}-authority.crt" \
        -CAkey "${authority_name}-authority.key" \
        -CAcreateserial -out "$crt_file" -days "$DAYS_ENTITY" -sm3 \
        -extfile "$cnf_file" -extensions "$ext_section" \
        $SIGOPT_ARGS $SM2ID_X509_ARGS

    success "${entity_prefix} 证书生成完成: ${crt_file}"
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

    # 部署用实体链 (单证书叶子 + authority + intermediate)
    for ap in http:h grpc:g docker:d applications:a libp2p:l; do
        auth=${ap%%:*}
        prefix=${ap#*:}
        cat "${prefix}server.crt" "${auth}-authority.crt" intermediate.crt \
            > "${prefix}server-chain.crt"
        cat "${prefix}client.crt" "${auth}-authority.crt" intermediate.crt \
            > "${prefix}client-chain.crt"
        success "生成: ${prefix}server-chain.crt / ${prefix}client-chain.crt"
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
    if [ "$RESIGN_CA_ONLY" -ne 1 ]; then
        for ap in http:h grpc:g docker:d applications:a libp2p:l; do
            auth=${ap%%:*}
            prefix=${ap#*:}
            for role in server client; do
                for type in sign enc; do
                    verify_one "${auth}-full-chain.crt" "${prefix}${role}_${type}.crt" "${prefix}${role}_${type}"
                done
                # RFC8998 单证书
                verify_one "${auth}-full-chain.crt" "${prefix}${role}.crt" "${prefix}${role}"
            done
        done
    fi

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
    echo "  铜锁国密证书生成完成 (双证书 + RFC8998 单证书)"
    echo "============================================================"
    echo ""
    echo "  证书层级结构:"
    echo "    root (根CA)"
    echo "    └── intermediate (中间CA)"
    for ap in http:h grpc:g docker:d applications:a libp2p:l; do
        auth=${ap%%:*}
        prefix=${ap#*:}
        echo "        ├── ${auth}-authority"
        echo "        │   ├── ${prefix}server_sign.crt / ${prefix}server_enc.crt  (双证书)"
        echo "        │   ├── ${prefix}client_sign.crt / ${prefix}client_enc.crt  (双证书)"
        echo "        │   ├── ${prefix}server.crt / ${prefix}client.crt           (RFC8998 单证书)"
        echo "        │   └── ${prefix}server-chain.crt / ${prefix}client-chain.crt (部署用实体链)"
    done
    echo ""
    echo "  文件统计:"
    key_count=$(ls *.key 2>/dev/null | wc -l)
    crt_count=$(ls *.crt 2>/dev/null | wc -l)
    cnf_count=$(ls *.cnf 2>/dev/null | wc -l)
    echo "    配置文件 (.cnf): ${cnf_count} 个"
    echo "    私钥文件 (.key): ${key_count} 个"
    echo "    证书文件 (.crt): ${crt_count} 个"
    echo ""
    echo "  算法: SM2 (密钥) + SM3 (摘要)"
    echo "  模式: 同一 CA 下 双证书 (GM/T0024) + 单证书 (RFC8998/GM TLS 1.3)"
    echo ""
    echo "============================================================"
}

usage() {
    cat <<'EOF'
铜锁国密证书一键生成脚本 (POSIX sh): GM/T0024 双证书 + RFC8998 单证书 合并版

用法:
  ./mkcert.sh [选项]

选项:
  (无)              全新生成: CA 层 + 五组实体 (双证书 + 单证书, 全链强制 SM2 ID)
  --keep-ca-keys    保留 CA 私钥, 仅重签 CA 层证书 + 全部实体
  --resign-ca       保留 CA 私钥, 仅重签 CA 层证书 (root/intermediate/authority)
  --reset-cnf       强制按模板重建全部实体 cnf (手改过的 SAN/扩展会丢失)
  --cnf-only        只落盘/重建 cnf 模板, 不签发任何证书 (供手改 SAN)
  -h, --help        显示本帮助

说明:
  - 自包含版: 全部 cnf 模板 (CA 层 + 双证书 + 单证书) 内嵌, 新环境只需拷贝
    本文件即可生成全套证书; 已有 cnf 默认保留 (便于手改 SAN)

环境变量:
  TONGSUOSSL=...    铜锁命令 (默认 tongsuo; 原生 openssl 亦可, auto 下不强制 sigopt)
  SM2_SIGOPT=auto|1|0   SM2 标识是否显式强制 (默认 auto: tongsuo 强制 / openssl 关闭)
  SM2_ID=...        自定义 SM2 标识 (默认 1234567812345678)

每个实体生成 3 张叶子证书:
  {prefix}server_sign.crt + {prefix}server_enc.crt   GM/T0024 双证书 (NTLS)
  {prefix}server.crt                                  RFC8998 单证书 (GM TLS 1.3)
  client 同理; 另附部署用 {prefix}server-chain.crt / {prefix}client-chain.crt

其他:
  - 所有 SM2 签名强制 sm2_id=1234567812345678 (GM/T 标准值),
    生成的链可被 Go gmsm / 原生 OpenSSL / 铜锁按标准 ID 校验
  - 双证书与单证书共用同一套 CA 链 (root -> intermediate -> {app}-authority)
  - CA 重签保留私钥, 下级已签发证书仍可被铜锁验证; 但旧叶子用非标准
    ID 签名, 要被 Go gmsm/原生 OpenSSL 校验需执行 ./renew.sh renew-all
  - 与 mkcert-official.sh (铜锁官方文档风格) 的区别见 README.md
EOF
}

# ============================================================
# 主流程
# ============================================================

main() {
    KEEP_CA_KEYS=0
    RESIGN_CA_ONLY=0
    RESET_CNF=0
    CNF_ONLY=0

    for arg in "$@"; do
        case "$arg" in
            -h|--help) usage; exit 0 ;;
            --keep-ca-keys) KEEP_CA_KEYS=1 ;;
            --resign-ca) KEEP_CA_KEYS=1; RESIGN_CA_ONLY=1 ;;
            --reset-cnf) RESET_CNF=1 ;;
            --cnf-only) CNF_ONLY=1 ;;
            *) error "未知参数: $arg (用 -h 查看帮助)" ;;
        esac
    done

    echo ""
    echo "============================================================"
    echo "  铜锁国密证书一键生成 (双证书 GM/T0024 + 单证书 RFC8998)"
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
    gen_authority "http" "h" "HTTP"
    gen_authority "grpc" "g" "GRPC"
    gen_authority "docker" "d" "DOCKER"
    gen_authority "applications" "a" "APPLICATIONS"
    gen_authority "libp2p" "l" "LIBP2P"

    if [ "$RESIGN_CA_ONLY" -eq 1 ]; then
        info "resign-ca 模式: 仅重签 CA 层完成, 实体证书未动"
        info "如需让旧实体证书可被标准 SM2 ID 校验, 运行: ./renew.sh renew-all"
    else
        # 7. 第四/五层: 实体双证书 + RFC8998 单证书
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
    fi

    # 8. 生成证书链
    gen_chains

    # 9. 验证证书链
    verify_certs

    # 10. 输出统计
    show_summary
}

# 执行主流程
main "$@"
