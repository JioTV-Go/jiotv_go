
# Initial Setup

open terminal and create directory to your Jio TV Go docker directory,
Create **docker-compose.yml** in the directory as follows:

```
services:
  JioTVGo:
    container_name: 'JioTV_Go'
    image: ghcr.io/jiotv-go/jiotv_go:latest
    ports:
      - 5002:5001
    env_file:
      - ./.env
```

Create **.env** file in the same directory where **docker-compose.yml** is created
All possible variables for this file are specified at <a href="https://jiotv_go.rabil.me/config.html">Config</a> page
```
debug = false
disable_ts_handler = false
drm = false
epg = false
disable_url_encryption = false
```
Jio TV Go will be accessible at following URL:
http://localhost:5002
# Run the JioTV Go
Execute following commands to run Jio TV Go in docker after doing initial setup,
open terminal and change directory to your Jio TV Go docker Folder,
to download the Jio TV Go docker image execute:
```
docker compose pull
```
then to start the container
```
docker compose up -d
```

# Update JioTV Go docker image
Execute following commands to update Jio TV Go,
First stop running JioTV Go container
```
docker compose stop
```
then to download the Jio TV Go docker image execute:
```
docker compose pull
```
then start the container
```
docker compose up -d
```
