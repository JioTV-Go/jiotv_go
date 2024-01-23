# Miscellaneous

## Enforce a specific quality

If I want to use the `high` quality, I will use the following URL:

```
http://localhost:5001/playlist.m3u?q=high
```

Where `high` can be replaced with `low`, `medium`, `high`, or `l`, `m`, `h`.

If your internet speed is low, you can use the `medium` or `low` quality.

## Check if you can run JioTV Go in your VPS/Cloud Server

JioTV APIs are geo-restricted (India only) and IP-restricted (residential IPs only). So you need to check if you can run JioTV Go in your VPS/Cloud Server.

Execute this command in VPS,

```bash
curl -v "https://jiotv.data.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?os=android&devicetype=phone&usertype=tvYR7NSNn7rymo3F&version=285"
```

If you get a full JSON response then it will work.
Otherwise if you don't get any response it won't.

You can use residential proxies to bypass this restriction. Read the [Proxy](./cloud_hosting.md#residential-proxy) section in the [Cloud Hosting](./cloud_hosting.md) page for more information.
