package main

import (
	"context"
	"expvar"
	"log"
	"runtime"
	"social/internal/auth"
	"social/internal/db"
	"social/internal/env"
	"social/internal/ratelimiter"
	"social/internal/store"
	"social/internal/store/cache"
	"social/internal/store/mongodb"
	"time"

	mailer "social/internal/mailer"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GO blog training
//	@description	API server for GO blog
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				enter token to access this api
func main() {

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiUrl:      env.GetString("EXTERNAL_URL", "http://localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DATABASE_URL", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mongo: mongoConfig{
			addr: env.GetString("MONGODB_URI", "mongodb://localhost:27017"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pass:    env.GetString("REDIS_PASSWORD", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			fromEmail: env.GetString("FROM_EMAIL", ""),
			expiry:    time.Hour * 24 * 3,
			sendGrid: sendgridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user:     env.GetString("BASIC_AUTH_USER", ""),
				password: env.GetString("BASIC_AUTH_PASSWORD", ""),
			},
			token: tokenConfig{
				secret: env.GetString("JWT_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "GoBlog",
				aud:    "GoBlog",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}
	//Logger
	logger := zap.Must(zap.NewProduction(zap.AddStacktrace(zap.FatalLevel + 1))).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	logger.Info("Database connection pull established")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.mongo.addr))
	if err != nil {
		logger.Fatal("MongoDB connection error", zap.Error(err))
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		logger.Fatal("MongoDB ping failed", zap.Error(err))
	}
	defer client.Disconnect(context.Background())
	logger.Info("MongoDB connection pull established")

	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pass, cfg.redisCfg.db)
		logger.Info("redis cache connection established")

		defer rdb.Close()
	}

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	mongo := mongodb.NewMongoStorage(client.Database("analytics"))
	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	mailer := mailer.NewSendgrid(
		cfg.mail.sendGrid.apiKey,
		cfg.mail.fromEmail,
	)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.aud, cfg.auth.token.iss)

	app := &application{
		config:        cfg,
		store:         store,
		mongo:         mongo,
		cacheStore:    cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	//metrics
	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
