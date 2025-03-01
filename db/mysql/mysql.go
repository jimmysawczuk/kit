package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type silentLogger struct{}

func (s silentLogger) Print(...any) {}

type Config struct {
	User     string `envconfig:"USER" required:"true" default:"root"`
	Password string `envconfig:"PASSWD"`
	Net      string `envconfig:"NET"  required:"true" default:"tcp"`
	Addr     string `envconfig:"ADDR"`
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" default:"3306"`
	DB       string `envconfig:"DB" required:"true"`

	TLS        string `envconfig:"TLS"`
	CACertPath string `envconfig:"CA_CERT_PATH"`

	MaxOpenConns int `envconfig:"MAX_OPEN_CONNS" default:"10"`
	MaxIdleConns int `envconfig:"MAX_IDLE_CONNS" default:"10"`
}

func (c Config) addr() string {
	if c.Host != "" {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}

	return c.Addr
}

func (c Config) MySQLConfig() (*mysql.Config, error) {
	if c.CACertPath != "" {
		if c.TLS == "" {
			return nil, fmt.Errorf("tls must not be empty")
		}

		pem, err := os.ReadFile(c.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("couldn't read ca cert file (path: %s)", c.CACertPath)
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("can't decode pem file (file: %s)", c.CACertPath)
		}

		if err := mysql.RegisterTLSConfig(c.TLS, &tls.Config{
			RootCAs: certPool,
		}); err != nil {
			return nil, fmt.Errorf("mysql: register tls config: %w", err)
		}
	}

	cfg := mysql.NewConfig()
	cfg.Addr = c.addr()
	cfg.DBName = c.DB
	cfg.User = c.User
	cfg.Passwd = c.Password

	if c.TLS != "" {
		cfg.TLSConfig = c.TLS
	}

	cfg.ParseTime = true

	return cfg, nil
}

// Open uses the provided Config to connect to a MySQL database. If the connection is opened successfully,
// Open will also try to ping the connection to validate it further.
func Open(c Config) (*sqlx.DB, error) {
	cfg, err := c.MySQLConfig()
	if err != nil {
		return nil, fmt.Errorf("mysql config: %w", err)
	}

	mysql.SetLogger(silentLogger{})

	db, err := sqlx.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("sqlx: open: %w", err)
	}

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlx: ping: %w", err)
	}

	return db, nil
}
