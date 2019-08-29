module github.com/selectel/cert-manager-webhook-selectel

go 1.12

require (
	github.com/go-acme/lego/v3 v3.0.2
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.10.0-alpha.0
	github.com/stretchr/testify v1.3.0
	k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190413052642-108c485f896e

replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550

replace github.com/miekg/dns => github.com/miekg/dns v0.0.0-20170721150254-0f3adef2e220
