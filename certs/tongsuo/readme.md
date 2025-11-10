# 手动使用 tongsuossl 生成证书

## 准备工作

* 安装 tongsuossl

## 步骤

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
