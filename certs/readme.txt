1. cp gserver.conf lclient.conf
2. 修改lclient.conf 中的commonName dns信息
3. openssl genrsa -out lclient.key 2048
4. openssl req -new -key lclient.key -out lclient.csr -config lclient.conf
5. openssl x509 -req -in lclient.csr -CA root-ca.crt -CAkey root-ca.key -CAcreateserial -out lclient.crt -days 365 -sha256 -extensions v3_req -extfile lclient.conf
