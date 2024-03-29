FROM cfssl/cfssl AS cfssl

LABEL author="Dipjyoti Metia"
LABEL version="1.0"


WORKDIR /app

COPY *.json /app/

# Generate the certificates
RUN cfssl gencert -initca ca-csr.json | cfssljson -bare ca -
RUN cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=peer peer.json | cfssljson -bare peer
RUN cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server server.json | cfssljson -bare server
RUN cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client client.json | cfssljson -bare client

FROM debian:12-slim

COPY --from=cfssl /app/*.pem /app/*.csr /certs/

WORKDIR /certs

CMD ["cp", "-R", "/certs", "/certs_volume"]