package interceptors

import (
	"context"
	"net/http"
	"time"

	"github.com/tuotoo/qrcode"
)

func CheckQRCode(uri string) bool {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	qrmatrix, err := qrcode.Decode(resp.Body)
	if err != nil {
		return false
	}
	if len(qrmatrix.Content) > 0 {
		return true
	}
	return false
}
