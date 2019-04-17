package interceptors

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/tuotoo/qrcode"
)

func CheckQRCode(ctx context.Context, uri string) bool {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		session.Logger(ctx).Errorf("CheckQRCode ERROR: %+v", err)
		return false
	}
	defer resp.Body.Close()

	qrmatrix, err := qrcode.Decode(resp.Body)
	if err != nil {
		session.Logger(ctx).Errorf("CheckQRCode Decode ERROR: %+v", err)
		if strings.Contains(err.Error(), "level and mask") {
			return true
		}

		return false
	}
	session.Logger(ctx).Infof("CheckQRCode qrmatrix: %d URI: %s", len(qrmatrix.Content), uri)
	if len(qrmatrix.Content) > 0 {
		return true
	}
	return false
}
