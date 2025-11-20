package server

import (
	"context"

	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) initTrans() map[string]*com.TranslationElements {
	trans, err := s.commonClient.TranslationsGet()
	if err != nil {
		s.errors <- err
	}

	lang := s.config.Localization.GetDefaultClientLocale()
	if err := models.TranslationsInit(trans, lang); err != nil {
		path := "inventory.server.initTrans"
		err := &models.InternalError{Err: err, Msg: "failed to init translations", Path: path}
		s.errors <- err
	}

	return trans
}

func (s *Server) initDB() {
	pool, err := pgxpool.New(context.Background(), s.config.Sql.GetDataSource())
	if err != nil {
		path := "inventory.server.initDB"
		err := &models.InternalError{Err: err, Msg: "failed to init db pool", Path: path}
		s.errors <- err
	}
	s.dbConn = pool
}
