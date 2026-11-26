# API

The field `config_serve_addr` expose some API for configuration reload and management.

Body format are all **plain text**

- `/api/text` GET

Get this API document with plain text format

- `/api/html` GET

Get this API document with html format (for browser visit)

- `/info` GET
  
See current running config

- `/modify/accept?ip=192.168.1.1` POST

**When default policy is drop**, use this to allow traffic from 192.168.1.1 (or any other ip)

- `/modify/drop?ip=192.168.1.1` POST

**When default policy is accept**, use this to drop traffic from 192.168.1.1 (or any other ip)

- `/save` POST

Save current running configuration into `config.json`

- `/reload` POST

Reload the configuration (from config.json file)

- `/shutdown` POST

Shutdown service

---

After you use the API to modify, policy will not be saved into config file
use `/save` as need
