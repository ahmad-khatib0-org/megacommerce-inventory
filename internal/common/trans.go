package common

import (
	"context"
	"time"

	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	"github.com/ahmad-khatib0-org/megacommerce-shared-go/pkg/models"
)

func (cc *CommonClient) TranslationsGet() (map[string]*com.TranslationElements, *models.InternalError) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ie := func(err error, msg string) *models.InternalError {
		return &models.InternalError{Path: "user.common.TranslationsGet", Err: err, Msg: msg}
	}

	res, err := cc.client.TranslationsGet(ctx, &com.TranslationsGetRequest{})
	if err != nil {
		return nil, ie(err, "failed to get translations from the common service")
	}

	if res.Error != nil {
		return nil, ie(models.AppErrorFromProto(nil, res.Error), "failed to get translations from the common service")
	}

	return res.Data, nil
}
