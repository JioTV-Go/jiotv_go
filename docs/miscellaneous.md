# Miscellaneous

## Buffering issues on IPTV Players

If you are facing buffering issues on IPTV players, try enforcing a specific quality. 

If I want to use the `high` quality, I will use the following URL:

```
http://localhost:5001/playlist.m3u?q=high
```

Where `q` can be `low`, `medium`, `high`, or `l`, `m`, `h`.

If your internet speed is low, you can use the `medium` or `low` quality.

## Third-party EPG Providers

For the individuals having issue with EPG in Android TV, they can use mitthu786's JioTV EPG as a source in IPTV Settings.

Link: https://github.com/mitthu786/tvepg

## Check if you can run JioTV Go in your VPS/Cloud Server

Execute this command in VPS,

```bash
curl -v "https://jiotv.data.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?os=android&devicetype=phone&usertype=tvYR7NSNn7rymo3F&version=285"
```

If you get a full JSON response then it will work.
Otherwise if you don't get any response it won't.