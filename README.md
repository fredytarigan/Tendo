# Tendo

A simple and basic tools to auto update Tencent SSL Certificate based on Cert-Manager certificate secret.

## Background

<p align="justify">At the moment we release this tools, Tencent SSL Certificate only support ssl certificate in kubernetes secret with secret type `Opaque`. We are using Cert-Manager in our kubernetes cluster to auto generate and verify SSL Certificate from Let's Encrypt. Unfortunately, the secret created by Cert-Manager come with type `kubernetes.io/tls`. As a result, we can not export the certificate into Tencent SSL Certificate.</p>

