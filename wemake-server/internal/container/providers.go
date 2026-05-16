package container

import (
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/config"
)

// Set contains all Wire providers for dependency injection
// These are the basic providers - add more as needed
var Set = wire.NewSet(
	provideConfig,
	provideDatabase,
)

// provideConfig loads and provides application configuration
func provideConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

// provideDatabase initializes and provides database connection
func provideDatabase(cfg *config.Config) *sqlx.DB {
	db, err := config.InitDatabase(cfg)
	if err != nil {
		panic(err)
	}
	return db
}

// ============= HOW TO ADD MORE PROVIDERS =============
//
// 1. Define a provider function:
//    func provideMyService(db *sqlx.DB) *service.MyService {
//        return service.NewMyService(db)
//    }
//
// 2. Add to the Set:
//    var Set = wire.NewSet(
//        provideConfig,
//        provideDatabase,
//        provideMyService,  // NEW
//    )
//
// 3. Regenerate wire_gen.go:
//    go generate ./internal/container
//
// Wire will automatically figure out dependencies!
