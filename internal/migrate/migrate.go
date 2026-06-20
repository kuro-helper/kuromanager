package migrate

import (
	"os"

	kurohelperdb "kurohelperservice/db"

	"github.com/joho/godotenv"
)

func Run() error {
	_ = godotenv.Load()

	config := kurohelperdb.Config{
		DBOwner:    os.Getenv("DB_OWNER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
	}

	if err := kurohelperdb.InitDsn(config); err != nil {
		return err
	}
	return kurohelperdb.Migration(kurohelperdb.Dbs)
}
