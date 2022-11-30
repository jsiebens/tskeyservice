# tskeyservice

This lightweight service exchanges OIDC token from trusted issuer for a short-lived, one-time use [Tailscale](https://tailscale.com) [Auth Token](https://tailscale.com/kb/1085/auth-keys/).

## Why?

The idea for this service was born when connecting a GitHub Action to my Tailscale network.
While this is typically done by creating an ephemeral auth key, add it to the secrets of the workflow.
Although such Tailscale auth keys have an expiration by default, I always try to avoid "static" secrets.

## How?

GitHub Actions has support for OpenID Connect (OIDC) tokens, allowing your workflows to exchange short-lived tokens from e.g. your cloud provider.
This service is such an implementation and creates ephemeral, short-lived, one-time auth keys for your Tailscale network.

## Configuration

Configuration is done by settings environment variables:

- TS_TAILNET: the name of your tailnet
- TS_API_KEY: a Tailscale API key
- TS_KEYS_ISSUER: a trusted OIDC Issuer url (e.g. https://token.actions.githubusercontent.com)
- TS_KEYS_TAGS: comma-separated list of ACL tags
- TS_KEYS_BEXPR: a boolean expression to filter OIDC tokens (based on claims) when creating auth keys

The last setting is quit important as it allows you to filter OIDC tokens and only creating auth keys when the token has some certain claims.

Example:

In case of GitHub Action, the issued token has a `repository` claim.
The following expression will only create auth keys for a workflow from this repository:

```shell
export TS_KEYS_BEXPR='repository == "jsiebens/tskeys-example"'
```

The boolean expression is implemented using the [HashiCorp go-bexpr](https://github.com/hashicorp/go-bexpr) library

## Deployment

When using in GitHub Actions, this service should be publicly available. E.g. on Google Cloud Run of [fly.io](https://fly.io).

A Docker image is available at `ghcr.io/jsiebens/tskeyservice`

## Example workflow

```yaml
name: GitHub Action Sample

on:
  workflow_dispatch:

permissions:
  id-token: write

jobs:
  sample:
    runs-on: ubuntu-latest
    steps:

      - name: Get Tailscale key
        shell: bash
        env:
          TSKEYSERVICE_URL: "<your tskeservice url e.g. https://tskeys.example.com/key>"
        run: |
          OIDC_TOKEN=$(curl -sLS "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=tskeyservice" -H "User-Agent: actions/oidc-client" -H "Authorization: Bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" | jq -j '.value')
          TS_KEY=$(curl -sLS $TSKEYSERVICE_URL -H "User-Agent: actions/oidc-client" -H "Authorization: Bearer $OIDC_TOKEN" | jq -j '.key')
          echo "TAILSCALE_AUTHKEY=$TS_KEY" >> $GITHUB_ENV
          
      - name: Tailscale
        uses: tailscale/github-action@main
        with:
          authkey: ${{ env.TAILSCALE_AUTHKEY }}
```

## Alternatives

- [vault-plugin-tailscale](https://github.com/davidsbond/vault-plugin-tailscale)
