# Cloud Hosting

JioTV Go can't be hosted on a typical web hosting service as JioTV API has geo-restrictions (India only) and IP-restrictions (Residential IPs only). This means that you can't host JioTV Go on a VPS or a cloud server. However, you can host it on your home server or a Raspberry Pi.

You can host JioTV Go on a VPS or a cloud server if you use a residential proxy. However, residential proxies are expensive and not worth it.

## Residential Proxy

You can use a residential proxy to bypass the geo-restrictions and IP-restrictions. 

In the [proxy config](./config.md#proxy) page, you can set the `proxy` config value to the proxy URL or set the `JIOTV_PROXY` environment variable to the proxy URL. 

Example `config.toml` value:

```toml
proxy = "http://username:password@proxy_ip:proxy_port"
```

## Home Server

You can host JioTV Go on your home server. This is the recommended way to host JioTV Go. You can use a Raspberry Pi / old phone / any other device to host JioTV Go.

## Exposing Your Home Server to the Internet

If you want to expose your home server to the internet, we recommend using [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/). It's free and easy to set up. You can also use [ngrok](https://ngrok.com/) or [serveo](https://serveo.net/) to expose your home server to the internet.

## Miscellaneous

In cases where remote server permissions prevent the creation of the `jiotv_credentials_v2.json` file, follow these steps:

1. On your local machine, log in to JioTV Go to generate your `jiotv_credentials_v2.json`.

2. Configure the following environment variables on your remote server:
   - `JIOTV_SSO_TOKEN` - The `ssoToken` from the credentials file.
   - `JIOTV_CRM` - The `crm` from the credentials file.
   - `JIOTV_UNIQUE_ID` - The `uniqueId` from the credentials file.

With these environment variables set, the credentials file will be bypassed.
