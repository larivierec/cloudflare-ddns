# cloudflare-ddns

# Purpose

This project essentially is continuously running code that updates your primary DNS record in Cloudflare.

While some might argue that a cronjob could be better suited for this, I believe that the container uses such a small amount of resources
that you can only benefit by having it run all the time.

# Using this program

Some environment variables are required:

* `ACCOUNT_TOKEN` - The API token for your Cloudflare account.

OR

* `API_KEY` - The API Key for the Cloudflare account
* `ACCOUNT_EMAIL` - The Cloudflare account email address

## Program Arguments

1. `--zone-name`: The name of the zone you wish to modify.
2. `--record-name`: The name of the record you wish to modify inside your zone.
3. `--provider`: The provider you wish to use can be: `ipify`, `icanhazip`, `icanhaz`. Anything other than these three will be considered as "random"

## Example

[Helm Release - Flux Example](https://github.com/larivierec/home-cluster/blob/main/kubernetes/apps/networking/ddns/app/helm-release.yaml)