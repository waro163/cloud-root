package cloudroot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DbStore struct {
	SqlStore   *sqlx.DB
	CacheStore *redis.Client
	MongoStore *mongo.Database
}

var (
	ErrSQLDriverNotSupported       = errors.New("SQL driver not supported")
	ErrMissingConnectionStringData = errors.New("missing connection string data")
)

func CreateStores(cfg map[string]DatabaseConfig) (map[string]DbStore, error) {
	storeMap := make(map[string]DbStore)
	for name, dbConfig := range cfg {
		switch dbConfig.Driver {
		case "mysql":
			db, err := newSqlStore(dbConfig)
			if err != nil {
				return nil, err
			}
			storeMap[name] = DbStore{SqlStore: db}
		case "redis":
			db, err := newRedisStore(dbConfig)
			if err != nil {
				return nil, err
			}
			storeMap[name] = DbStore{CacheStore: db}
		case "mongo":
			db, err := newMongoStore(dbConfig)
			if err != nil {
				return nil, err
			}
			storeMap[name] = DbStore{MongoStore: db}
		default:
			return nil, fmt.Errorf("not found store driver: %s", dbConfig.Driver)
		}
	}
	return storeMap, nil
}

func newRedisStore(cfg DatabaseConfig) (*redis.Client, error) {
	var addr = cfg.Host
	if cfg.Port != "" {
		addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	}
	db, err := strconv.Atoi(cfg.DatabaseName)
	if err != nil {
		return nil, err
	}
	cli := redis.NewClient(&redis.Options{
		Addr:       addr,
		ClientName: cfg.User,
		Password:   cfg.Password,
		DB:         db,
	})
	if cfg.Verify {
		if err := cli.Ping(context.Background()).Err(); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

func newMongoStore(cfg DatabaseConfig) (*mongo.Database, error) {
	var host = cfg.Host
	if cfg.Port != "" {
		host = fmt.Sprintf("%s:%s", host, cfg.Port)
	}
	if !strings.HasPrefix(host, "mongodb://") {
		host = "mongodb://" + host
	}
	credential := options.Credential{
		// AuthSource: "admin", //如果不填写，默认为admin
		Username: cfg.User,
		Password: cfg.Password,
	}
	cliOpt := options.Client().ApplyURI(host).SetAuth(credential)
	cli, err := mongo.Connect(context.Background(), cliOpt)
	if err != nil {
		return nil, err
	}
	if cfg.Verify {
		if err = cli.Ping(context.Background(), readpref.Primary()); err != nil {
			return nil, err
		}
	}
	db := cli.Database(cfg.DatabaseName)
	return db, nil
}

func newSqlStore(cfg DatabaseConfig) (*sqlx.DB, error) {
	dbConnection, err := GetConnectionByDriver(cfg.Driver, cfg)
	if err != nil {
		return nil, err
	}
	cli, err := sqlx.Open(cfg.Driver, dbConnection)
	if err != nil {
		return nil, err
	}
	if cfg.Verify {
		if err := cli.PingContext(context.Background()); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

func GetConnectionByDriver(driver string, opts DatabaseConfig) (string, error) {
	if driver == "" || opts.Host == "" {
		return "", ErrMissingConnectionStringData
	}

	host := opts.Host
	if opts.Port != "" {
		// add port to host
		host = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
	}

	switch driver {
	case "mysql":
		if opts.DatabaseName == "" {
			return "", ErrMissingConnectionStringData
		}
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&allowCleartextPasswords=1&interpolateParams=%t", opts.User, opts.Password, host, opts.DatabaseName, true), nil
	case "sqlite":
		if host == "" {
			return "", ErrMissingConnectionStringData
		}
		return fmt.Sprintf("file:%s?cache=shared&mode=memory", host), nil
	case "mongodb":
		if opts.DatabaseName == "" {
			return "", ErrMissingConnectionStringData
		}
		//mongodb host format is "mongodb://localhost:27017/?authSource=admin"
		if !strings.HasPrefix(host, "mongodb://") {
			host = "mongodb://" + host
		}
		u, err := url.Parse(host)
		if err != nil {
			return "", err
		}
		u.Path = opts.DatabaseName
		values := u.Query()
		u.RawQuery = values.Encode()

		if opts.User != "" {
			u.User = url.UserPassword(opts.User, opts.Password)
		}
		return u.String(), nil
	default:
		return "", ErrSQLDriverNotSupported
	}
}
