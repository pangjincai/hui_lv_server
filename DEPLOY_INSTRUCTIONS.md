# Deployment Instructions

## 1. Prepare Server Directory
SSH into your CentOS server and create the application directory:
```bash
mkdir -p /data/hui_lv
```

## 2. Upload Files
Upload the following files from the `deploy` folder and your `config.json` to `/data/hui_lv` on the server:
- `hui_lv_server_linux` (The binary)
- `..\config.json` (Your configuration file)
- `hui_lv.service` (Systemd service)
- `hapi.dabeiliu.com.conf` (Nginx config)

You can use `scp` (run this from your local machine):
```bash
scp deploy/hui_lv_server_linux deploy/hui_lv.service deploy/hapi.dabeiliu.com.conf config.json root@<YOUR_SERVER_IP>:/data/hui_lv/
```

## 3. Setup Systemd Service
Move the service file to the systemd directory and start the service:
```bash
# Correct permissions
chmod +x /data/hui_lv/hui_lv_server_linux

# Install Service
mv /data/hui_lv/hui_lv.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable hui_lv
systemctl start hui_lv

# Check status
systemctl status hui_lv
```

## 4. Setup Nginx
Move the Nginx config and reload:
```bash
mv /data/hui_lv/hapi.dabeiliu.com.conf /etc/nginx/conf.d/
nginx -t
nginx -s reload
```

## 5. Verify
Visit `http://hapi.dabeiliu.com/api/rates/latest` to verify the API is reachable.
