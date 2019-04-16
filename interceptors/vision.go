package interceptors

import (
	"context"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	pp "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func CheckSex(ctx context.Context, uri string) bool {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		session.Logger(ctx).Errorf("CheckSex NewImageAnnotatorClient ERROR: %+v", err)
		return true
	}
	image := vision.NewImageFromURI(uri)
	safe, err := client.DetectSafeSearch(ctx, image, nil)
	if err != nil {
		session.Logger(ctx).Errorf("CheckSex DetectSafeSearch ERROR: %+v, URI: %s", err, uri)
		return true
	}
	if safe.Adult >= pp.Likelihood_LIKELY {
		return true
	}
	return false
}
