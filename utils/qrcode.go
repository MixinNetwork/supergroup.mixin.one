package utils

import (
	"bytes"
	"context"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/tuotoo/qrcode"
)

func CheckQRCode(ctx context.Context, data []byte) (bool, error) {
	qrmatrix, err := qrcode.Decode(bytes.NewReader(data))
	if err != nil {
		session.Logger(ctx).Errorf("CheckQRCode Decode ERROR: %+v", err)
		return false, err
	}
	session.Logger(ctx).Infof("CheckQRCode qrmatrix: %d", len(qrmatrix.Content))
	if len(qrmatrix.Content) > 0 {
		return true, nil
	}
	return false, nil
}
