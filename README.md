# cert-manager-webhook-selectel
[![Build Status](https://travis-ci.org/selectel/cert-manager-webhook-selectel.svg?branch=master)](https://travis-ci.org/selectel/cert-manager-webhook-selectel)
[![Go Report Card](https://goreportcard.com/badge/github.com/selectel/cert-manager-webhook-selectel)](https://goreportcard.com/report/github.com/selectel/cert-manager-webhook-selectel)

Cert-manager ACME DNS webhook provider for Selectel.

## Installing

To install with helm in namespace: cert-manager, run:

```bash
$ helm repo add selectel https://selectel.github.io/cert-manager-webhook-selectel
$ helm repo update
$ helm install cert-manager-webhook-selectel selectel/cert-manager-webhook-selectel -n cert-manager
```

OR

```bash
$ git clone https://github.com/selectel/cert-manager-webhook-selectel.git
$ cd cert-manager-webhook-selectel/deploy/cert-manager-webhook-selectel
$ helm install cert-manager-webhook-selectel . -n cert-manager
```

### Issuer/ClusterIssuer

An example issuer:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: selectel-dns-credentials
  namespace: cert-manager
type: Opaque
stringData:
  username: KEYSTONE_USER
  password: KEYSTONE_PASSWORD
  account_id: ACCOUNT_ID
  project_id: SELECTEL_PROJECT_ID
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging
  namespace: cert-manager
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: certmaster@selectel.ru
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.selectel.ru
          solverName: selectel
          config:
            dnsSecretRef:
              name: selectel-dns-credentials
            # Optional config, shown with default values
            #   all times in seconds
            ttl: 120 # Default: 60
            timeout: 30 # Default 40
```

And then you can issue a cert:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sel-letsencrypt-crt
  namespace: cert-manager
spec:
  secretName: example-com-tls
  commonName: example.com
  issuerRef:
    name: letsencrypt-staging
    kind: Issuer
  dnsNames:
  - example.com
  - www.example.com
```

## Legacy version

Cert-manager webhook provider for Selectel support two versions API.
They are not compatible. They utilize different API and work with zones live on different authoritative servers.
Zone created in v2 API not available via v1 api.

### Legacy version installing

To install with helm in namespace: cert-manager, run:

```bash
$ helm repo add selectel https://selectel.github.io/cert-manager-webhook-selectel
$ helm repo update
$ helm install cert-manager-webhook-selectel selectel/cert-manager-webhook-selectel -n cert-manager --version 1.2.4
```

OR

```bash
$ git clone https://github.com/selectel/cert-manager-webhook-selectel.git --branch 1.2.4
$ cd cert-manager-webhook-selectel/deploy/cert-manager-webhook-selectel
$ helm install cert-manager-webhook-selectel . -n cert-manager
```

### Legacy Issuer/ClusterIssuer

An example issuer:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: selectel-api-key
  namespace: cert-manager
type: Opaque
stringData:
  token: APITOKEN_FROM_MY_SELECTEL_RU
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging
  namespace: cert-manager
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: certmaster@selectel.ru
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.selectel.ru
          solverName: selectel
          config:
            apiKeySecretRef:
              name: selectel-api-key
              key: token

            # Optional config, shown with default values
            #   all times in seconds
            ttl: 120
            timeout: 30
            propagationTimeout: 120
            pollingInterval: 2
```

And then you can issue a cert:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sel-letsencrypt-crt
  namespace: cert-manager
spec:
  secretName: example-com-tls
  commonName: example.com
  issuerRef:
    name: letsencrypt-staging
    kind: Issuer
  dnsNames:
  - example.com
  - www.example.com
```

## Development

### Running the test suite

You can run the test suite with:

1. Go to `https://my.selectel.ru/profile/users_management/users`, get one or create new user.
2. Fill in the appropriate values in `testdata/selectel/dns-credentials.yml` and `testdata/selectel/config.json`.
    - Insert values `testdata/selectel/dns-credentials.yml`.
    - Check that `metadata.name` in `testdata/selectel/dns-credentials.yml` equals value in `testdata/selectel/config.json` for key `dnsSecretRef.name`.

```bash
$ TEST_ZONE_NAME=example.com. make test
```
