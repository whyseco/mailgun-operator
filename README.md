# Mailgun operator
Declare your mailgun configuration using Kubernetes custom resources

Available :
- Create your domain
- Configure webhooks
- Configure routes
  
# Instalation

```
kubectl apply -f https://raw.githubusercontent.com/whyseco/mailgun-operator/master/deploy/bundle.yaml
```

# Usage
## Prerequisite
You need to have a mailgun account and retrieve the api key available at https://app.mailgun.com/app/account/security/api_keys

Store your apiKey in a kubernetes secret :

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mailgun-secret
type: Opaque
stringData:
  apiKey: <apiKey>
```
The name of the secret will be use when declaring mailgun object

And store it in your kubernetes cluster `kubectl apply -f secret.yaml`

## Domain
The mailgun operator can create the domain in mailgun for you. It will store domain dns information on the object status

mg-domain.yaml
```yaml
apiVersion: mailgun.com/v1alpha1
kind: MailgunDomain
metadata:
  name: example-mailgundomain
spec:
  domain: mg.foo.com
  secretName: mailgun-secret
```
Where the domain is the domain name you wanna use in mailgun (see https://help.mailgun.com/hc/en-us/articles/202256730-How-Do-I-Pick-a-Domain-Name-for-My-Mailgun-Account-)

You can also set the following parameters :
```yaml
dkimKeySize: 2048
forceDkimAuthority: true
password: 123456
spamAction: tag
webScheme: https
wildcard: false
```

For more information see mailgun api documentation https://documentation.mailgun.com/en/latest/api-domains.html#domains

Execute `kubectl apply -f mg-domain.yaml` to create your domain, **deleting the object will delete the domain on mailgun**


## Webhooks
The mailgun operator can configure your domain webhooks in mailgun.

mg-webhooks.yaml
```yaml
apiVersion: mailgun.com/v1alpha1
kind: MailgunWebhook
metadata:
  name: mailgun-test
spec:
  domain: mg.foo.com
  secretName: mailgun-secret
  opened:
    - https://myapi.foo.com/api/mailgun
  clicked:
    - https://myapi.foo.com/api/mailgun
```
You can use all available kind `clicked`, `complained`, `delivered`, `opened`, `permanentFail`, `temporaryFail`, `unsubscribed` and max 3 urls by kind

For more information see mailgun api documentation https://documentation.mailgun.com/en/latest/api-webhooks.html

Execute `kubectl apply -f mg-webhooks.yaml` to configure webhooks, **deleting the object will delete the webhooks on mailgun**

## Routes
The mailgun operator can configure your routes in mailgun.

mg-routes.yaml
```yaml
apiVersion: mailgun.com/v1alpha1
kind: MailgunRoute
metadata:
  name: example-mailgunroute
spec:
  domain: mg.foo.com
  secretName: mailgun-secret
  expression: match_recipient(".*@bar.com")
  priority: 0
  actions:
    - forward("mailbox@foo.com")
    - forward("http://myapi.foo.com/messages")
```

For more information see mailgun api documentation https://documentation.mailgun.com/en/latest/api-webhooks.html

Execute `kubectl apply -f mg-routes.yaml` to configure routes, **deleting the object will delete the routes on mailgun**

# Motivations
All external services are part of your infrastructure and their configurations is almost always done by a human.
I needed to version my external service configuration and have an automatic way to provide configuration for every environment. 
Kubernetes custom resources with a mailgun operator is the perfect way to have a declarative configuration for my external services.