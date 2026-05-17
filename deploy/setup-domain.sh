#!/bin/bash
set -e

# ============================================================
# 智算引擎 域名 HTTPS 一键配置脚本
# ============================================================
# Usage:
#   bash setup-domain.sh your-domain.com
# ============================================================

DOMAIN="${1:-}"

if [ -z "$DOMAIN" ]; then
  echo "用法: bash setup-domain.sh 你的域名.com"
  echo "示例: bash setup-domain.sh api.example.com"
  exit 1
fi

if [ ! -f /opt/zsyq/deploy/.env ]; then
  echo "错误: 未找到 /opt/zsyq/deploy/.env，请先完成系统部署。"
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "错误: 未检测到 Docker，请先完成系统部署。"
  exit 1
fi

echo "[1/5] 安装 Caddy..."
apt update
apt install -y caddy

echo "[2/5] 将智算引擎绑定到本机 127.0.0.1:8080..."
cd /opt/zsyq/deploy
sed -i "s|^BIND_HOST=.*|BIND_HOST=127.0.0.1|" .env

echo "[3/5] 重启智算引擎..."
docker compose -f docker-compose.build.yml up -d

echo "[4/5] 写入 Caddy 反向代理配置..."
cat > /etc/caddy/Caddyfile << EOF
${DOMAIN} {
    reverse_proxy 127.0.0.1:8080
}
EOF

caddy fmt --overwrite /etc/caddy/Caddyfile >/dev/null

echo "[5/5] 启动 Caddy..."
systemctl enable caddy >/dev/null
systemctl restart caddy

sleep 2

echo ""
echo "============================================"
echo "  域名配置完成"
echo ""
echo "  访问地址: https://${DOMAIN}"
echo ""
echo "  如果打不开，请确认："
echo "  1. Cloudflare DNS A记录已指向本服务器IP"
echo "  2. 服务器防火墙/服务商安全组放行 80 和 443"
echo "  3. Cloudflare SSL/TLS 使用 Full 或 Full (strict)"
echo ""
echo "  查看 Caddy 日志："
echo "  journalctl -u caddy -n 100 --no-pager"
echo "============================================"
