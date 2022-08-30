package azureutils

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

//creates SAS and returns service client with sas
func createAccountSASURL(ctx context.Context, account string, parameters BucketAccessClassParameters, cloud *azure.Cloud) (string, error) {
	cred, err := azblob.NewSharedKeyCredential(account, "myAccountKey")
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
		permission.Delete = true
	}

	start := time.Now()
	expiry := start.AddDate(0, 0, parameters.signedExpiry)
	sasURL, err := serviceClient.GetSASURL(resources, permission, start, expiry)
	if err != nil {
		return "", err
	}
	return sasURL, nil
}
