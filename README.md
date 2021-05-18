# arcade

Arcade is meant to run as a sidecar to generate authorization tokens and make them retrievable through a simple authenticated API. If the token has a defined expiration time, Arcade is set to cache the token for 90% of its lifetime.

## Providers

Arcade supports the following authorization token providers:

1. Google
2. Microsoft
3. Rancher

### Google

Using google's [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity), Arcade retrieves the token of the active GCP account.

### Microsoft

Use these variables to configure Microsoft.

```sh
MICROSOFT_ENABLED=        # Set to TRUE if Microsoft is a supported token provider
MICROSOFT_LOGIN_ENDPOINT= # Set to the 'login' endpoint, such as https://login.microsoftonline.com/someone.onmicrosoft.com/oauth2/token
MICROSOFT_CLIENT_ID=      # Set to your Microsoft client ID
MICROSOFT_CLIENT_SECRET=  # Set to your Microsoft client secret
MICROSOFT_RESOURCE=       # Set to the resource you are requesting, such as 'https://graph.microsoft.com'
```

### Rancher

Use these variables to configure Rancher

```sh
RANCHER_ENABLED=  # Set to TRUE if Rancher is a supported token provider
RANCHER_URL=      # Set to the 'login' endpoint of your Rancher instance
RANCHER_USERNAME= # Set to your Rancher username
RANCHER_PASSWORD= # Set to your Rancher password
```

Rancher kubeconfig tokens have an expiration time and Arcade will cache the token until it has expired before calling Rancher for a new one.

## Run Locally

You'll need default credentials set up. Run the following commands to build and generate a token.

### Build

```bash
go build cmd/arcade/arcade.go
```

### Run

```bash
export ARCADE_API_KEY=test
export RANCHER_ENABLED=TRUE
export RANCHER_URL=https://rancher.example.com/v3/activeDirectoryProviders/activedirectory?action=login
export RANCHER_USERNAME=myUsername
export RANCHER_PASSWORD=myPassword
export MICROSOFT_ENABLED=TRUE
export MICROSOFT_LOGIN_ENDPOINT=https://login.microsoftonline.com/someone.onmicrosoft.com/oauth2/token
export MICROSOFT_CLIENT_ID=<YOUR_CLIENT_ID>
export MICROSOFT_CLIENT_SECRET=<YOUR_CLIENT_SECRET>
export MICROSOFT_RESOURCE=https://graph.microsoft.com
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
curl localhost:1982/tokens?provider=microsoft -H "Api-Key: test"
```
**Rancher**
```bash
curl localhost:1982/tokens?provider=rancher -H "Api-Key: test"
```
