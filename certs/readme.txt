1. cp gserver.conf lclient.conf
2. 修改lclient.conf 中的commonName dns信息
3. openssl genrsa -out lclient.key 2048
4. openssl req -new -key lclient.key -out lclient.csr -config lclient.conf
5. openssl x509 -req -in lclient.csr -CA root-ca.crt -CAkey root-ca.key -CAcreateserial -out lclient.crt -days 365 -sha256 -extensions v3_req -extfile lclient.conf


###### gmssl
0. 安装
	git clone https://github.com/guanzhi/GmSSL.git gmssl
	https://github.com/guanzhi/GmSSL/archive/master.zip
	这里 v3.1.1 貌似不是最新的代码 使用master分支执行
	cd gmssl && git checkout v3.1.1
    mkdir build && cd build
	cmake ..
    make
    make test
    sudo make install
	ldd /usr/local/bin/gmssl # 能够看到缺少libgmssl.so.3  # 如果不缺就不放
	sudo cp ./bin/libgmssl.so.3 /usr/bin/ # 应该是libgmssl.so.3 缺失
	gmssl version
1. root key
	gmssl sm2keygen -out gmroot.key -pubout gmroot.pub -pass xxxxxxxx
	gmssl certgen -C CN -ST Xinjiang -O jianlu -OU IT -CN root -days 3650 -key gmroot.key -pass xxxxxxxx -out gmroot.crt -key_usage keyCertSign -key_usage cRLSign -key_usage digitalSignature -gen_authority_key_id -gen_subject_key_id -ca

2. http
	gmssl sm2keygen -out gmhserver.key -pass gmhttp -pubout gmhserver.pub
	gmssl reqgen -C CN -ST Xinjiang -L Urumqi -O jianlu -OU IT -CN http -key gmroot.key -pass xxxxxxxx -out gmhserver.csr
	gmssl reqsign -in gmhserver.csr -days 365 -key_usage keyCertSign -path_len_constraint 0 -cacert gmroot.crt -key gmroot.key -pass xxxxxxxx -gen_authority_key_id -gen_subject_key_id -serial_len 12 -out gmhserver.crt -subject_dns_name localhost -subject_dns_name jianlu.site -subject_dns_name http -issuer_dns_name 127.0.0.1 -issuer_dns_name 192.168.58.110 -ext_key_usage serverAuth

# ca证书颁发签名证书和加密证书
    gmssl sm2keygen -pass 1234 -out signkey.pem
    gmssl reqgen -C CN -ST Beijing -L Haidian -O PKU -OU CS -CN localhost -key signkey.pem -pass 1234 -out signreq.csr
    gmssl reqsign -in signreq.csr -days 365 -key_usage digitalSignature -cacert cacert.cer -key cakey.pem -pass 1234 -out signcert.cer

    gmssl sm2keygen -pass 1234 -out enckey.pem
    gmssl reqgen -C CN -ST Beijing -L Haidian -O PKU -OU CS -CN localhost -key enckey.pem -pass 1234 -out encreq.csr
    gmssl reqsign -in encreq.csr -days 365 -key_usage keyEncipherment -cacert cacert.cer -key cakey.pem -pass 1234 -out enccert.cer

# 合并ca证书和签名证书 并验证
    cat signcert.cer > certs.cer
    cat cacert.cer >> certs.cer
    gmssl certverify -in certs.cer -cacert rootcacert.cer
