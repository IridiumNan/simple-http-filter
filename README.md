# Simple HTTP Filter

This is a simple http traffic filter (reverse proxy) to expose service with rule filter.

## USAGE

- Step 1

Generate default configuration file.

```bash
./hf config
```

- Step 2

Modify the configuration file  then start service

```bash
vim config.json

# then start
./hf
```

- Step 3 manage (or reload) by API

```bash
# the config_serve_addr load from config.json
# this endpoint return API usage
curl http://config_serve_addr/api/text

# try to reload and save
curl http://config_serve_addr/reload

curl http://config_serve_addr/save
```
