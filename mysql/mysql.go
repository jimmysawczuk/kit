package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmysawczuk/try"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type silentLogger struct{}

func (s silentLogger) Print(...interface{}) {}

type Config struct {
	User       string `envconfig:"USER" required:"true" default:"root"`
	Password   string `envconfig:"PASSWD"`
	Host       string `envconfig:"HOST" required:"true" default:"localhost"`
	Port       int    `envconfig:"PORT" required:"true" default:"3306"`
	DB         string `envconfig:"DB" required:"true"`
	TLS        string `envconfig:"TLS"`
	IAMAuth    bool   `envconfig:"IAM_AUTH"`
	CACertPath string `envconfig:"CA_CERT_PATH"`
}

type DBMux struct {
	db  *sqlx.DB
	mux *sync.RWMutex
}

func NewDBMux(db *sqlx.DB) *DBMux {
	return &DBMux{
		db:  db,
		mux: &sync.RWMutex{},
	}
}

func (m *DBMux) DB() *sqlx.DB {
	m.mux.RLock()
	d := m.db
	m.mux.RUnlock()

	return d
}

// Refresh periodically refreshes the passed DB connection, if the connection is using IAM authentication. Credentials
// generated this way last for 15 minutes, per AWS's documentation:
//
// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html
//
// This function should be run as a goroutine, i.e.:
//
// db, _ := mysql.Open(cfg)
// go mysql.Refresh(log, cfg, 10 * time.Minute, db)
func (m *DBMux) Refresh(log logrus.FieldLogger, c Config, dur time.Duration) {
	// If we're not actually in IAM mode, we don't need to do this.
	if !c.IAMAuth {
		return
	}

	for {
		select {
		case <-time.Tick(dur):
			oldDB := new(sqlx.DB)
			newDB, err := Open(c)
			if err != nil {
				panic(errors.Wrap(err, "mysql: open"))
			}

			log.Info("replacing db connection")

			m.mux.Lock()
			*oldDB, *m.db = *m.db, *newDB
			m.mux.Unlock()

			if err := oldDB.Close(); err != nil {
				log.WithError(err).Warn("couldn't close old DB connection")
			}
		}
	}
}

// Open uses the provided Config to connect to a MySQL database. If the connection is opened successfully,
// Open will also try to ping the connection to validate it further.
func Open(c Config) (*sqlx.DB, error) {
	if c.IAMAuth {
		sess, err := session.NewSession()
		if err != nil {
			return nil, errors.Wrap(err, "aws: session: new session")
		}

		token, err := rdsutils.BuildAuthToken(fmt.Sprintf("%s:%d", c.Host, c.Port), *sess.Config.Region, c.User, sess.Config.Credentials)
		if err != nil {
			return nil, errors.Wrap(err, "aws: rdsutils: build auth token")
		}

		c.Password = token
	}

	if c.CACertPath != "" {
		pem, err := ioutil.ReadFile(c.CACertPath)
		if err != nil {
			return nil, errors.Errorf("couldn't read ca cert file (path: %s)", c.CACertPath)
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(pem); !ok {
			return nil, errors.Errorf("can't decode pem file (file: %s)", c.CACertPath)
		}

		if err := mysql.RegisterTLSConfig(c.TLS, &tls.Config{
			RootCAs: certPool,
		}); err != nil {
			return nil, errors.Wrap(err, "mysql: register tls config")
		}
	}

	cfg := mysql.Config{
		User:      c.User,
		Passwd:    c.Password,
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%d", c.Host, c.Port),
		DBName:    c.DB,
		TLSConfig: c.TLS,

		ParseTime:            true,
		AllowNativePasswords: true,
		MaxAllowedPacket:     4194304,
	}

	if c.IAMAuth {
		cfg.AllowCleartextPasswords = true
	}

	mysql.SetLogger(silentLogger{})

	db, err := sqlx.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "ping")
	}

	return db, nil
}

// Try wraps Open in a retry loop, retrying failed connections every interval until dur is elapsed.
func Try(c Config, dur, interval time.Duration) (*sqlx.DB, error) {
	var db *sqlx.DB

	err := try.Try(func() error {
		d, err := Open(c)
		if err != nil {
			return errors.Wrap(err, "mysql: open")
		}

		db = d
		return nil
	}, dur, interval)
	if err != nil {
		return nil, errors.Wrap(err, "try")
	}

	return db, nil
}
