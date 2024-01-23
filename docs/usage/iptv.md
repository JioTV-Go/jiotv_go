# IPTV Guide for JioTV Go

Explore the possibilities of integrating JioTV Go into your IPTV setup with these simple steps. Whether you're interested in generating playlists, setting up an Electronic Program Guide (EPG), or exploring catch-up options, we've got you covered.

## Generate M3U Playlist

JioTV Go offers a convenient M3U playlist endpoint to enhance your IPTV experience. Simply follow these steps:

1. Copy and paste the following URL into your IPTV player:

    ```
    http://localhost:5001/playlist.m3u
    ```

2. If you desire a specific quality, append the `q` query parameter:

    ```
    http://localhost:5001/playlist.m3u?q=high
    ```

    Available options for `q` include `low`, `medium`, `high`, or their shorthand forms `l`, `m`, `h`.

## Electronic Program Guide (EPG)

Take advantage of JioTV Go's Electronic Program Guide to enrich your IPTV setup. Follow these steps:

1. **Enable EPG:**
   - Set the `epg` config value to `true`. For detailed instructions, refer to the [Config](./config.md#epg-electronic-program-guide) page. Or you can also use the `epg generate` command. For additional details, consult the [EPG Command](./usage.md#3-epg-command) section.

2. **Access EPG in Your IPTV Player:**
   - Once enabled, wait a few minutes for EPG generation.
   - Use the following URL in your IPTV player: 
   
      ```
      http://localhost:5001/epg.xml.gz
      ```

   EPG updates every 24 hours, providing program information for a 2-day duration.

## Catchup

Please note that JioTV Go currently does not support catch-up functionality. If you possess the expertise to implement this feature, we welcome your contribution! Open a pull request, and we appreciate your valuable input.

Enjoy the seamless integration of JioTV Go into your IPTV setup. For any queries or assistance, refer to our user-friendly documentation or connect with our support team. Happy streaming!