# IPTV Guide for JioTV Go

This is not an installation guide. If you are looking for installation instructions, 
For Android TV, see [Android TV](./android_tv.md).
For Android Phones, see [Android](./android.md).
For Linux, see [Linux](./linux-macos.md).
For Windows, see [Windows](./windows.md).

Explore the possibilities of integrating JioTV Go into your IPTV setup with these simple steps. Whether you're interested in generating playlists, setting up an Electronic Program Guide (EPG), or exploring catch-up options, we've got you covered.

## Generate M3U Playlist

JioTV Go offers a convenient M3U playlist endpoint to enhance your IPTV experience. Simply follow these steps:

1. Copy and paste the following URL into your IPTV player:

    ```
    http://localhost:5001/playlist.m3u
    ```

2. If you desire a specific quality, append the `q=` query parameter:

    ```
    http://localhost:5001/playlist.m3u?q=high
    ```

    Available options for `q` include `low`, `medium`, `high`, or their shorthand forms `l`, `m`, `h`.

3. If you would like split category on M3U playlist, append the `c=split` query parameter:

    ```
    http://localhost:5001/playlist.m3u?c=split
    ```

    This will split the playlist into categories like `Movie - Kannada`, `Movie - Malayalam`, `News - English`, etc.
	
4. If you would like to filter only specific languages on M3U playlist, append the `l=Tamil,English,Malayalam` (comma-separated) query parameter:

    ```
    http://localhost:5001/playlist.m3u?l=Tamil,English,Bengali
    ```
	
	This will filter only the specified languages (Tamil, English and Bengali).

   The playlist will be split into categories like `Movies`, `Enterntainment`, `News`, `Music`, etc. but will only contain channels in the specified languages.
	
	Available Languages to filter `Hindi, Marathi, Punjabi, Urdu, Bengali, English, Malayalam, Tamil, Gujarati, Odia, Telugu, Bhojpuri, Kannada, Assamese, Nepali, French, Other`

5. If you would like to group the M3U playlist by language only, append the `c=language` query parameter:

    ```
    http://localhost:5001/playlist.m3u?c=language
    ```

    This will group the playlist by language only, like `Hindi`, `Kannada`, `Marathi`, etc.

   Please note that either `c=split` or `c=language` can be used at a time.

6. If you would like to skip one or more genre from playlist, you can use  `sg=Educational,Lifestyle`
   ```
   http://localhost:5001/playlist.m3u?sg=Educational,Lifestyle
   ```

   This will skip all channels from provided list of genres.

For both specific quality and split category, append the `q=` and `c=` query parameters:

```
http://localhost:5001/playlist.m3u?q=high&c=split
```

You can also combine the language grouping and language filtering:

```
http://localhost:5001/playlist.m3u?c=language&l=Hindi,Kannada,Marathi
```


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

3. **Disable EPG:**
   - If you have enabled EPG via configuration, set the `epg` config value to `false`. 
   - Then run 
     
     ```
     jiotv_go epg delete
     ```

   This will delete the existing EPG file if it exists and disable EPG on the server.

## Buffering issues on IPTV Players

If you are facing buffering issues on IPTV players, try enforcing a specific quality. 

If I want to use the `high` quality, I will use the following URL:

```
http://localhost:5001/playlist.m3u?q=high
```

Where `q` can be `low`, `medium`, `high`, or `l`, `m`, `h`.

If your internet speed is low, you can use the `medium` or `low` quality.

## Catchup

Please note that JioTV Go currently does not support catch-up functionality. If you possess the expertise to implement this feature, we welcome your contribution! Open a pull request, and we appreciate your valuable input.

Enjoy the seamless integration of JioTV Go into your IPTV setup. For any queries or assistance, refer to our user-friendly documentation or connect with our community on [Telegram](/#community). Happy streaming!