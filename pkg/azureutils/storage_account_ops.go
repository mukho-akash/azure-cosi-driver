// Copyright 2021 The Kubernetes Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azureutils

import (
	"context"
	"fmt"
	"project/azure-cosi-driver/pkg/types"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

func DeleteStorageAccount(
	ctx context.Context,
	id *types.BucketID,
	cloud *azure.Cloud) error {
	SAClient := cloud.StorageAccountClient
	err := SAClient.Delete(ctx, id.SubID, id.ResourceGroup, getStorageAccountNameFromContainerURL(id.URL))
	if err != nil {
		return err.Error()
	}
	return nil
}

func createStorageAccountBucket(ctx context.Context,
	bucketName string,
	parameters *BucketClassParameters,
	cloud *azure.Cloud) (string, error) {
	accName, _, err := cloud.EnsureStorageAccount(ctx, getAccountOptions(parameters), "")
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Could not create storage account: %v", err))
	}

	accURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accName)

	id := types.BucketID{
		ResourceGroup: parameters.resourceGroup,
		URL:           accURL,
	}
	if parameters.subscriptionID != "" {
		id.SubID = parameters.subscriptionID
	} else {
		id.SubID = cloud.SubscriptionID
	}
	base64ID, err := id.Encode()
	if err != nil {
		return "", status.Error(codes.InvalidArgument, fmt.Sprintf("could not encode ID: %v", err))
	}

	return base64ID, nil
}

// creates SAS and returns service client with sas
func createAccountSASURL(ctx context.Context, bucketID string, parameters *BucketAccessClassParameters) (string, string, error) {
	account := getStorageAccountNameFromContainerURL(bucketID)
	cred, err := azblob.NewSharedKeyCredential(account, parameters.key)
	if err != nil {
		return "", "", err
	}

	resources := azblob.AccountSASResourceTypes{}
	if parameters.allowServiceSignedResourceType {
		resources.Service = true
	}
	if parameters.allowContainerSignedResourceType {
		resources.Container = true
	}
	if parameters.allowObjectSignedResourceType {
		resources.Object = true
	}

	permission := azblob.AccountSASPermissions{}
	permission.List = parameters.enableList
	permission.Read = parameters.enableRead
	permission.Write = parameters.enableWrite
	permission.Delete = parameters.enableDelete
	permission.DeletePreviousVersion = parameters.enablePermanentDelete
	permission.Add = parameters.enableAdd
	permission.Tag = parameters.enableTags
	permission.FilterByTags = parameters.enableFilter

	start := time.Now()
	expiry := start.Add(time.Millisecond * time.Duration(parameters.validationPeriod))
	sasQueryParams, err := azblob.AccountSASSignatureValues{
		Protocol:      parameters.signedProtocol,
		StartTime:     start,
		ExpiryTime:    expiry,
		Permissions:   permission.String(),
		ResourceTypes: resources.String(),
		Services:      azblob.AccountSASServices{Blob: true}.String(),
		IPRange:       parameters.signedIP,
		Version:       parameters.signedversion,
	}.Sign(cred)
	if err != nil {
		return "", "", err
	}
	queryParams := sasQueryParams.Encode()
	sasURL := fmt.Sprintf("%s/?%s", bucketID, queryParams)
	return sasURL, bucketID, nil
}
