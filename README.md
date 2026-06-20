# kuromanager

kurohelper 資料庫相關管理TUI

## 功能

- **migrate**：執行 `kurohelper-service` 的資料庫 schema migration

## 設定

複製 `.env.example` 為 `.env`，填入資料庫連線資訊：

```
DB_NAME=
DB_OWNER=
DB_PASSWORD=
DB_PORT=
```

## 執行

```bash
go run .
```

## 專案結構

```
kuromanager/
├── main.go
├── internal/
│   ├── migrate/   # 呼叫 kurohelper-service/db
│   └── tui/       # Bubble Tea 介面
│       ├── tui.go
│       ├── menu.go
│       └── migrate.go
```
