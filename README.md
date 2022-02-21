# arcade

Arcade is meant to run as a sidecar to generate authorization tokens and make them retrievable through a simple authenticated API. If the token has a defined expiration time, Arcade is set to cache the token for 90% of its lifetime.

## Providers

Arcade supports the following authorization token providers:

1. Google
2. Microsoft
3. Rancher

Token provider configuration files containing the credentials are placed in the `ARCADE_CONFIG_DIRECTORY` directory (default location is `/secret/arcade/providers`)

### Google

Using google's [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity), Arcade retrieves the token of the active GCP account.

```json5
{
  "type": "", // Required, set to 'google'
  "name": ""  // Required, set to a unique name identifying this token provider
}
```

### Microsoft

Use this JSON structure to configure a Microsoft token provider

```json5
{
  "type": "",          // Required, set to 'microsoft'
  "name": "",          // Required, set to a unique name identifying this token provider
  "loginEndpoint": "", // Reqoured, set to the 'login' endpoint, such as https://login.microsoftonline.com/someone.onmicrosoft.com/oauth2/token
  "clientId": "",      // Required, set to your Microsoft Client ID
  "clientSecret": "",  // Required, set to your Microsoft Client Secret
  "resource": ""       // Optional, set to the resource you are requesting, such as 'https://graph.microsoft.com'
}
```

### Rancher

Use this JSON structure to configure a Rancher token provider

```json5
{
  "type": "",     // Required, set to 'rancher'
  "name": "",     // Required, set to a unique name identifying this token provider
  "url": "",      // Reqoured, set to the 'login' endpoint of your Rancher instance
  "username": "", // Required, set to your Rancher username
  "password": "", // Required, set to your Rancher upassword
  "rootCA": ""    // Optional, set to a certificate to add to the trusted root CAs
}
```

Rancher kubeconfig tokens have an expiration time and Arcade will cache the token until it has expired before calling Rancher for a new one.

## Run Locally

Prerequisites:
- the [gcloud CLI](https://cloud.google.com/sdk/gcloud#download_and_install_the) must be installed
- you'll need [default credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/login) set up

Run the following commands to build and generate a token.

### Build

```bash
go build cmd/arcade/arcade.go
```

### Run

```bash
export ARCADE_API_KEY=test
export ARCADE_CONFIG_DIRECTORY=/tmp/arcade

[[ ! -d ${ARCADE_CONFIG_DIRECTORY} ]] && mkdir ${ARCADE_CONFIG_DIRECTORY}

echo '{
  "type": "google",
  "name": "google"
}' > ${ARCADE_CONFIG_DIRECTORY}/google.json

echo '{
  "type": "rancher",
  "name": "rancher.example.com",
  "url": "https://rancher.example.com/v3/activeDirectoryProviders/activedirectory?action=login",
  "username": "<YOUR_USERNAME>",
  "password": "<YOUR_PASSWORD>"
}' > ${ARCADE_CONFIG_DIRECTORY}/rancher.json

echo '{
  "type": "microsoft",
  "name": "microsoftonline",
  "loginEndpoint": "https://login.microsoftonline.com/someone.onmicrosoft.com/oauth2/token",
  "clientId": "<YOUR_CLIENT_ID>",
  "clientSecret": "<YOUR_CLIENT_SECRET>",
  "resource": "https://graph.microsoft.com"
}' > ${ARCADE_CONFIG_DIRECTORY}/microsoft.json

./arcade
```

### Test

**Google**
```bash
curl localhost:1982/tokens?provider=google -H "Api-Key: test"
```
The default token provider is google, so this is equivalent to the call above
```bash
curl localhost:1982/tokens -H "Api-Key: test"
```
**Microsoft**
```bash
curl localhost:1982/tokens?provider=microsoftonline -H "Api-Key: test"
```
**Rancher**
```bash
curl localhost:1982/tokens?provider=rancher.example.com -H "Api-Key: test"
```
