package interceptors

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	pp "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func CheckSex(ctx context.Context, data []byte) (bool, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		session.Logger(ctx).Errorf("CheckSex NewImageAnnotatorClient ERROR: %+v", err)
		return true, err
	}
	image, err := vision.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		session.Logger(ctx).Errorf("CheckSex NewImageFromReader ERROR: %+v", err)
		return true, err
	}
	safe, err := client.DetectSafeSearch(ctx, image, nil)
	if err != nil {
		session.Logger(ctx).Errorf("CheckSex DetectSafeSearch ERROR: %+v", err)
		return true, err
	}
	session.Logger(ctx).Infof("CheckSex DetectSafeSearch Adult: %s", safe.Adult)
	if safe.Adult >= pp.Likelihood_LIKELY {
		return true, errors.New(fmt.Sprintf("Adult level %s", safe.Adult))
	}
	return false, nil
}
