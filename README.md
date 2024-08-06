# Tendo

A simple and basic tools to auto update Tencent SSL Certificate based on Cert-Manager certificate secret.

## Background

At the moment we release this tools, Tencent SSL Certificate only support ssl certificate in kubernetes secret with secret type `Opaque`. 

We are using Cert-Manager in our kubernetes cluster to auto generate and verify SSL Certificate from Let's Encrypt. Unfortunately, the secret created by Cert-Manager come with type `kubernetes.io/tls`. 

As a result, we can not export the certificate into Tencent SSL Certificate.

## Motivation

Making kubernetes secret with type `kubernetes.io/tls` working with Tencent SSL Certificates is not really hard. All we need is:

* Create a new certificate in Tencent Cloud with `Upload Certificate` option.
* Signing certificate is the `tls.crt` data from kubernetes secret. 
* Signing private key is the `tls.key` data from kubernetes secret.
* Create a secret with type `Opaque` in kubernetes cluster and add the Tencent Cloud SSL Certificate ID

```yaml
apiVersion: v1
data:
  qcloud_cert_id: Q2VydGlmaWNhdGVTb21lSUQ=
kind: Secret
metadata:
  name: example-domain-opaque
  namespace: traefik
type: Opaque
```