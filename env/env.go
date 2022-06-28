package env

import "os"

type convertible interface {
	~[]byte | ~string
}

var (
	HS256_SECRET           []byte
	ED25519_PRIV_KEY_PATH  string
	JWKS_PATH              string
	NSQD_TCP_ADDR          string
	NSQD_API_ADDR          string
	EXCHANGE_NSQD_TCP_ADDR string
	NSQLOOKUPD_ADDR        string
	DB_CONN                string
	REDIS_CONN             string
	APP_PORT               string
	SERVER_ID              string
)

func initEnv[T convertible](dst *T, key string) {
	v := os.Getenv(key)
	if v == "" {
		os.Exit(1)
	}
	*dst = T(os.Getenv(key))
}

func init() {
	initEnv(&HS256_SECRET, "HS256_SECRET")
	initEnv(&ED25519_PRIV_KEY_PATH, "ED25519_PRIV_KEY_PATH")
	initEnv(&JWKS_PATH, "JWKS_PATH")
	initEnv(&NSQD_TCP_ADDR, "NSQD_TCP_ADDR")
	initEnv(&NSQD_API_ADDR, "NSQD_API_ADDR")
	initEnv(&EXCHANGE_NSQD_TCP_ADDR, "EXCHANGE_NSQD_TCP_ADDR")
	initEnv(&NSQLOOKUPD_ADDR, "NSQLOOKUPD_ADDR")
	initEnv(&DB_CONN, "DB_CONN")
	initEnv(&REDIS_CONN, "REDIS_CONN")
	initEnv(&APP_PORT, "APP_PORT")
	initEnv(&SERVER_ID, "SERVER_ID")
}
