#!/bin/bash

# ===========================================
# HTTPS 证书生成脚本 (generate_certs.sh)
# ===========================================

# 创建证书目录
mkdir -p certs
cd certs

echo "🔐 开始生成自签名证书..."

# 1. 生成 CA 根证书私钥
echo "📝 生成 CA 根证书私钥..."
openssl genrsa -out ca-key.pem 4096

# 2. 生成 CA 根证书
echo "📝 生成 CA 根证书..."
openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem -subj "/C=AU/ST=WA/L=Perth/O=DJJ Equipment/OU=IT Department/CN=DJJ-CA"

# 3. 生成服务端私钥
echo "📝 生成服务端私钥..."
openssl genrsa -out server-key.pem 4096

# 4. 生成服务端证书签名请求
echo "📝 生成服务端证书签名请求..."
openssl req -subj "/C=AU/ST=WA/L=Perth/O=DJJ Equipment/OU=IT Department/CN=inventory.local" -sha256 -new -key server-key.pem -out server.csr

# 5. 创建扩展文件，添加域名和IP
echo "📝 创建证书扩展配置..."
cat > server-extfile.cnf <<EOF
subjectAltName = DNS:inventory.local,DNS:localhost,IP:127.0.0.1,IP:172.27.10.254,IP:192.168.1.244
extendedKeyUsage = serverAuth
EOF

# 6. 用 CA 签发服务端证书
echo "📝 用 CA 签发服务端证书..."
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -out server-cert.pem -extfile server-extfile.cnf -CAcreateserial

# 7. 清理临时文件
rm server.csr server-extfile.cnf

# 8. 设置文件权限
chmod 400 ca-key.pem server-key.pem
chmod 444 ca.pem server-cert.pem

echo "✅ 证书生成完成！"
echo ""
echo "📋 生成的文件："
echo "  - ca.pem (CA根证书，需要导入到客户端信任列表)"
echo "  - ca-key.pem (CA私钥)"
echo "  - server-cert.pem (服务端证书)"
echo "  - server-key.pem (服务端私钥)"
echo ""
echo "🔧 接下来需要："
echo "  1. 将 ca.pem 导入到浏览器的受信任根证书列表"
echo "  2. 在 /etc/hosts 中添加: 127.0.0.1 inventory.local"
echo "  3. 启动 HTTPS 服务器"
echo ""

# 显示证书信息
echo "📋 服务端证书信息："
openssl x509 -in server-cert.pem -text -noout | grep -A 1 "Subject:"
openssl x509 -in server-cert.pem -text -noout | grep -A 3 "DNS:"

cd ..

# ===========================================
# Go 服务器 HTTPS 配置
# ===========================================

# 创建 HTTPS 服务器配置文件
cat > cmd/https_server.go <<'EOF'
// cmd/https_server.go
package main

import (
	"djj-inventory-system/config"
	"djj-inventory-system/internal/database"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/pkg/setup"
	"fmt"
	"log"

	"go.uber.org/zap/zapcore"
)

func main() {
	// 初始化配置和日志
	config.Load()
	if err := logger.Init("./logs/app.log", zapcore.DebugLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 连接数据库
	db := database.Connect()

	// 创建路由
	router := setup.NewRouter(db)

	// HTTPS 服务器配置
	certFile := "./certs/server-cert.pem"
	keyFile := "./certs/server-key.pem"
	addr := fmt.Sprintf("%s:%s", config.Get("SERVER_IP"), config.Get("SERVER_PORT"))

	log.Printf("🚀 Starting HTTPS server on https://%s", addr)
	log.Printf("🔐 Using cert: %s", certFile)
	log.Printf("🔑 Using key: %s", keyFile)

	// 启动 HTTPS 服务器
	if err := router.RunTLS(addr, certFile, keyFile); err != nil {
		panic(fmt.Sprintf("Failed to start HTTPS server: %v", err))
	}
}
EOF

# ===========================================
# Docker Compose 配置 (如果使用 Docker)
# ===========================================

cat > docker-compose.https.yml <<'EOF'
version: '3.8'

services:
  djj-inventory:
    build: .
    ports:
      - "8443:8443"  # HTTPS端口
    environment:
      - SERVER_IP=0.0.0.0
      - SERVER_PORT=8443
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=longi
      - DB_PASS=qq123456
      - DB_NAME=longinventory
    volumes:
      - ./certs:/app/certs:ro  # 只读挂载证书
      - ./logs:/app/logs
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: longinventory
      POSTGRES_USER: longi
      POSTGRES_PASSWORD: qq123456
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
EOF

# ===========================================
# Nginx 反向代理配置 (可选)
# ===========================================

cat > nginx.https.conf <<'EOF'
# nginx.https.conf
upstream djj_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name inventory.local;

    # SSL 配置
    ssl_certificate /path/to/certs/server-cert.pem;
    ssl_certificate_key /path/to/certs/server-key.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20:!aNULL:!MD5:!DSS;
    ssl_prefer_server_ciphers off;

    # 代理到后端
    location / {
        proxy_pass http://djj_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket 支持
    location /ws/ {
        proxy_pass http://djj_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name inventory.local;
    return 301 https://$server_name$request_uri;
}
EOF

# ===========================================
# 更新环境变量配置
# ===========================================

cat >> cmd/.env.https <<'EOF'
# HTTPS 配置
DB_HOST=172.27.10.254
DB_PORT=5432
DB_USER=longi
DB_PASS=qq123456
DB_NAME=longinventory
STORAGE_PATH=uploads
SERVER_PORT=8443
SERVER_IP=0.0.0.0
UPLOAD_URL=https://inventory.local:8443
CERT_FILE=./certs/server-cert.pem
KEY_FILE=./certs/server-key.pem
EOF

# ===========================================
# 客户端证书导入说明
# ===========================================

cat > CERTIFICATE_SETUP.md <<'EOF'
# 🔐 HTTPS 证书设置指南

## 1. 生成证书
```bash
# 运行证书生成脚本
chmod +x generate_certs.sh
./generate_certs.sh
```

## 2. 配置域名解析
在 `/etc/hosts` (Linux/Mac) 或 `C:\Windows\System32\drivers\etc\hosts` (Windows) 中添加：
```
127.0.0.1 inventory.local
172.27.10.254 inventory.local
```

## 3. 导入 CA 证书到浏览器

### Chrome/Edge:
1. 打开设置 → 隐私和安全 → 安全 → 管理证书
2. 点击"受信任的根证书颁发机构" → 导入
3. 选择 `certs/ca.pem` 文件
4. 重启浏览器

### Firefox:
1. 打开设置 → 隐私与安全 → 证书 → 查看证书
2. 点击"证书颁发机构" → 导入
3. 选择 `certs/ca.pem` 文件
4. 勾选"信任此 CA 来标识网站"

### Safari (Mac):
1. 双击 `certs/ca.pem` 文件
2. 在钥匙串访问中找到导入的证书
3. 右键 → 显示简介 → 信任 → 选择"始终信任"

## 4. 启动 HTTPS 服务器
```bash
# 方式1: 直接运行
go run cmd/https_server.go

# 方式2: 编译后运行
go build -o bin/https_server cmd/https_server.go
./bin/https_server

# 方式3: 使用 Docker
docker-compose -f docker-compose.https.yml up
```

## 5. 测试连接
```bash
# 测试 HTTPS 连接
curl -k https://inventory.local:8443/api/auth/roles

# 测试 WSS 连接 (使用 websocat 工具)
websocat wss://inventory.local:8443/ws/inventory
```

## 6. 前端配置更新
更新前端配置，将所有 HTTP 请求改为 HTTPS：
```javascript
// 旧配置
const API_BASE_URL = 'http://172.27.10.254:8080'
const WS_BASE_URL = 'ws://172.27.10.254:8080'

// 新配置
const API_BASE_URL = 'https://inventory.local:8443'
const WS_BASE_URL = 'wss://inventory.local:8443'
```

## 7. 故障排除

### 问题: "您的连接不是私密连接"
- 确保已正确导入 CA 证书
- 检查域名是否正确配置
- 尝试清除浏览器缓存

### 问题: WebSocket 连接失败
- 确保使用 `wss://` 而不是 `ws://`
- 检查防火墙设置
- 验证证书中包含正确的域名

### 问题: CORS 错误
更新 Gin 的 CORS 配置：
```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://inventory.local:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
    AllowCredentials: true,
}))
```

## 8. 生产环境建议
- 使用真实的 SSL 证书（Let's Encrypt 或购买的证书）
- 配置 HSTS 头
- 启用 OCSP Stapling
- 定期更新证书
- 使用 CDN 和负载均衡器

## 9. 证书续期
证书有效期为 365 天，到期前需要重新生成：
```bash
# 重新运行生成脚本
./generate_certs.sh

# 重启服务器
systemctl restart djj-inventory
```
EOF

echo ""
echo "🎉 HTTPS 配置文件已生成完成！"
echo ""
echo "📁 生成的文件："
echo "  - generate_certs.sh (证书生成脚本)"
echo "  - cmd/https_server.go (HTTPS 服务器代码)"
echo "  - docker-compose.https.yml (Docker 配置)"
echo "  - nginx.https.conf (Nginx 配置)"
echo "  - cmd/.env.https (HTTPS 环境变量)"
echo "  - CERTIFICATE_SETUP.md (详细设置指南)"
echo ""
echo "🚀 下一步："
echo "  1. 运行: chmod +x generate_certs.sh && ./generate_certs.sh"
echo "  2. 导入 CA 证书到浏览器"
echo "  3. 配置 /etc/hosts"
echo "  4. 运行: go run cmd/https_server.go"
echo "  5. 访问: https://inventory.local:8443"