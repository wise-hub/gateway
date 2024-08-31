package configuration

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

func Init() (*Dependencies, error) {
	// Get the environment type from command-line flag
	env := flag.String("env", "TEST", "Set the environment type (e.g., TEST, PROD)")
	flag.Parse()

	cfg, err := loadCfg("./cfg/config.json")
	if err != nil {
		return nil, err
	}

	// Select the configuration based on the input environment
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

	// Set up logging
	accessLogFile, err := os.OpenFile("log/access_log_"+time.Now().Format("2006_01_02")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open access log file: %v", err)
	}

	errorLogFile, err := os.OpenFile("log/error_log_"+time.Now().Format("2006_01_02")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open error log file: %v", err)
	}

	accessLogger := log.New(accessLogFile, "", log.Ldate|log.Ltime)
	errorLogger := log.New(errorLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Dependencies{
		Cfg:          curCfg,
		Db:           db,
		AccessLogger: accessLogger,
		ErrorLogger:  errorLogger,
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
