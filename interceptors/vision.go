package interceptors

import (
	"context"

	vision "cloud.google.com/go/vision/apiv1"
	pp "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func CheckSex(ctx context.Context, uri string) bool {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return true
	}
	image := vision.NewImageFromURI(uri)
	safe, err := client.DetectSafeSearch(ctx, image, nil)
	if err != nil {
		return true
	}
	if safe.Adult >= pp.Likelihood_LIKELY {
		return true
	}
	return false
}
