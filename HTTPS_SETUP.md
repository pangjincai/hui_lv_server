# HTTPS Setup Guide (CentOS 7 + Nginx)

This guide will help you secure `hapi.dabeiliu.com` with a free SSL certificate from Let's Encrypt using Certbot.

## Prerequisites
- You must have **root** access to the server.
- The domain `hapi.dabeiliu.com` must point to your server's IP.
- Nginx should be installed and running (as per previous deployment steps).

## 1. Install Certbot
Certbot is the tool that automatically obtains and renews certificates.

```bash
# Install EPEL repository (if not already installed)
yum install -y epel-release

# Install Certbot and the Nginx plugin
yum install -y certbot python2-certbot-nginx
```

## 2. Obtain Certificate
Run Certbot to get the certificate. It will automatically edit your Nginx configuration to enable HTTPS.

```bash
certbot --nginx -d hapi.dabeiliu.com
```

- When prompted for an email address, enter your email (for renewal notices).
- Agree to the Terms of Service.
- If asked to redirect HTTP traffic to HTTPS, choose **2 (Redirect)**. This ensures all traffic is secure.

## 3. Verify Auto-Renewal
Certificates expire after 90 days. Certbot usually sets up a timer to renew them automatically. You can test this:

```bash
certbot renew --dry-run
```
If this command completes without errors, your auto-renewal is set up correctly.

## 4. Restart Nginx
To make sure everything is applied:

```bash
systemctl restart nginx
```

## 5. Verify HTTPS
Open `https://hapi.dabeiliu.com/api/rates/latest` in your browser. You should see the lock icon indicating a secure connection.
