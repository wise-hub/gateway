package configuration

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Load and initialize the application configuration
func Init() (*Dependencies, error) {
	env := flag.String("env", "TEST", "Set the environment type (DEV, TEST, PROD)")
	cfgPath := flag.String("cfg", "./config.json", "Set the configuration file path")
	flag.Parse()

	cfg, err := loadCfg(*cfgPath)
	if err != nil {
		return nil, err
	}

	var curCfg *ConfigItem
	for _, config := range cfg.Config {
		if config.EnvType == *env {
			curCfg = &config
			break
		}
	}

	if curCfg == nil {
		return nil, fmt.Errorf("no configuration found for environment: %s", *env)
	}

	dbPool, err := connectPostgres(curCfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Dependencies{
		Cfg: curCfg,
		Db:  dbPool,
	}, nil
}

func LoadAllowedEndpoints(filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	allowedEndpoints = make(map[string]struct{})

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allowedEndpoints[scanner.Text()] = struct{}{}
	}

	return scanner.Err()
}

// Connect to PostgreSQL database using pgxpool
func connectPostgres(dbCfg Database) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		dbCfg.Username,
		dbCfg.Password,
		dbCfg.Server,
		dbCfg.Port,
		dbCfg.Service,
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL connection string: %w", err)
	}

	config.MaxConns = int32(dbCfg.MaxOpenConns)
	config.MaxConnLifetime = dbCfg.ConnMaxLifetime * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return pool, nil
}

// Load the configuration file
func loadCfg(filePath string) (*MainConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg MainConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
