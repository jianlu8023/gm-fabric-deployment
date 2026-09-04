#export PATH=/opt/tongsuo/bin:$PATH

rm -f *.key *.csr *.crt
rm -rf {newcerts,db,private,crl}
mkdir {newcerts,db,private,crl}
touch db/{index,serial}
echo 00 > db/serial

# sm2 ca
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out ca.key

tongsuossl req -config ca.cnf -new -key ca.key -out ca.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=root ca"

tongsuossl ca -selfsign -config ca.cnf -in ca.csr -keyfile ca.key -extensions v3_ca -days 3650 -notext -out ca.crt -md sm3 -batch

# sm2 middle ca
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out subca.key

tongsuossl req -config ca.cnf -new -key subca.key -out subca.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=sub ca"

tongsuossl ca -config ca.cnf -extensions v3_intermediate_ca -days 3650 -in subca.csr -notext -out subca.crt -md sm3 -batch

cat ca.crt subca.crt > chain-ca.crt

# server sm2 double certs
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out server_sign.key

tongsuossl req -config subca.cnf -key server_sign.key -new -out server_sign.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=grpc"

tongsuossl ca -config subca.cnf -extensions server_sign_req -days 3650 -in server_sign.csr -notext -out server_sign.crt -md sm3 -batch

tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out server_enc.key

tongsuossl req -config subca.cnf -key server_enc.key -new -out server_enc.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=grpc"

tongsuossl ca -config subca.cnf -extensions server_enc_req -days 3650 -in server_enc.csr -notext -out server_enc.crt -md sm3 -batch

# client sm2 double certs
tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out client_sign.key

tongsuossl req -config subca.cnf -key client_sign.key -new -out client_sign.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=grpc"

tongsuossl ca -config subca.cnf -extensions client_sign_req -days 3650 -in client_sign.csr -notext -out client_sign.crt -md sm3 -batch

tongsuossl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:sm2 -out client_enc.key

tongsuossl req -config subca.cnf -key client_enc.key -new -out client_enc.csr -sm3 -nodes -subj "/C=CN/ST=Xinjiang/O=jianlu/OU=IT/CN=grpc"

tongsuossl ca -config subca.cnf -extensions client_enc_req -days 3650 -in client_enc.csr -notext -out client_enc.crt -md sm3 -batch
