# cert-manager-webhook-selectel

Cert-manager ACME DNS webhook provider for Selectel.

## Installing

To install with helm, run:

```bash
$ git clone https://github.com/selectel/cert-manager-webhook-selectel.git
$ cd cert-manager-webhook-selectel/deploy/cert-manager-webhook-selectel
$ helm install --name cert-manager-webhook-selectel .
```

Without helm, run:

```bash
$ make rendered-manifest.yaml
$ kubectl apply -f _out/rendered-manifest.yaml
```

### Issuer/ClusterIssuer

An example issuer:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: selectel-api-key
type: Opaque
stringData:
  key: APITOKEN_FROM_MY_SELECTEL_RU
---
apiVersion: certmanager.io/v1alpha1
kind: Issuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    email: certmaster@selectel.ru
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    dns01:
      providers:
        - name: selectel
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
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: sel-letsencrypt-crt
  namespace: default
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

1. Go to `https://my.selectel.ru/profile/apikeys`, get one or create new api token
2. Fill in the appropriate values in `testdata/selectel/apikey.yml` and `testdata/selectel/config.json` 

```bash
$ ./scripts/fetch-test-binaries.sh
$ TEST_ZONE_NAME=example.com. go test .
```
