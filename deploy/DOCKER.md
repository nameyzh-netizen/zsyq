# 智算引擎 Docker Image

智算引擎 is an intelligent compute and AI service gateway for distributing and managing subscription API quotas.

## Quick Start

```bash
docker run -d \
  --name zsyq \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/zsyq" \
  -e REDIS_URL="redis://host:6379" \
  ghcr.io/nameyzh-netizen/zsyq:latest
```

## Docker Compose

```yaml
version: '3.8'

services:
  zsyq:
    image: ${ZSYQ_IMAGE:-ghcr.io/nameyzh-netizen/zsyq:latest}
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@db:5432/zsyq?sslmode=disable
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=zsyq
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `REDIS_URL` | Redis connection string | Yes | - |
| `PORT` | Server port | No | `8080` |
| `GIN_MODE` | Gin framework mode (`debug`/`release`) | No | `release` |

## Supported Architectures

- `linux/amd64`
- `linux/arm64`

## Tags

- `latest` - Latest stable release
- `x.y.z` - Specific version
- `x.y` - Latest patch of minor version
- `x` - Latest minor of major version

## Links

- [GitHub Repository](https://github.com/nameyzh-netizen/zsyq)
- [Documentation](https://github.com/nameyzh-netizen/zsyq#readme)
