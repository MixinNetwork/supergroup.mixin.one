package interceptors

import (
	"bytes"
	"context"
	"strings"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/tuotoo/qrcode"
)

func CheckQRCode(ctx context.Context, data []byte) bool {
	qrmatrix, err := qrcode.Decode(bytes.NewReader(data))
	if err != nil {
		session.Logger(ctx).Errorf("CheckQRCode Decode ERROR: %+v", err)
		if strings.Contains(err.Error(), "level and mask") {
			return true
		}

		return false
	}
	session.Logger(ctx).Infof("CheckQRCode qrmatrix: %d", len(qrmatrix.Content))
	if len(qrmatrix.Content) > 0 {
		return true
	}
	return false
}
