package configuration

import (
	"database/sql"
	"log"
	"time"
)

type Dependencies struct {
	Cfg          *ConfigItem
	Db           *sql.DB
	AccessLogger *log.Logger
	ErrorLogger  *log.Logger
}

type MainConfig struct {
	AppGroupName         string       `json:"AppGroupName" validate:"required"`
	Version              string       `json:"Version" validate:"required"`
	AllowedEndpointsFile string       `json:"AllowedEndpointsFile" validate:"required"`
	Config               []ConfigItem `json:"Config" validate:"required,dive"`
}

type ConfigItem struct {
	EnvType          string        `json:"EnvType" validate:"required,oneof=DEV TEST PROD"`
	Port             string        `json:"Port" validate:"required"`
	LoadEndpointsPwd string        `json:"LoadEndpointsPwd" validate:"required"`
	InternalWsPwd    string        `json:"InternalWsPwd" validate:"required"`
	PublicFQDN       string        `json:"PublicFQDN" validate:"required,url"`
	ProtectedFQDN    string        `json:"ProtectedFQDN" validate:"required,url"`
	ReadTimeout      time.Duration `json:"ReadTimeout" validate:"required"`
	WriteTimeout     time.Duration `json:"WriteTimeout" validate:"required"`
	IdleTimeout      time.Duration `json:"IdleTimeout" validate:"required"`
	TokenDbCheck     string        `json:"TokenDbCheck" validate:"required,oneof=Y N"`
	Database         Database      `json:"Database" validate:"required,dive"`
}

type Database struct {
	Server          string        `json:"Server" validate:"required"`
	Port            string        `json:"Port" validate:"required"`
	Service         string        `json:"Service" validate:"required"`
	Username        string        `json:"Username" validate:"required"`
	Password        string        `json:"Password" validate:"required"`
	MaxOpenConns    int           `json:"MaxOpenConns" validate:"required"`
	MaxIdleConns    int           `json:"MaxIdleConns" validate:"required"`
	ConnMaxLifetime time.Duration `json:"ConnMaxLifetime" validate:"required"`
}
