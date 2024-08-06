# Tendo

A simple and basic tool to auto update Tencent SSL Certificate based on Cert Manager certificate secret.

## Background

At the moment we release this tool, Tencent SSL Certificate only support SSL Certificates in kubernetes secret with secret type `Opaque`. 

We are using Cert-Manager in our kubernetes cluster to auto generate and verify SSL Certificates from Let's Encrypt. Unfortunately, the secret created by Cert Manager comes with the type `kubernetes.io/tls`. 

As a result, we can not export the certificate into Tencent SSL Certificate.

## Motivation

Making kubernetes secret with type `kubernetes.io/tls` working with Tencent SSL Certificates is not really hard. All we need is:

* Create a new certificate in Tencent Cloud with `Upload Certificate` option.
* Signing certificate is the `tls.crt` data from kubernetes secret. 
* Signing private key is the `tls.key` data from kubernetes secret.
* Create a secret with type `Opaque` in kubernetes cluster and add the Tencent Cloud SSL Certificate ID

```yaml
# Opaque Secret Example
apiVersion: v1
data:
  qcloud_cert_id: Q2VydGlmaWNhdGVTb21lSUQ=
kind: Secret
metadata:
  name: example-domain-opaque
  namespace: tendo
type: Opaque
```

While manually adding and modifying the certificate is not a hard task, we need to automate the process because let's encrypt certificates that need to be renewed every three months. With this tool, the whole process will be done automatically.

## Building

To build the tool, make sure golang already available on your system or you can build the docker image also.

Make sure the `config.yaml` file is placed in the same directory with the current directory you are running the binary is. Take a look for [config example](./config/config.yaml.example) for config references.

### Manual Build

Run this to build the binary

```bash
go mod tidy
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o app main.go

# run the tool
./app
```

### Docker Build

Build the docker image in your local system.

```bash
docker build -t tendo:latest -f dockerbuild/Dockerfile .
```

Or just pull the released version from dockerhub.

```bash
docker pull fredytarigan/tendo:latest
```

## Kubernetes Deployment

There is an example for kubernetes deployment in [deploy](./deploy/) directory. You need to adjust the namespace and configmap into your needs.



