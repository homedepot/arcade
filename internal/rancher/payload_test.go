package rancher_test

const payloadKubeconfigToken = `{
  "authProvider": "local",
  "baseType": "token",
  "clusterId": null,
  "created": "2021-03-25T10:38:18Z",
  "createdTS": 1616668698000,
  "creatorId": null,
  "current": false,
  "description": "Kubeconfig token",
  "enabled": true,
  "expired": false,
  "expiresAt": "2021-03-25T20:38:18Z",
  "groupPrincipals": null,
  "id": "kubeconfig-u-i76rfanbw5",
  "isDerived": true,
  "labels": {
    "authn.management.cattle.io/kind": "kubeconfig",
    "authn.management.cattle.io/token-userId": "u-i76rfanbw5",
    "cattle.io/creator": "norman"
  },
  "lastUpdateTime": "",
  "links": {
    "self": "https://rancher.example.com/v3/tokens/kubeconfig-u-i76rfanbw5"
  },
  "name": "kubeconfig-u-i76rfanbw5",
  "token": "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9",
  "ttl": 36000000,
  "type": "token",
  "userId": "u-i76rfanbw5",
  "userPrincipal": "map[metadata:map[creationTimestamp:<nil>]]",
  "uuid": "bf897e53-3cb3-49d0-af44-09343f75ec2e"
}`

const payloadKubeconfigTokenCached = `{
  "created": "2021-03-25T10:38:18Z",
	"expiresAt": "9999-12-31T00:00:00Z",
	"token": "fake.token.cached"
}`

const payloadKubeconfigTokenAnother = `{
	"created": "2012-12-31T00:00:00Z",
	"expiresAt": "2999-01-01T00:00:00Z",
	"token": "another.token"
}`
