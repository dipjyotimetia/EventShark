# Certificate Generator Dockerfile

This Dockerfile allows you to generate client-server certificates using CFSSL and export them to a local volume.

## Prerequisites

- Docker should be installed on your machine.

## Usage

1. Clone the repository and navigate to the project directory.

2. Place your configuration files (`ca-csr.json`, `ca-config.json`, `peer.json`, `server.json`, `client.json`) in the project directory.

3. Build the Docker image:

```shell
docker build -t certificate-generator .
```

4. Run a container from the image, mapping the `/certs_volume` directory to a local volume on your host machine:

```shell
docker run -v "$(pwd):/certs_volume" certificate-generator
```

5. After running the container, the generated certificates will be available in the specified local volume.

6. At the end verify the generated pem files

```shell
openssl x509 -in ca.pem -text -noout
openssl x509 -in server.pem -text -noout
openssl x509 -in client.pem -text -noout
```

7. Extract public key from a certificate

```shell
openssl x509 -in certificate.pem -pubkey -noout > public.pem
```
