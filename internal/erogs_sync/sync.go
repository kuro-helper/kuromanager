package erogssync

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	kurohelperdb "kurohelperservice/db"
	"kurohelperservice/provider/erogs"
)

// 是一次同步寫入的資料筆數。
type Summary struct {
	BrandCount int
	GameCount  int
}

// 目前正在同步的品牌資料。
type Progress struct {
	Current   int
	Total     int
	BrandID   int
	BrandName string
}

// 從 Erogs 取得指定 Brand ID 範圍的資料並寫入資料庫。
func Run(minID, maxID int) (Summary, error) {
	return RunWithProgress(minID, maxID, nil)
}

// RunWithProgress 從 Erogs 取得指定 Brand ID 範圍的資料並寫入資料庫，
// 並在開始處理每個品牌前呼叫 onProgress。
func RunWithProgress(minID, maxID int, onProgress func(Progress)) (Summary, error) {
	if minID <= 0 || maxID <= 0 || minID > maxID {
		return Summary{}, fmt.Errorf("brand ID 範圍無效：min 與 max 必須大於 0，且 min 不可大於 max")
	}

	_ = godotenv.Load()

	if err := initDB(); err != nil {
		return Summary{}, err
	}

	erogs.InitRateLimit(time.Duration(envInt("EROGS_RATE_LIMIT_RESET_TIME", 10)))

	jsonText, err := erogs.ExecuteSQL(buildDumpSQL(minID, maxID))
	if err != nil {
		return Summary{}, err
	}

	brands, err := parseDumpBrands(jsonText)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{
		BrandCount: len(brands),
		GameCount:  countGames(brands),
	}
	if err := saveBrands(brands, onProgress); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func initDB() error {
	config := kurohelperdb.Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBOwner:    os.Getenv("DB_OWNER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
	}
	return kurohelperdb.InitDsn(config)
}

func parseDumpBrands(jsonText string) ([]DumpBrand, error) {
	jsonText = strings.TrimSpace(jsonText)
	if jsonText == "" || jsonText == "null" {
		return nil, nil
	}

	var brands []DumpBrand
	if err := json.Unmarshal([]byte(jsonText), &brands); err != nil {
		return nil, fmt.Errorf("解析 Erogs 回傳 JSON：%w", err)
	}
	return brands, nil
}

func countGames(brands []DumpBrand) int {
	total := 0
	for _, brand := range brands {
		total += len(brand.Gamelist)
	}
	return total
}

func saveBrands(brands []DumpBrand, onProgress func(Progress)) error {
	for i, brand := range brands {
		if onProgress != nil {
			onProgress(Progress{
				Current:   i + 1,
				Total:     len(brands),
				BrandID:   brand.ID,
				BrandName: brand.Name,
			})
		}

		if err := kurohelperdb.Dbs.Transaction(func(tx *gorm.DB) error {
			gameCount := len(brand.Gamelist)
			existingBrand, err := kurohelperdb.EnsureBrandErogs(
				tx, brand.ID, brand.Name, brand.Disband, gameCount,
			)
			if err != nil {
				return err
			}

			incomingBrand := &kurohelperdb.BrandErogs{
				Name:      brand.Name,
				Disband:   brand.Disband,
				GameCount: gameCount,
			}
			if brandChanged(existingBrand, incomingBrand) {
				if err := kurohelperdb.UpdateBrandErogs(tx, brand.ID, incomingBrand); err != nil {
					return err
				}
			}

			for _, game := range brand.Gamelist {
				existingGame, err := kurohelperdb.EnsureGameErogs(
					tx, game.ID, game.Name, game.Image, game.BrandErogsID, game.Category,
				)
				if err != nil {
					return err
				}

				incomingGame := &kurohelperdb.GameErogs{
					Name:         game.Name,
					BrandErogsID: game.BrandErogsID,
					Image:        game.Image,
					Category:     game.Category,
				}
				if gameChanged(existingGame, incomingGame) {
					if err := kurohelperdb.UpdateGameErogs(tx, game.ID, incomingGame); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("儲存 brand %d：%w", brand.ID, err)
		}
	}
	return nil
}

func envInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return n
}

func brandChanged(existing, incoming *kurohelperdb.BrandErogs) bool {
	return existing.Name != incoming.Name ||
		existing.Disband != incoming.Disband ||
		existing.GameCount != incoming.GameCount
}

func gameChanged(existing, incoming *kurohelperdb.GameErogs) bool {
	return existing.Name != incoming.Name ||
		existing.BrandErogsID != incoming.BrandErogsID ||
		existing.Image != incoming.Image ||
		existing.Category != incoming.Category
}
