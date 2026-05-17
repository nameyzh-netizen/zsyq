# 智算引擎

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**サブスクリプションクォータ配分のための AI API ゲートウェイプラットフォーム**

[English](README.md) | [中文](项目介绍.md) | 日本語

</div>

> **智算引擎 が公式に使用しているドメインは `zsyq.cloud` と `pincc.ai` のみです。智算引擎 の名称を使用している他のウェブサイトは、サードパーティによるデプロイやサービスであり、本プロジェクトとは一切関係がありません。ご利用の際はご自身で確認・判断をお願いします。**

---

## 概要

智算引擎 は、AI 製品のサブスクリプションから API クォータを配分・管理するために設計された AI API ゲートウェイプラットフォームです。ユーザーはプラットフォームが生成した API キーを通じて上流の AI サービスにアクセスでき、プラットフォームは認証、課金、負荷分散、リクエスト転送を処理します。

## 機能

- **マルチアカウント管理** - 複数の上流アカウントタイプ（OAuth、APIキー）をサポート
- **APIキー配布** - ユーザー向けの APIキーの生成と管理
- **精密な課金** - トークンレベルの使用量追跡とコスト計算
- **スマートスケジューリング** - スティッキーセッション付きのインテリジェントなアカウント選択
- **同時実行制御** - ユーザーごと・アカウントごとの同時実行数制限
- **レート制限** - 設定可能なリクエスト数およびトークンレート制限
- **内蔵決済システム** - EasyPay、Alipay、WeChat Pay、Stripe に対応。ユーザーのセルフサービスチャージが可能で、別途決済サービスのデプロイは不要（[設定ガイド](docs/支付系统配置指南_EN.md)）
- **管理ダッシュボード** - 監視・管理のための Web インターフェース
- **外部システム連携** - 外部システム（チケット管理など）を iframe 経由で管理ダッシュボードに埋め込み可能

## エコシステム

智算引擎 を拡張・統合するコミュニティプロジェクト:

| プロジェクト | 説明 | 機能 |
|---------|-------------|----------|
| ~~[智算引擎Pay](https://github.com/touwaeriol/zsyqpay)~~ | ~~セルフサービス決済システム~~ | **内蔵済み** — 決済機能は 智算引擎 に統合されました。別途デプロイは不要です。[決済設定ガイド](docs/支付系统配置指南_EN.md)をご参照ください |
| [zsyq-mobile](https://github.com/ckken/zsyq-mobile) | モバイル管理コンソール | ユーザー管理、アカウント管理、監視ダッシュボード、マルチバックエンド切り替えが可能なクロスプラットフォームアプリ（iOS/Android/Web）。Expo + React Native で構築 |

## 技術スタック

| コンポーネント | 技術 |
|-----------|------------|
| バックエンド | Go 1.26+, Gin, Ent |
| フロントエンド | Vue 3.4+, Vite 5+, TailwindCSS |
| データベース | PostgreSQL 15+ |
| キャッシュ/キュー | Redis 7+ |

---

## Nginx リバースプロキシに関する注意

智算引擎（または CRS）を Nginx でリバースプロキシし、Codex CLI と組み合わせて使用する場合、Nginx の `http` ブロックに以下の設定を追加してください:

```nginx
underscores_in_headers on;
```

Nginx はデフォルトでアンダースコアを含むヘッダー（例: `session_id`）を破棄するため、マルチアカウント構成でのスティッキーセッションルーティングに支障をきたします。

---

## デプロイ

### 方法1: Docker Compose ソースビルド（推奨）

本番運用にはこの方式を推奨します。スクリプトが Docker のインストール、ソースコードのクローン、シークレット生成、サーバー CPU/メモリに応じた自動チューニングを行い、`docker-compose.build.yml` でフロントエンドとバックエンドをソースからビルドします。フロントエンド変更後もいつでも再ビルドできます。

#### 前提条件

- Debian 12 / Ubuntu 22.04+ Linux サーバー（amd64 または arm64）
- root 権限
- ドメイン HTTPS を使用する場合: ドメインの A レコードを事前にサーバー IP に向け、80/443 ポートを開放してください

#### クイックスタート（ワンクリック自動デプロイ）

```bash
bash <(curl -sL https://raw.githubusercontent.com/nameyzh-netizen/zsyq/main/deploy/deploy.sh)
```

スクリプト実行中にアクセス方式を選択できます:

- `1`: IP + ポートでアクセス（`http://YOUR_SERVER_IP:8080`）
- `2`: ドメイン HTTPS でアクセス。ドメインを入力すると、Caddy のインストール、HTTPS 証明書の発行/更新、リバースプロキシ設定を自動で行います

GitHub Raw にアクセスできない場合は、以下の代替方式を使用してください:

```bash
git clone https://github.com/nameyzh-netizen/zsyq.git /opt/zsyq
bash /opt/zsyq/deploy/deploy.sh
```

**ワンクリックスクリプトが自動で行う内容:**

- システム更新、Swap、ネットワーク sysctl、ファイルディスクリプタ上限などの基本チューニング
- Docker と Docker Compose v2 のインストール
- ソースコードを `/opt/zsyq` にクローン/更新
- `POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY`、`ADMIN_PASSWORD` の生成
- CPU/メモリに応じて、マシンリソースの約 80% をアプリ、PostgreSQL、Redis、DB 接続プールへ自動配分
- ローカルディレクトリにデータを保存: `deploy/data`、`deploy/postgres_data`、`deploy/redis_data`
- `docker-compose.build.yml` でソースからビルドして起動
- ドメインモードでは Caddy HTTPS リバースプロキシを自動設定

完了時にアクセス URL、管理者アカウント、シークレットが表示されます。安全に保存してください。デフォルト管理者:

```text
Email: admin@zsyq.local
Password: スクリプトがランダム生成し、完了時に表示します（ログイン後すぐに変更してください）
```

#### よく使うコマンド

```bash
cd /opt/zsyq/deploy

# コンテナ状態を確認
docker compose -f docker-compose.build.yml ps

# アプリログを表示
docker compose -f docker-compose.build.yml logs -f zsyq

# サービスを再起動
docker compose -f docker-compose.build.yml restart

# フロントエンド/バックエンド変更後に再ビルド
docker compose -f docker-compose.build.yml up -d --build

# すべてのサービスを停止
docker compose -f docker-compose.build.yml down
```

#### アップグレード

```bash
cd /opt/zsyq
git pull
cd deploy
docker compose -f docker-compose.build.yml up -d --build
```

#### 後からドメイン HTTPS を有効化

初回に IP アクセスを選んだ場合でも、後からドメイン HTTPS を有効化できます:

```bash
cd /opt/zsyq
git pull
bash /opt/zsyq/deploy/setup-domain.sh your-domain.com
```

Caddy が HTTPS 証明書を自動発行し、自動更新します。

#### バックアップと移行

```bash
cd /opt/zsyq/deploy

docker compose -f docker-compose.build.yml down
tar czf ~/zsyq-backup-$(date +%Y%m%d).tar.gz data postgres_data redis_data .env
```

新しいサーバーでは、これらのディレクトリと `.env` を復元してから以下を実行してください:

```bash
docker compose -f docker-compose.build.yml up -d --build
```

#### デプロイバージョン

| バージョン | イメージソース | データストレージ | 推奨用途 |
|---------|-------------|-----------|----------|
| **docker-compose.build.yml** | ローカルソースビルド | ローカルディレクトリ | 推奨本番デプロイ、フロントエンドカスタマイズ |
| **docker-compose.local.yml** | ビルド済みイメージ | ローカルディレクトリ | コード変更なし、素早く起動 |
| **docker-compose.yml** | ビルド済みイメージ | Docker 名前付きボリューム | 簡単な試用 |

#### データ管理機能の有効化（datamanagementd）

管理ダッシュボードの「データ管理」機能を有効にするには、ホスト上の `datamanagementd` プロセスを追加でデプロイします。

主要ポイント:

- メインプロセスは `/tmp/zsyq-datamanagement.sock` を検出
- このソケットが到達可能な場合のみデータ管理機能が有効化
- Docker 環境ではホストのソケットをコンテナ内の同じパスにマウント

詳細なデプロイ手順は `deploy/数据管理服务部署说明.md` を参照してください。

---

### 方法2: スクリプトによるインストール

GitHub Releases からビルド済みバイナリをダウンロードするワンクリックインストールスクリプトです。

#### 前提条件

- Linux サーバー（amd64 または arm64）
- PostgreSQL 15+（インストール済みかつ稼働中）
- Redis 7+（インストール済みかつ稼働中）
- root 権限

#### インストール手順

```bash
curl -sSL https://raw.githubusercontent.com/nameyzh-netizen/zsyq/main/deploy/install.sh | sudo bash
```

スクリプトは以下を実行します:
1. システムアーキテクチャの検出
2. 最新リリースのダウンロード
3. バイナリを `/opt/zsyq` にインストール
4. systemd サービスの作成
5. システムユーザーと権限の設定

#### インストール後の作業

```bash
# 1. サービスを起動
sudo systemctl start zsyq

# 2. 起動時の自動起動を有効化
sudo systemctl enable zsyq

# 3. ブラウザでセットアップウィザードを開く
# http://YOUR_SERVER_IP:8080
```

セットアップウィザードでは以下の設定を行います:
- データベース設定
- Redis 設定
- 管理者アカウントの作成

#### アップグレード

**管理ダッシュボード**の左上にある**アップデートを確認**ボタンをクリックすることで、ダッシュボードから直接アップグレードできます。

Web インターフェースでは以下が可能です:
- 新しいバージョンの自動確認
- ワンクリックでのアップデートのダウンロードと適用
- 必要に応じたロールバック

#### よく使うコマンド

```bash
# ステータスを確認
sudo systemctl status zsyq

# ログを表示
sudo journalctl -u zsyq -f

# サービスを再起動
sudo systemctl restart zsyq

# アンインストール
curl -sSL https://raw.githubusercontent.com/nameyzh-netizen/zsyq/main/deploy/install.sh | sudo bash -s -- uninstall -y
```

---

### 方法3: ソースからビルド

開発やカスタマイズのためにソースコードからビルドして実行します。

#### 前提条件

- Go 1.26+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+

#### ビルド手順

```bash
# 1. リポジトリをクローン
git clone https://github.com/nameyzh-netizen/zsyq.git
cd zsyq

# 2. Corepack を有効化し pnpm をインストール（未インストールの場合）
corepack enable
corepack prepare pnpm@10.33.0 --activate

# 3. フロントエンドをビルド
cd frontend
pnpm install
pnpm run build
# 出力先: ../backend/internal/web/dist/

# 4. フロントエンドを組み込んだバックエンドをビルド
cd ../backend
go build -tags embed -o zsyq ./cmd/server

# 5. 設定ファイルを作成
cp ../deploy/config.example.yaml ./config.yaml

# 6. 設定を編集
nano config.yaml
```

> **注意:** `-tags embed` フラグはフロントエンドをバイナリに組み込みます。このフラグがない場合、バイナリはフロントエンド UI を提供しません。

**`config.yaml` の主要設定:**

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  dbname: "zsyq"

redis:
  host: "localhost"
  port: 6379
  password: ""

jwt:
  secret: "change-this-to-a-secure-random-string"
  expire_hour: 24

default:
  user_concurrency: 5
  user_balance: 0
  api_key_prefix: "sk-"
  rate_multiplier: 1.0
```

---

### Sora ステータス（削除済み）

> Sora 関連の機能は v1.0.0 で削除されました（データベースのテーブルとカラムはマイグレーションで削除済み）。`gateway.sora_*` および `sora:` 設定キーは機能せず、今後のバージョンでクリーンアップ予定です。

`config.yaml` では追加のセキュリティ関連オプションも利用できます:

- `cors.allowed_origins` - CORS 許可リスト
- `security.url_allowlist` - 上流/価格/CRS ホストの許可リスト
- `security.url_allowlist.enabled` - URL バリデーションの無効化（注意して使用）
- `security.url_allowlist.allow_insecure_http` - バリデーション無効時に HTTP URL を許可
- `security.url_allowlist.allow_private_hosts` - プライベート/ローカル IP アドレスを許可
- `security.response_headers.enabled` - 設定可能なレスポンスヘッダーフィルタリングを有効化（無効時はデフォルトの許可リストを使用）
- `security.csp` - Content-Security-Policy ヘッダーの制御
- `billing.circuit_breaker` - 課金エラー時にフェイルクローズ
- `server.trusted_proxies` - X-Forwarded-For パースの有効化
- `turnstile.required` - リリースモードでの Turnstile 必須化

**ゲートウェイ防御の多層化推奨（重要）**

- `gateway.upstream_response_read_max_bytes`: 非ストリーミング上流レスポンスの読み取りサイズを制限（デフォルト `8MB`）。異常レスポンスによるメモリ増大を防止。
- `gateway.proxy_probe_response_read_max_bytes`: プロキシプローブレスポンスの読み取りサイズを制限（デフォルト `1MB`）。
- `gateway.gemini_debug_response_headers`: デフォルト `false`。トラブルシューティング時のみ短時間有効化し、高頻度リクエストログのオーバーヘッドを回避。
- `/auth/register`、`/auth/login`、`/auth/login/2fa`、`/auth/send-verify-code` はサーバー側のフォールバックレート制限を提供（Redis 障害時はフェイルクローズ）。
- 推奨: WAF/CDN を第一防御層、サーバー側のレート制限とレスポンス読み取り上限を第二防御層として使用。両層を保持し、バイパス流量や誤設定のリスクを防止。

**⚠️ セキュリティ警告: HTTP URL 設定**

`security.url_allowlist.enabled=false` の場合、システムはデフォルトで最小限の URL バリデーションを行い、**HTTP URL を拒否**して HTTPS のみを許可します。HTTP URL を許可するには（開発環境や内部テスト用など）、以下を明示的に設定する必要があります:

```yaml
security:
  url_allowlist:
    enabled: false                # 許可リストチェックを無効化
    allow_insecure_http: true     # HTTP URL を許可（⚠️ セキュリティリスクあり）
```

**または環境変数で設定:**

```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
```

**HTTP を許可するリスク:**
- API キーとデータが**平文**で送信される（傍受の危険性）
- **中間者攻撃（MITM）**を受けやすい
- **本番環境には不適切**

**HTTP を使用すべき場面:**
- ✅ ローカルサーバーでの開発・テスト（http://localhost）
- ✅ 信頼できるエンドポイントを持つ内部ネットワーク
- ✅ HTTPS 取得前のアカウント接続テスト
- ❌ 本番環境（HTTPS のみを使用）

**この設定なしで表示されるエラー例:**
```
Invalid base URL: invalid url scheme: http
```

URL バリデーションまたはレスポンスヘッダーフィルタリングを無効にする場合は、ネットワーク層を強化してください:
- 上流ドメイン/IP のエグレス許可リストを適用
- プライベート/ループバック/リンクローカル範囲をブロック
- TLS のみのアウトバウンドトラフィックを強制
- プロキシで機密性の高い上流レスポンスヘッダーを除去

```bash
# 7. アプリケーションを実行
./zsyq
```

#### HTTP/2 (h2c) と HTTP/1.1 フォールバック

バックエンドのプレーンテキストポートはデフォルトで h2c をサポートし、WebSocket とレガシークライアント用に HTTP/1.1 フォールバックを保持します。ブラウザは通常 h2c をサポートしないため、性能向上は主にリバースプロキシや内部ネットワークリンクで得られます。

**リバースプロキシ例（Caddy）：**

```caddyfile
transport http {
	versions h2c h1
}
```

**検証：**

```bash
# h2c prior knowledge
curl --http2-prior-knowledge -I http://localhost:8080/health
# HTTP/1.1 フォールバック
curl --http1.1 -I http://localhost:8080/health
# WebSocket フォールバック検証（管理者トークンが必要）
websocat -H="Sec-WebSocket-Protocol: zsyq-admin, jwt.<ADMIN_TOKEN>" ws://localhost:8080/api/v1/admin/ops/ws/qps
```

#### 開発モード

```bash
# バックエンド（ホットリロード付き）
cd backend
go run ./cmd/server

# フロントエンド（ホットリロード付き）
cd frontend
pnpm run dev
```

#### コード生成

`backend/ent/schema` を編集した場合、Ent + Wire を再生成してください:

```bash
cd backend
go generate ./ent
go generate ./cmd/server
```

---

## シンプルモード

シンプルモードは、フル SaaS 機能を必要とせず、素早くアクセスしたい個人開発者や社内チーム向けに設計されています。

- 有効化: 環境変数 `RUN_MODE=simple` を設定
- 違い: SaaS 関連機能を非表示にし、課金プロセスをスキップ
- セキュリティに関する注意: 本番環境では `SIMPLE_MODE_CONFIRM=true` も設定する必要があります

---

## Antigravity サポート

智算引擎 は [Antigravity](https://antigravity.so/) アカウントをサポートしています。認証後、Claude および Gemini モデル用の専用エンドポイントが利用可能になります。

### 専用エンドポイント

| エンドポイント | モデル |
|----------|-------|
| `/antigravity/v1/messages` | Claude モデル |
| `/antigravity/v1beta/` | Gemini モデル |

### Claude Code の設定

```bash
export ANTHROPIC_BASE_URL="http://localhost:8080/antigravity"
export ANTHROPIC_AUTH_TOKEN="sk-xxx"
```

### ハイブリッドスケジューリングモード

Antigravity アカウントはオプションの**ハイブリッドスケジューリング**をサポートしています。有効にすると、汎用エンドポイント `/v1/messages` および `/v1beta/` も Antigravity アカウントにリクエストをルーティングします。

> **⚠️ 警告**: Anthropic Claude と Antigravity Claude は**同じ会話コンテキスト内で混在させることはできません**。グループを使用して適切に分離してください。

### 既知の問題

Claude Code では、Plan Mode を自動的に終了できません。（通常、ネイティブの Claude API を使用する場合、計画が完了すると Claude Code はユーザーに計画を承認または拒否するオプションをポップアップ表示します。）

**回避策**: `Shift + Tab` を押して手動で Plan Mode を終了し、計画を承認または拒否するためのレスポンスを入力してください。

---

## プロジェクト構成

```
zsyq/
├── backend/                  # Go バックエンドサービス
│   ├── cmd/server/           # アプリケーションエントリ
│   ├── internal/             # 内部モジュール
│   │   ├── config/           # 設定
│   │   ├── model/            # データモデル
│   │   ├── service/          # ビジネスロジック
│   │   ├── handler/          # HTTP ハンドラー
│   │   └── gateway/          # API ゲートウェイコア
│   └── resources/            # 静的リソース
│
├── frontend/                 # Vue 3 フロントエンド
│   └── src/
│       ├── api/              # API 呼び出し
│       ├── stores/           # 状態管理
│       ├── views/            # ページコンポーネント
│       └── components/       # 再利用可能なコンポーネント
│
└── deploy/                   # デプロイファイル
    ├── docker-compose.yml    # Docker Compose 設定
    ├── .env.example          # Docker Compose 用環境変数
    ├── config.example.yaml   # バイナリデプロイ用フル設定ファイル
    └── install.sh            # ワンクリックインストールスクリプト
```

## 免責事項

> **本プロジェクトをご利用の前に、以下をよくお読みください:**
>
> :rotating_light: **利用規約違反のリスク**: 本プロジェクトの使用は Anthropic の利用規約に違反する可能性があります。使用前に Anthropic のユーザー契約をよくお読みください。本プロジェクトの使用に起因するすべてのリスクは、ユーザー自身が負うものとします。
>
> :book: **免責事項**: 本プロジェクトは技術的な学習および研究目的のみで提供されています。作者は、本プロジェクトの使用によるアカウント停止、サービス中断、その他の損失について一切の責任を負いません。

---

## スター履歴

<a href="https://star-history.com/#nameyzh-netizen/zsyq&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=nameyzh-netizen/zsyq&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=nameyzh-netizen/zsyq&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=nameyzh-netizen/zsyq&type=Date" />
 </picture>
</a>

---

## ライセンス

本プロジェクトは [GNU Lesser General Public License v3.0](LICENSE)（またはそれ以降のバージョン）の下でライセンスされています。

Copyright (c) 2026 nameyzh-netizen

---

<div align="center">

**このプロジェクトが役に立ったら、ぜひスターをお願いします！**

</div>
