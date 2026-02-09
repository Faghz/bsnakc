package bootstrap

import (
	"context"

	identitiespostgres "github.com/elzestia/go-boilerplate/internal/identities/adapters/postgres"
	identitydomain "github.com/elzestia/go-boilerplate/internal/identities/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/crypto"
	userspostgres "github.com/elzestia/go-boilerplate/internal/users/adapters/postgres"
	usersdomain "github.com/elzestia/go-boilerplate/internal/users/domain"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	UserRepo     usersdomain.Repository
	IdentityRepo identitydomain.Repository
}

func initRepositories(
	db *sqlx.DB,
	logger ports.Logger,
	userCryptoService *crypto.CryptoService,
	identityCryptoService *crypto.CryptoService,
) *Repositories {
	logger.Info(context.Background(), "Initializing repositories")
	return &Repositories{
		UserRepo:     userspostgres.NewUserRepository(db, logger, userCryptoService),
		IdentityRepo: identitiespostgres.NewIdentityRepository(db, logger, identityCryptoService),
	}
}
