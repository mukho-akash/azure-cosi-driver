package azureutils

import (
	"context"
	"fmt"
	"project/azure-cosi-driver/pkg/constant"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateBucketSASURL(ctx context.Context, bucketID string, parameters map[string]string) (string, error) {
	bucketAccessClassParams, err := parseBucketAccessClassParameters(parameters)
	if err != nil {
		return "", err
	}

	switch bucketAccessClassParams.bucketUnitType {
	case constant.Container:
		return createContainerSASURL(ctx, bucketID, *bucketAccessClassParams)
	case constant.StorageAccount:
		return createAccountSASURL(ctx, bucketID, *bucketAccessClassParams)
	}
	return "", status.Error(codes.InvalidArgument, "invalid bucket type")
}

func createContainerSASURL(ctx context.Context, bucketID string, parameters BucketAccessClassParameters) (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}

	containerClient, err := azblob.NewContainerClient(bucketID, cred, nil)
	if err != nil {
		return "", err
	}

	permission := azblob.ContainerSASPermissions{}
	if parameters.enableList == true {
		permission.List = true
	}
	if parameters.enableRead == true {
		permission.Read = true
	}
	if parameters.enableWrite == true {
		permission.Write = true
	}
	if parameters.enablePermanentDelete == true {
		permission.DeletePreviousVersion = true
	}

	var start time.Time
	if parameters.signedStart != nil {
		start = *parameters.signedStart
	} else {
		start = time.Now()
	}

	expiry := start.AddDate(0, 0, parameters.signedExpiry)

	sasURL, err := containerClient.GetSASURL(permission, start, expiry)
	if err != nil {
		return "", err
	}
	return sasURL, nil
}

//creates SAS and returns service client with sas
func createAccountSASURL(ctx context.Context, bucketID string, parameters BucketAccessClassParameters) (string, error) {
	account := getStorageAccountNameFromContainerURL(bucketID)
	cred, err := azblob.NewSharedKeyCredential(account, parameters.key)
	if err != nil {
		return "", err
	}
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", account)
	serviceClient, err := azblob.NewServiceClientWithSharedKey(serviceURL, cred, nil)
	if err != nil {
		return "", err
	}

	resources := azblob.AccountSASResourceTypes{}
	if parameters.allowServiceSignedResourceType == true {
		resources.Service = true
	}
	if parameters.allowContainerSignedResourceType == true {
		resources.Container = true
	}
	if parameters.allowObjectSignedResourceType == true {
		resources.Object = true
	}

	permission := azblob.AccountSASPermissions{}
	if parameters.enableList == true {
		permission.List = true
	}
	if parameters.enableRead == true {
		permission.Read = true
	}
	if parameters.enableWrite == true {
		permission.Write = true
	}
	if parameters.enablePermanentDelete == true {
		permission.DeletePreviousVersion = true
	}

	var start time.Time
	if parameters.signedStart != nil {
		start = *parameters.signedStart
	} else {
		start = time.Now()
	}

	expiry := start.AddDate(0, 0, parameters.signedExpiry)
	sasURL, err := serviceClient.GetSASURL(resources, permission, start, expiry)
	if err != nil {
		return "", err
	}
	return sasURL, nil
}
