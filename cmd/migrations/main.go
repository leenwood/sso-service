package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file" // для миграций через файл
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	MigrationsDir string
	BatchSize     int
	DatabaseURL   string
}

func main() {
	fmt.Println("Миграции:")
	fmt.Println("--up для применения миграций вверх.")
	fmt.Println("--down для отката миграций вниз.")
	fmt.Println("--batch для указания размера пачки миграций (по умолчанию 3).")
	fmt.Println("Укажите путь к папке миграций и ссылку на базу данных.")

	cfg := loadConfig()

	if err := changeDir(cfg.MigrationsDir); err != nil {
		log.Fatalf("Ошибка перехода в папку миграций: %v", err)
	}

	// Можно также учесть аргументы командной строки
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "--up":
			fmt.Println("\nЗапуск миграций вверх (применение изменений)...")
			// Логика для --up
			if err := migrateUp(cfg.BatchSize, cfg.DatabaseURL); err != nil {
				log.Fatalf("Ошибка при миграции вверх: %v", err)
			}
		case "--down":
			fmt.Println("\nЗапуск миграций вниз (откат изменений)...")
			// Логика для --down
			if err := migrateDown(cfg.BatchSize, cfg.DatabaseURL); err != nil {
				log.Fatalf("Ошибка при откате миграции: %v", err)
			}
		default:
			fmt.Printf("\nНеизвестный аргумент: %s\n", arg)
		}
		return
	}

	fmt.Println("Укажите флаг --up или --down")
	return
}

// loadConfig загружает конфиг из флагов и env
func loadConfig() Config {
	// Флаги командной строки
	dirFlag := flag.String(
		"migrations-dir",
		"",
		"Путь к папке с миграциями (можно задать через MIGRATIONS_DIR)",
	)
	dbURLFlag := flag.String(
		"db-url",
		"",
		"Ссылка на базу данных (например, postgres://user:password@localhost:5432/database)",
	)
	batchFlag := flag.Int(
		"batch",
		3,
		"Размер пачки миграций (по умолчанию 3)",
	)
	flag.Parse()

	// Переменные окружения
	dirEnv := os.Getenv("MIGRATIONS_DIR")
	dbURLEnv := os.Getenv("DATABASE_URL")

	// Определение конечного пути для папки миграций
	dir := *dirFlag
	if dir == "" {
		dir = dirEnv
	}
	if dir == "" {
		dir = "./migrations" // Значение по умолчанию
	}

	// Определение подключения к базе данных
	dbURL := *dbURLFlag
	if dbURL == "" {
		dbURL = dbURLEnv
	}
	if dbURL == "" {
		log.Fatal("Не задано подключение к базе данных. Укажите --db-url или переменную окружения DATABASE_URL.")
	}

	return Config{
		MigrationsDir: dir,
		BatchSize:     *batchFlag,
		DatabaseURL:   dbURL,
	}
}

// changeDir переходит в указанную директорию
func changeDir(dir string) error {
	// Абсолютный путь (если передан относительный)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("ошибка преобразования пути: %w", err)
	}

	// Проверка существования
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("папка не найдена: %s", absDir)
	}

	// Смена директории
	if err := os.Chdir(absDir); err != nil {
		return fmt.Errorf("ошибка перехода в папку: %w", err)
	}

	log.Printf("Успешно перешли в: %s", absDir)
	return nil
}

// migrateUp выполняет миграции вверх в пачках
func migrateUp(batchSize int, dbURL string) error {
	// Создание нового мигратора
	m, err := migrate.New(
		"file://"+dbURL, // путь к миграциям (текущая папка)
		dbURL,           // DSN базы данных
	)
	if err != nil {
		return err
	}
	defer m.Close()

	// Выполнение миграций пачками
	for i := 0; i < batchSize; i++ {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		fmt.Printf("Миграция %d выполнена.\n", i+1)
	}

	fmt.Println("Миграции успешно применены!")
	return nil
}

// migrateDown выполняет откат миграций вниз в пачках
func migrateDown(batchSize int, dbURL string) error {
	// Создание нового мигратора
	m, err := migrate.New(
		"file://"+dbURL, // путь к миграциям (текущая папка)
		dbURL,           // DSN базы данных
	)
	if err != nil {
		return err
	}
	defer m.Close()

	// Выполнение отката миграций пачками
	for i := 0; i < batchSize; i++ {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		fmt.Printf("Миграция %d откатана.\n", i+1)
	}

	fmt.Println("Миграции успешно откатаны!")
	return nil
}
