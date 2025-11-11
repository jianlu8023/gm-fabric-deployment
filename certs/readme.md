# openssl 证书链 (root-> intermediate -> {http-authority,grpc-authority,docker-authority} -> {hserver,hclient,gserver,gclient})

## 生成 root 证书

* 生成 root 证书私钥

```shell
openssl genrsa -out root.key 4096
```

* 编写 root.conf

```
# openssl-root.cnf
[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_ca # 指定使用 v3_ca 扩展

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed Certificate Authority
# OU = Certificate Authority
CN = Self-Signed Root Certificate Authority

[ v3_ca ]
# Extensions for a root CA
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer # 对于自签名证书，issuer就是自己
basicConstraints = critical, CA:TRUE, pathlen:2 # pathlen:1 允许其签发一个中间CA和实体证书。如果是0，则只能签发实体证书。对于根CA，通常可以不指定pathlen或设为较大值。
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
# extendedKeyUsage = serverAuth, clientAuth # 根CA通常不直接用于服务器/客户端认证，所以通常不加。如果加了，它也可以直接用于终端实体证书的功能

```

* 生成 root 证书

```shell
openssl req -new -x509 -nodes -key root.key -sha256 -days 3650 -out root.crt -config root.conf
```

## 生成 intermediate 证书 (作为中间证书)

* 生成 intermediate 证书私钥

```shell
openssl genrsa -out intermediate.key 4096
```

* 编写 intermediate.conf

```
# openssl-intermediate.cnf

[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_intermediate_ca # 注意：这一行只在 openssl req -x509 时才有效，生成CSR时不使用

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed Intermediate Certificate Authority
OU = Intermediate Certificate Authority
CN = Self-Signed Intermediate Certificate Authority # 中间CA的Common Name

[ v3_intermediate_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE, pathlen:1 # pathlen:0 表示这个中间CA只能签署实体证书，不能再签署其他中间CA
keyUsage = critical, digitalSignature, cRLSign, keyCertSign

```

* 签发 intermediate 证书

```shell
openssl req -new -key intermediate.key -sha256 -out intermediate.csr -config intermediate.conf
openssl x509 -req -in intermediate.csr -CA root.crt -CAkey root.key -CAcreateserial -out intermediate.crt -days 1825 -sha256 -extfile intermediate.conf -extensions v3_intermediate_ca
```

## 生成 {xxx}-authority 证书

* 生成 {xxx}-authority 证书私钥

```shell
openssl genrsa -out {xxx}-authority.key 4096
```

* 编写 {xxx}-authority 证书配置文件

```
# openssl-applications-authority.cnf

[ req ]
default_bits = 4096
default_md = sha256
prompt = no
encrypt_key = no
distinguished_name = req_distinguished_name
x509_extensions = v3_application_ca # 注意：这一行只在 openssl req -x509 时才有效，生成CSR时不使用

[ req_distinguished_name ]
C = CN
ST = Xinjiang
L = Urumqi
O = The Self-Signed Applications Certificate Authority
OU = Applications Certificate Authority
CN = Self-Signed Applications Certificate Authority # 中间CA的Common Name

[ v3_application_ca ]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE, pathlen:0 # pathlen:0 表示这个中间CA只能签署实体证书，不能再签署其他中间CA
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
extendedKeyUsage = serverAuth, clientAuth  # 允许签发服务器和客户端证书

```

* 签发 {xxx}-authority 证书

```shell
openssl req -new -key {xxx}-authority.key -sha256 -out {xxx}-authority.csr -config {xxx}-authority.cnf
openssl x509 -req -in {xxx}-authority.csr -CA intermediate.crt -CAkey intermediate.key -CAcreateserial -out {xxx}-authority.crt -days 1095 -sha256 -extfile {xxx}-authority.cnf -extensions v3_application_ca
```

## 生成 hserver 证书

* 生成 hserver 证书私钥

```shell
openssl genrsa -out hserver.key 4096
```

* 编写 hserver 证书配置文件

```
# openssl-http.conf
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
organizationName = The Self-Signed Http Certificate
organizationalUnitName = Http Certificate Authority
commonName = www.jianlu.site
# emailAddress = jianlu8023@gmail.com

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = *.jianlu.site
DNS.2 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

# =========================================================
# 新增部分：用于最终HTTP服务器证书的扩展
# =========================================================
[v3_http_cert]
# 基本约束：这是一个终端实体证书，不能用于签发其他证书
basicConstraints = critical, CA:FALSE
# 定义此证书的主题公钥标识符
subjectKeyIdentifier = hash
# 定义签发此证书的CA的公钥标识符
authorityKeyIdentifier = keyid:always,issuer
# 密钥用途：指定此证书的私钥可以用于哪些操作
# digitalSignature: 用于TLS握手中的数字签名
# keyEncipherment: 用于加密TLS会话密钥
keyUsage = critical, digitalSignature, keyEncipherment
# 扩展密钥用途：指定此证书的具体用途
# serverAuth: 允许作为TLS服务器进行身份验证
# clientAuth: 允许作为TLS客户端进行身份验证 (可选，但通常也加上)
extendedKeyUsage = serverAuth, clientAuth
# 主题备用名称：非常重要，现代浏览器强制要求使用SANs，并且它必须在最终证书中
subjectAltName = @alt_names

```

* 签发 hserver 证书

```shell
openssl req -new -key hserver.key -sha256 -out hserver.csr -config hserver.conf
openssl x509 -req -in hserver.csr -CA http-authority.crt -CAkey http-authority.key -CAcreateserial -out hserver.crt -days 365 -sha256 -extfile hserver.conf -extensions v3_http_cert
```

## 生成 gserver 证书

* 生成 gserver 证书私钥

```shell
openssl genrsa -out gserver.key 4096
```

* 编写 gserver.conf

```
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
organizationName = The Self-Signed Grpc Server Certificate
organizationalUnitName = Grpc Server Certificate Authority
commonName = gserver.jianlu.site
# emailAddress = jianlu8023@gmail.com

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = gserver.jianlu.site
DNS.2 = *.jianlu.site
DNS.4 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

# =========================================================
# 用于最终 gRPC 服务器证书的扩展
# =========================================================
[v3_grpc_server_cert]
basicConstraints = critical, CA:FALSE # 这是一个终端实体证书，不能签发其他证书
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = critical, digitalSignature, keyEncipherment # 数字签名和密钥加密是TLS服务器必需的
extendedKeyUsage = serverAuth # 明确指定此证书用于TLS服务器认证
subjectAltName = @alt_names # 确保SANs最终包含在签发的证书中

```

* 生成 gserver 证书

```shell
# 后续步骤和 hserver 一样 最后 x509 的那步 extensions v3_grpc_server_cert
```

## 生成 gclient 证书

* 生成 gclient 证书私钥

```shell
# 省略
```

* 编写 gclient 证书配置文件

```
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
organizationName = The Self-Signed Grpc Client Certificate
organizationalUnitName = Grpc Client Certificate Authority
commonName = gclient.jianlu.site
# emailAddress = jianlu8023@gmail.com

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = gclient.jianlu.site
DNS.2 = *.jianlu.site
DNS.3 = localhost
DNS.4 = jianlu.site
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

# =========================================================
# 用于最终 gRPC 客户端证书的扩展
# =========================================================
[v3_grpc_client_cert]
basicConstraints = critical, CA:FALSE # 这是一个终端实体证书
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = critical, digitalSignature # 数字签名是TLS客户端必需的
extendedKeyUsage = clientAuth # 明确指定此证书用于TLS客户端认证
# 如果客户端也需要SANs，则取消注释以下一行
subjectAltName = @alt_names

```

* 签发 gclient 证书

```shell
# -extensions v3_grpc_client_cert
```

# linux 添加根证书到系统信任中

```
sudo cp root.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

# 简化版颁发证书 (root -> {hserver,gserver,lserver等等})

```text
cp gserver.conf lclient.conf
修改lclient.conf 中的commonName dns信息
openssl genrsa -out lclient.key 2048
openssl req -new -key lclient.key -out lclient.csr -config lclient.conf
openssl x509 -req -in lclient.csr -CA root-ca.crt -CAkey root-ca.key -CAcreateserial -out lclient.crt -days 365 -sha256 -extensions v3_req -extfile lclient.conf
```

# gmssl (编写时的v3版本只能签发带密码的证书)

## 安装

```shell
git clone https://github.com/guanzhi/GmSSL.git gmssl
# https://github.com/guanzhi/GmSSL/archive/master.zip 链接 
# 这里 v3.1.1 貌似不是最新的代码 使用master分支执行
cd gmssl && git checkout v3.1.1 # 这一步 应该使用master分支
mkdir build && cd build
cmake ..
make
make test
sudo make install
ldd /usr/local/bin/gmssl # 能够看到缺少libgmssl.so.3 # 如果不缺就不放
sudo cp ./bin/libgmssl.so.3 /usr/bin/ # 应该是libgmssl.so.3 缺失
gmssl version
```

## 生成 root 证书 (使用gmssl 生成的证书其实用不成(或者说自己不会用), 因为是带密码的)

* 生成 root 证书私钥

```shell
gmssl sm2keygen -out gmroot.key -pubout gmroot.pub -pass xxxxxxxx
```

* 生成 root 证书

```shell
gmssl certgen -C CN -ST Xinjiang -O jianlu -OU IT -CN root -days 3650 -key gmroot.key -pass xxxxxxxx -out gmroot.crt -key_usage keyCertSign -key_usage cRLSign -key_usage digitalSignature -gen_authority_key_id -gen_subject_key_id -ca
```

## 生成http 证书

* 生成 http 证书私钥

```shell
gmssl sm2keygen -out gmhserver.key -pass gmhttp -pubout gmhserver.pub
```

* 生成 http csr

```shell
gmssl reqgen -C CN -ST Xinjiang -L Urumqi -O jianlu -OU IT -CN http -key gmroot.key -pass xxxxxxxx -out gmhserver.csr
```

* 生成 http 证书

```shell
gmssl reqsign -in gmhserver.csr -days 365 -key_usage keyCertSign -path_len_constraint 0 -cacert gmroot.crt -key gmroot.key -pass xxxxxxxx -gen_authority_key_id -gen_subject_key_id -serial_len 12 -out gmhserver.crt -subject_dns_name localhost -subject_dns_name jianlu.site -subject_dns_name http -issuer_dns_name 127.0.0.1 -issuer_dns_name 192.168.58.110 -ext_key_usage serverAuth
```

## 证书颁发签名证书和加密证书 (如果使用gm证书启动http服务,需要两套keypair 一套签名 一套加密)

```shell
gmssl sm2keygen -pass 1234 -out signkey.pem
gmssl reqgen -C CN -ST Beijing -L Haidian -O PKU -OU CS -CN localhost -key signkey.pem -pass 1234 -out signreq.csr
gmssl reqsign -in signreq.csr -days 365 -key_usage digitalSignature -cacert cacert.cer -key cakey.pem -pass 1234 -out signcert.cer

gmssl sm2keygen -pass 1234 -out enckey.pem
gmssl reqgen -C CN -ST Beijing -L Haidian -O PKU -OU CS -CN localhost -key enckey.pem -pass 1234 -out encreq.csr
gmssl reqsign -in encreq.csr -days 365 -key_usage keyEncipherment -cacert cacert.cer -key cakey.pem -pass 1234 -out enccert.cer
```

## 合并ca证书和签名证书 并验证

```shell
cat signcert.cer > certs.cer
cat cacert.cer >> certs.cer
gmssl certverify -in certs.cer -cacert rootcacert.cer
```

# tongsuossl (openssl的gm实现,可签发不带密码的证书)

## 安装

```shell

```

## tongsuossl 证书链 (root -> intermediate -> {http-authority,grpc-authority} -> {hserver_sign.crt,hserver_enc.crt,hclient_sign.crt,hclient_enc.crt})

### 生成根证书

* 生成私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out root.key
```

* 编辑 root.cnf

```text
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
```

* 生成证书

```shell
# 不知道为什么需要-subj 在cnf中指定不行
tongsuossl req -new -x509 -nodes -key root.key -sm3 -days 3650 -out root.crt -config root.cnf -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=The Self-Signed Certificate Authority/CN=Self-Signed Root Certificate Authority"
```

### 生成中间证书

* 生成中间证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out intermediate.key
```

* 编写intermediate.cnf

```text
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
```

* 生成中间证书签名请求

```shell
tongsuossl req -config intermediate.cnf -new -key intermediate.key -out intermediate.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=The Self-Signed Intermediate Certificate Authority/OU=Intermediate Certificate Authority/CN=Self-Signed Intermediate Certificate Authority"
```

* 使用根证书签名中间证书

```shell
tongsuossl x509 -req -in intermediate.csr -CA root.crt -CAkey root.key -CAcreateserial -out intermediate.crt -days 1825 -sm3 -extfile intermediate.cnf -extensions v3_intermediate_ca
```

### 生成 http-authority 证书

* 生成 http-authority 证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out http-authority.key
```

* 编写 http-authority.cnf

```text
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
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
extendedKeyUsage = serverAuth, clientAuth
```

* 生成 http-authority 证书签名请求

```shell
tongsuossl req -config http-authority.cnf -new -key http-authority.key -out http-authority.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=The Self-Signed HTTP Certificate Authority/OU=HTTP Certificate Authority/CN=Self-Signed HTTP Certificate Authority"
```

* 使用中间证书签名 http-authority 证书

```shell
tongsuossl x509 -req -in http-authority.csr -CA intermediate.crt -CAkey intermediate.key -CAcreateserial -out http-authority.crt -days 1095 -sm3 -extfile http-authority.cnf -extensions v3_intermediate_ca
```

### 生成 hserver 证书

Tips:

* 若 使用 tjfoc/gmsm 库 go build -tags=single_cert 则可按照上面openssl 生成单证书模式
* 若 使用 tjfoc/gmsm 库 go build 则需要按照下方生成sign enc 两套证书
* hclient 同理
* 使用 单证书/双证书 取决于 go build 是否携带single_cert的tag，tjfoc/gmsm 默认是双证书

#### hserver_sign 签名证书

* 生成 hserver_sign 证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out hserver_sign.key
```

* 编写 hserver_sign.cnf

```text
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only
default_md          = sha256
req_extensions = v3_req

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
DNS.1 = *.jianlu.site
DNS.2 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

[ server_sign_req ]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = nonRepudiation, digitalSignature
subjectAltName = @alt_names
extendedKeyUsage = serverAuth
```

* 生成 hserver_sign 证书签名请求

```shell
tongsuossl req -config hserver_sign.cnf -new -key hserver_sign.key -out hserver_sign.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=HTTP Certificate Authority/OU=HTTP Certificate/CN=www.jianlu.site"
```

* 使用 http-authority 证书签名 hserver_sign 证书

```shell
tongsuossl x509 -req -in hserver_sign.csr -CA http-authority.crt -CAkey http-authority.key -CAcreateserial -out hserver_sign.crt -days 365 -sm3 -extfile hserver_sign.cnf -extensions server_sign_req
```

#### hserver_enc 加密证书

* 生成 hserver_enc 证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out hserver_enc.key
```

* 编写 hserver_enc.cnf

```text
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only
default_md          = sha256
req_extensions = v3_req

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
DNS.1 = *.jianlu.site
DNS.2 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

[ server_enc_req ]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = keyAgreement, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
extendedKeyUsage = serverAuth
```

* 生成 hserver_enc 证书签名请求

```shell
tongsuossl req -config hserver_enc.cnf -new -key hserver_enc.key -out hserver_enc.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=HTTP Certificate Authority/OU=HTTP Certificate/CN=www.jianlu.site"
```

* 使用 http-authority 证书签名 hserver_enc 证书

```shell
tongsuossl x509 -req -in hserver_enc.csr -CA http-authority.crt -CAkey http-authority.key -CAcreateserial -out hserver_enc.crt -days 365 -sm3 -extfile hserver_enc.cnf -extensions server_enc_req
```

### 生成 hclient 证书

#### hclient_sign 签名证书

* 生成 hclient_sign 证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out hclient_sign.key
```

* 编写 hclient_sign.cnf

```text
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only
default_md          = sha256
req_extensions = v3_req

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
DNS.1 = *.jianlu.site
DNS.2 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

[ client_sign_req ]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = nonRepudiation, digitalSignature
subjectAltName = @alt_names
extendedKeyUsage = clientAuth
```

* 生成 hclient_sign 证书签名请求

```shell
tongsuossl req -config hclient_sign.cnf -new -key hclient_sign.key -out hclient_sign.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=HTTP Certificate Authority/OU=HTTP Certificate/CN=hclient.jianlu.site"
```

* 使用 http-authority 证书签名 hclient_sign 证书

```shell
tongsuossl x509 -req -in hclient_sign.csr -CA http-authority.crt -CAkey http-authority.key -CAcreateserial -out hclient_sign.crt -days 365 -sm3 -extfile hclient_sign.cnf -extensions client_sign_req
```

#### hclient_enc 加密证书

* 生成 hclient_enc 证书私钥

```shell
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out hclient_enc.key
```

* 编写 hclient_enc.cnf

```text
[ req ]
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only
default_md          = sha256
req_extensions = v3_req

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
DNS.1 = *.jianlu.site
DNS.2 = jianlu.site
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.58.110

[ client_enc_req ]
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = keyAgreement, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
extendedKeyUsage = clientAuth
```

* 生成 hclient_enc 证书签名请求

```shell
tongsuossl req -config hclient_enc.cnf -new -key hclient_enc.key -out hclient_enc.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/L=Urumqi/O=HTTP Certificate Authority/OU=HTTP Certificate/CN=hclient.jianlu.site"
```

* 使用 http-authority 证书签名 hclient_enc 证书

```shell
tongsuossl x509 -req -in hclient_enc.csr -CA http-authority.crt -CAkey http-authority.key -CAcreateserial -out hclient_enc.crt -days 365 -sm3 -extfile hclient_enc.cnf -extensions client_enc_req
```
