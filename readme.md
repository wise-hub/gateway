go run ./cmd


go build -o ./bin/api-gateway-ws ./cmd

CGO_ENABLED=1 GOARCH=arm64 GOOS=darwin go build -o ./bin/api-gateway-ws ./cmd
CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o ./bin/api-gateway-ws ./cmd


curl --location 'http://localhost:8200/api/public/accounts' \
--header 'Authorization: generated_token_from_db'


curl --location 'http://localhost:8200/api/public/login' \
--header 'Content-Type: application/json' \
--data '{
  "user": "uuu",
  "pass": "ppp"
}'


curl --location 'http://localhost:8200/api/cache'


ab -n 10000 -c 100 -H "Authorization: generated_token_from_db" http://localhost:8200/api/cache

bombardier -c 200 -n 20000 -l -m POST -H "Content-Type: application/json"  -b '{"user": "your_username", "pass": "your_password"}' --timeout 30s http://localhost:8200/public/login

bombardier -c 50 -n 20000 -l -m POST \
-H "Content-Type: application/json" \
-b '{"user": "your_username", "pass": "your_password"}' \
--timeout 30s http://localhost:8200/public/login



nginx config example 

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    server {
        listen 443 ssl http2;
        server_name xxx.bg;

        ssl_certificate /etc/ssl/certs/cert.crt;
        ssl_certificate_key /etc/ssl/private/key.key;

        location / {
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # HTTP/2 plain text (h2c) connection to the upstream
            proxy_pass http://upstream_h2c;

            proxy_http_version 1.1;
            proxy_set_header Connection "";

            proxy_connect_timeout 10s;
            proxy_read_timeout 10s;
            proxy_send_timeout 10s;
        }
    }

    upstream upstream_h2c {
        server localhost:8080;
        keepalive 64;
    }
}

server {
    listen 80;
    server_name your_domain.com;

    location / {
        return 301 https://$host$request_uri;
    }
}



======
mac

sudo sysctl -w kern.ipc.somaxconn=1024
sudo sysctl -w net.inet.tcp.msl=3000
sudo sysctl -w net.inet.tcp.sendspace=65536
sudo sysctl -w net.inet.tcp.recvspace=65536
ulimit -n 65536



redhat

# Temporarily set the maximum number of queued connections
sudo sysctl -w net.core.somaxconn=1024

# Set the TCP FIN timeout to reduce the time connections spend in TIME_WAIT state
sudo sysctl -w net.ipv4.tcp_fin_timeout=30

# Set the TCP send and receive buffer sizes
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216"
sudo sysctl -w net.ipv4.tcp_rmem="4096 65536 16777216"

# Temporarily increase the file descriptor limit
ulimit -n 65536

# Make changes permanent by appending to /etc/sysctl.conf
echo "net.core.somaxconn=1024" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_fin_timeout=30" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_wmem=4096 65536 16777216" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_rmem=4096 65536 16777216" | sudo tee -a /etc/sysctl.conf

# Make the file descriptor limit permanent
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Ensure PAM limits are enforced
echo "session required pam_limits.so" | sudo tee -a /etc/pam.d/common-session
echo "session required pam_limits.so" | sudo tee -a /etc/pam.d/common-session-noninteractive

# Apply all sysctl changes immediately
sudo sysctl -p