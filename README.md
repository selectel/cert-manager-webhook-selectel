# cert-manager-webhook-selectel
[![Build Status](https://travis-ci.org/selectel/cert-manager-webhook-selectel.svg?branch=master)](https://travis-ci.org/selectel/cert-manager-webhook-selectel)
[![Go Report Card](https://goreportcard.com/badge/github.com/selectel/cert-manager-webhook-selectel)](https://goreportcard.com/report/github.com/selectel/cert-manager-webhook-selectel)

Cert-manager ACME DNS webhook provider for Selectel.

## Contents

* [Issuing certificate in DNS Hosting (actual)](#issuing-certificate-in-dns-hosting-actual)
  * [Installing](#installing)
  * [Setup credentials](#setup-credentials)
  * [Setup issuer](#setup-issuer)
  * [Issuing certificate](#issuing-certificate)
* [Issuing certificate in DNS Hosting (legacy)](#issuing-certificate-in-dns-hosting-legacy)
  * [Legacy version](#legacy-version)
  * [Installing](#installing-legacy)
  * [Setup credentials](#setup-credentials-legacy)
  * [Setup issuer](#setup-issuer-legacy)
  * [Issuing certificate](#issuing-certificate-legacy)
* [Development guide](#development-guide)
  * [Running the test suite](#running-the-test-suite)

## Issuing certificate in DNS Hosting (actual)

### Installing

To install with helm from helm-repository, run:

```bash
$ helm repo add selectel https://selectel.github.io/cert-manager-webhook-selectel
$ helm repo update
$ helm install cert-manager-webhook-selectel selectel/cert-manager-webhook-selectel -n cert-manager
```

Or install with helm from git-repository, run:

```bash
$ git clone https://github.com/selectel/cert-manager-webhook-selectel.git
$ cd cert-manager-webhook-selectel/deploy/cert-manager-webhook-selectel
$ helm install cert-manager-webhook-selectel . -n cert-manager
```

### Setup credentials

Create secret and fill authentication data.

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
```

**KEYSTONE_USER** - Name of the service user. To get the name, in the top right corner of the [Control panel](https://my.selectel.ru/profile/users_management/users?type=service), go to the account menu ⟶ **Profile and Settings** ⟶ **User management** ⟶ the **Service users** tab ⟶ copy the name of the required user. Learn more about [Service users](https://docs.selectel.ru/control-panel-actions/users-and-roles/user-types-and-roles/).

**KEYSTONE_PASSWORD** - Password of the service user.

**ACCOUNT_ID** - Selectel account ID. The account ID is in the top right corner of the [Control panel](https://my.selectel.ru/). Learn more about [Registration](https://docs.selectel.ru/control-panel-actions/account/registration/).

**SELECTEL_PROJECT_ID** - Unique identifier of the associated Cloud Platform project. To get the project ID, in the [Control panel](https://my.selectel.ru/vpc/), go to Cloud Platform ⟶ project name ⟶ copy the ID of the required project. Learn more about [Cloud Platform projects](https://docs.selectel.ru/cloud/servers/about/projects/).

### Setup issuer

An example issuer:

```yaml
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
            timeout: 60 # Default 40
```

### Issuing certificate

Issuing certificate:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sel-letsencrypt-crt
  namespace: cert-manager
spec:
  # Setup secret name
  secretName: example-com-tls
  commonName: example.com
  issuerRef:
    name: letsencrypt-staging
    kind: Issuer
  # Setup names of zones
  dnsNames:
  - example.com
  - www.example.com
```

## Issuing certificate in DNS Hosting (legacy)

### Legacy version

Cert-manager webhook provider for Selectel supporting two versions API.
They are not compatible. They utilize different API and work with zones live on different authoritative servers.
Zone created in v2 API not available via v1 api.

### Installing (legacy)

To install with helm from helm-repository, run:

```bash
$ helm repo add selectel https://selectel.github.io/cert-manager-webhook-selectel
$ helm repo update
$ helm install cert-manager-webhook-selectel selectel/cert-manager-webhook-selectel -n cert-manager --version 1.2.5
```

Or install with helm from git-repository, run:

```bash
$ git clone https://github.com/selectel/cert-manager-webhook-selectel.git --branch cert-manager-webhook-selectel-1.2.5
$ cd cert-manager-webhook-selectel/deploy/cert-manager-webhook-selectel
$ helm install cert-manager-webhook-selectel . -n cert-manager
```

### Setup credentials (legacy)

Create secret and fill **APITOKEN_FROM_MY_SELECTEL_RU**.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: selectel-api-key
  namespace: cert-manager
type: Opaque
stringData:
  token: APITOKEN_FROM_MY_SELECTEL_RU
```

**APITOKEN_FROM_MY_SELECTEL_RU** - Selectel Token (API Key). For obtain Selectel Token read [here](https://developers.selectel.ru/docs/control-panel/authorization/).

### Setup issuer (legacy)

An example issuer:

```yaml
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

### Issuing certificate (legacy)

Issuing certificate:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: sel-letsencrypt-crt
  namespace: cert-manager
spec:
  # Setup secret name
  secretName: example-com-tls
  commonName: example.com
  issuerRef:
    name: letsencrypt-staging
    kind: Issuer
  # Setup names of zones
  dnsNames:
  - example.com
  - www.example.com
```

## Development guide

### Running the test suite

You can run the test suite with:

1. Go to `https://my.selectel.ru/profile/users_management/users`, get one or create new user.
2. Fill in the appropriate values in `testdata/selectel/dns-credentials.yml` and `testdata/selectel/config.json`.
    * Insert values `testdata/selectel/dns-credentials.yml`.
    * Check that `metadata.name` in `testdata/selectel/dns-credentials.yml` equals value in `testdata/selectel/config.json` for key `dnsSecretRef.name`.
3. Run tests

```bash
$ TEST_ZONE_NAME=example.com. make test
```
