package configuration

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

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

	db, err := connectDb(&curCfg.Database)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		Cfg: curCfg,
		Db:  db,
	}, nil
}

func loadCfg(path string) (*MainConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &MainConfig{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func connectDb(cfg *Database) (*sql.DB, error) {
	con := fmt.Sprintf("oracle://%s:%s@%s:%s/%s?charset=utf8",
		cfg.Username,
		cfg.Password,
		cfg.Server,
		cfg.Port,
		cfg.Service)

	db, err := sql.Open("oracle", con)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Minute * time.Duration(cfg.ConnMaxLifetime))

	return db, nil
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

func IsEndpointAllowed(endpoint string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, allowed := allowedEndpoints[endpoint]
	return allowed
}
