# ddns

# Purpose

This project essentially is continuously running code that updates your primary DNS record in Cloudflare.

While some might argue that a cronjob could be better suited for this, I believe that the container uses such a small amount of resources
that you can only benefit by having it run all the time.

# Using this program

## Cloudflare

Some environment variables are required:

* `ACCOUNT_TOKEN` - The API token for your Cloudflare account.

OR

* `API_KEY` - The API Key for the Cloudflare account
* `ACCOUNT_EMAIL` - The Cloudflare account email address

## Program Arguments

1. `--zone-name`: The name of the zone you wish to modify.
2. `--record-name`: The name of the record you wish to modify inside your zone.
3. `--provider`: The provider you wish to use can be: `ipify`, `icanhazip`, `icanhaz`. Anything other than these three will be considered as "random"
4. `--ticker`: Directly uses Golang's implementation of `time.Duration`. Default is "3m", you can set this to whatever you'd like.
5. `--create-missing`: This will create the missing DNS record in the target if set to true, it is `false` by default.
6. `--record-ttl`: This will set the DNS record to this TTL. `300` by default.
7. `--proxied`: Boolean whether the record in cloudflare should be set to `proxy`, if true a lookup will return Cloudflare ips. `true` by default.

## Example

[Helm Release Example](https://github.com/larivierec/home-cluster/blob/main/kubernetes/main/apps/networking/ddns/app/helm-release.yaml)