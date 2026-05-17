#!/bin/bash
set -e

# ============================================================
# 智算引擎 一键部署脚本（Debian 12 / 8G内存 / 美西服务器）
# 系统调优 + Docker安装 + 源码构建 + 自动配置
# ============================================================
# 使用方法：
#   bash deploy.sh
# ============================================================

echo ""
echo "============================================"
echo "  智算引擎 一键部署脚本"
echo "============================================"
echo ""

ACCESS_MODE="ip"
DOMAIN=""

if [ -t 0 ]; then
  echo "请选择访问方式："
  echo "  1) IP + 端口访问（http://服务器IP:8080）"
  echo "  2) 域名 HTTPS 访问（https://你的域名，自动配置 Caddy）"
  read -r -p "请输入 1 或 2，默认 1: " ACCESS_CHOICE
  if [ "$ACCESS_CHOICE" = "2" ]; then
    ACCESS_MODE="domain"
    read -r -p "请输入已解析到本服务器的域名，例如 api.example.com: " DOMAIN
    if [ -z "$DOMAIN" ]; then
      echo "域名不能为空。"
      exit 1
    fi
  fi
else
  if [ -n "${ZSYQ_DOMAIN:-}" ]; then
    ACCESS_MODE="domain"
    DOMAIN="$ZSYQ_DOMAIN"
  fi
fi

if [ "$ACCESS_MODE" = "domain" ]; then
  echo "访问方式: 域名 HTTPS (${DOMAIN})"
else
  echo "访问方式: IP + 端口"
fi
echo ""

# ----------------------------------------------------------
# 1. 系统更新 + 清理
# ----------------------------------------------------------
echo "[1/8] 系统更新..."
apt update && apt upgrade -y
apt purge -y exim4* rpcbind avahi-daemon nfs-common 2>/dev/null || true
apt autoremove -y

# ----------------------------------------------------------
# 2. Swap + 内核调优 + 文件描述符
# ----------------------------------------------------------
echo "[2/8] 系统调优..."

# Swap
if [ ! -f /swapfile ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo "/swapfile none swap sw 0 0" >> /etc/fstab
fi

# 内核参数
sed -i '/^net\.netfilter\.nf_conntrack_/d' /etc/sysctl.conf /etc/sysctl.d/*.conf 2>/dev/null || true
cat > /etc/sysctl.d/99-zsyq.conf << 'EOF'
vm.swappiness=10
fs.file-max=65535
net.core.somaxconn=32768
net.core.netdev_max_backlog=16384
net.ipv4.tcp_max_syn_backlog=8192
net.ipv4.tcp_tw_reuse=1
net.ipv4.tcp_fin_timeout=15
net.ipv4.tcp_keepalive_time=300
net.ipv4.tcp_keepalive_intvl=15
net.ipv4.tcp_keepalive_probes=5
net.ipv4.ip_local_port_range=1024 65535
EOF
sysctl -p /etc/sysctl.d/99-zsyq.conf

# 文件描述符
grep -q "nofile 65535" /etc/security/limits.conf || cat >> /etc/security/limits.conf << 'EOF'
* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
EOF

# ----------------------------------------------------------
# 3. 安装Docker
# ----------------------------------------------------------
echo "[3/8] 安装Docker..."

apt install -y ca-certificates curl gnupg git vim htop

install -m 0755 -d /etc/apt/keyrings

curl -fsSL https://download.docker.com/linux/debian/gpg \
  | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable" \
  > /etc/apt/sources.list.d/docker.list

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Docker调优
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 65535,
      "Soft": 65535
    }
  },
  "live-restore": true,
  "userland-proxy": false
}
EOF
systemctl restart docker
systemctl enable docker

# ----------------------------------------------------------
# 4. 克隆源码
# ----------------------------------------------------------
echo "[4/8] 克隆源码..."

if [ ! -d /opt/zsyq ]; then
  git clone https://github.com/nameyzh-netizen/zsyq.git /opt/zsyq
else
  git -C /opt/zsyq pull
fi

cd /opt/zsyq/deploy

# ----------------------------------------------------------
# 5. 配置.env
# ----------------------------------------------------------
echo "[5/8] 配置环境..."

cp .env.example .env

# 生成密钥
PG_PASS=$(openssl rand -hex 16)
JWT=$(openssl rand -hex 32)
TOTP=$(openssl rand -hex 32)

# 写入密钥
sed -i "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${PG_PASS}|" .env
sed -i "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD=admin123|" .env
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${JWT}|" .env
sed -i "s|^TOTP_ENCRYPTION_KEY=.*|TOTP_ENCRYPTION_KEY=${TOTP}|" .env

# 8G内存适配
sed -i "s|^GOMEMLIMIT=.*|GOMEMLIMIT=700MiB|" .env
sed -i "s|^ZSYQ_MEMORY_LIMIT=.*|ZSYQ_MEMORY_LIMIT=1536M|" .env
sed -i "s|^POSTGRES_MEMORY_LIMIT=.*|POSTGRES_MEMORY_LIMIT=3G|" .env
sed -i "s|^POSTGRES_SHARED_BUFFERS=.*|POSTGRES_SHARED_BUFFERS=1GB|" .env
sed -i "s|^POSTGRES_EFFECTIVE_CACHE_SIZE=.*|POSTGRES_EFFECTIVE_CACHE_SIZE=2GB|" .env
sed -i "s|^DATABASE_MAX_OPEN_CONNS=.*|DATABASE_MAX_OPEN_CONNS=30|" .env
sed -i "s|^DATABASE_MAX_IDLE_CONNS=.*|DATABASE_MAX_IDLE_CONNS=8|" .env

if [ "$ACCESS_MODE" = "domain" ]; then
  sed -i "s|^BIND_HOST=.*|BIND_HOST=127.0.0.1|" .env
else
  sed -i "s|^BIND_HOST=.*|BIND_HOST=0.0.0.0|" .env
fi

# ----------------------------------------------------------
# 6. 创建数据目录
# ----------------------------------------------------------
echo "[6/8] 创建数据目录..."

mkdir -p data postgres_data redis_data

# ----------------------------------------------------------
# 7. 源码构建并启动
# ----------------------------------------------------------
echo "[7/8] 源码构建启动（首次约10-20分钟）..."

docker compose -f docker-compose.build.yml up -d --build

if [ "$ACCESS_MODE" = "domain" ]; then
  echo "[8/9] 配置域名 HTTPS..."
  apt install -y caddy
  cat > /etc/caddy/Caddyfile << EOF
${DOMAIN} {
    reverse_proxy 127.0.0.1:8080
}
EOF
  caddy fmt --overwrite /etc/caddy/Caddyfile >/dev/null
  systemctl enable caddy >/dev/null
  systemctl restart caddy
fi

# ----------------------------------------------------------
# 8. 完成
# ----------------------------------------------------------
if [ "$ACCESS_MODE" = "domain" ]; then
  echo "[9/9] 验证..."
else
  echo "[8/8] 验证..."
fi

sleep 5
docker compose -f docker-compose.build.yml ps

SERVER_IP=$(hostname -I | awk '{print $1}')
if [ "$ACCESS_MODE" = "domain" ]; then
  ACCESS_URL="https://${DOMAIN}"
else
  ACCESS_URL="http://${SERVER_IP}:8080"
fi

echo ""
echo "============================================"
echo "  部署完成！"
echo ""
echo "  访问地址: ${ACCESS_URL}"
echo "  管理员邮箱: admin@zsyq.local"
echo "  管理员密码: admin123（登录后请修改）"
echo "  数据库密码: ${PG_PASS}"
echo "  JWT密钥:    ${JWT}"
echo "  TOTP密钥:   ${TOTP}"
echo ""
echo "  修改前端后重新构建："
echo "  cd /opt/zsyq/deploy"
echo "  docker compose -f docker-compose.build.yml up -d --build"
echo ""
echo "  备份数据："
echo "  cd /opt/zsyq/deploy"
echo "  tar czf ~/zsyq-backup-\$(date +%Y%m%d).tar.gz data postgres_data redis_data .env"
echo ""
echo "  ⚠️  请立即："
echo "  1. 登录后修改管理员密码"
echo "  2. 截图保存上面这些密钥（只显示这一次）"
if [ "$ACCESS_MODE" = "domain" ]; then
  echo "  3. 确认 Cloudflare DNS 指向本机，SSL/TLS 使用 Full 或 Full (strict)"
  echo "  4. 确认服务器/服务商防火墙已放行 80 和 443"
fi
echo "============================================"