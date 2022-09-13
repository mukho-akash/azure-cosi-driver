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
	"net/url"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

const (
	AccessKey = ""
)

var (
	storageAccountRE = regexp.MustCompile(`https://(.+).blob.core.windows.net/([^/]*)/?(.*)`)
)

func createContainerBucket(
	ctx context.Context,
	bucketName string,
	parameters *BucketClassParameters,
	cloud *azure.Cloud) (string, error) {
	accOptions := getAccountOptions(parameters)
	_, key, err := cloud.EnsureStorageAccount(ctx, accOptions, "")
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Could not ensure storage account %s exists: %v", accOptions.Name, err))
	}
	containerParams := make(map[string]string) //NOTE: Container parameters still need to be filled/implemented
	return createAzureContainer(ctx, parameters.storageAccountName, key, bucketName, containerParams)
}

func DeleteContainerBucket(
	ctx context.Context,
	bucketID string,
	cloud *azure.Cloud) error {
	// Get storage account name from bucketID
	storageAccountName := getStorageAccountNameFromContainerURL(bucketID)
	// Get access keys for the storage account
	accessKey, err := cloud.GetStorageAccesskey(ctx, cloud.SubscriptionID, storageAccountName, cloud.ResourceGroup)
	if err != nil {
		return err
	}

	containerName := getContainerNameFromContainerURL(bucketID)
	err = deleteAzureContainer(ctx, storageAccountName, accessKey, containerName)
	if err != nil {
		return fmt.Errorf("Error deleting container %s in storage account %s : %v", containerName, storageAccountName, err)
	}

	// Now, we check and delete the storage account if its empty
	return nil
}

func getStorageAccountNameFromContainerURL(containerURL string) string {
	storageAccountName, _, _ := parsecontainerurl(containerURL)
	return storageAccountName
}

func getContainerNameFromContainerURL(containerURL string) string {
	_, containerName, _ := parsecontainerurl(containerURL)
	return containerName
}

func deleteAzureContainer(
	ctx context.Context,
	storageAccount,
	accessKey,
	containerName string) error {
	containerClient, err := createContainerClient(storageAccount, accessKey, containerName)

	if err != nil {
		return err
	}

	response, err = containerClient.Delete(ctx, nil)
	return err
}

func createContainerClient(
	storageAccount string,
	accessKey string,
	containerName string) (*azblob.ContainerClient, error) {
	// Create credentials
	credential, err := azblob.NewSharedKeyCredential(storageAccount, accessKey)
	if err != nil {
		return azblob.ContainerURL{}, fmt.Errorf("Invalid credentials with error : %v", err)
	}

	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, containerName)

	containerClient, err := azblob.NewContainerClientWithSharedKey(containerURL, credential, nil)

	return containerClient
)

func createContainerURL(
	storageAccount string,
	accessKey string,
	containerName string) (azblob.ContainerURL, error) {

	// Create a default request pipeline using credential
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	urlString, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", storageAccount))
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	serviceURL := azblob.NewServiceURL(*urlString, pipeline)

	// Create containerURL that wraps the service url and pipeline to make requests
	containerURL := serviceURL.NewContainerURL(containerName)

	return containerURL, nil
}

func parsecontainerurl(containerURL string) (string, string, string) {
	matches := storageAccountRE.FindStringSubmatch(containerURL)
	storageAccount := matches[1]
	containerName := matches[2]
	blobName := matches[3]
	return storageAccount, containerName, blobName
}

func createAzureContainer(
	ctx context.Context,
	storageAccount string,
	accessKey string,
	containerName string,
	parameters map[string]string) (string, error) {
	if len(storageAccount) == 0 || len(accessKey) == 0 {
		return "", fmt.Errorf("Invalid storage account or access key")
	}

	containerClient, err := createContainerClient(storageAccount, accessKey, containerName)
	if err != nil {
		return "", err
	}

	// Lets create a container with the containerClient
	response, err = containerClient.Create(ctx, &azblob.ContainerCreateOptions{
		Metadata: parameters,
		Access: nil,
	})
	if err != nil {
		return "", fmt.Errorf("Error creating container from containterURL : %s, Error : %v", containerURL.String(), err)
	}

	return containerURL.String(), nil
}

func createContainerSASURL(ctx context.Context, bucketID string, parameters *BucketAccessClassParameters) (string, string, error) {
	account := getStorageAccountNameFromContainerURL(bucketID)
	cred, err := azblob.NewSharedKeyCredential(account, parameters.key)
	if err != nil {
		return "", "", err
	}

	if err != nil {
		return "", "", err
	}

	permission := azblob.ContainerSASPermissions{}
	permission.List = parameters.enableList
	permission.Read = parameters.enableRead
	permission.Write = parameters.enableWrite
	permission.Delete = parameters.enableDelete
	permission.DeletePreviousVersion = parameters.enablePermanentDelete
	permission.Add = parameters.enableAdd
	permission.Tag = parameters.enableTags

	start := time.Now()
	expiry := start.Add(time.Millisecond * time.Duration(parameters.validationPeriod))

	sasQueryParams, err := azblob.BlobSASSignatureValues{
		Protocol:    azblob.SASProtocol(parameters.signedProtocol),
		StartTime:   start,
		ExpiryTime:  expiry,
		Permissions: permission.String(),
		IPRange:     azblob.IPRange(parameters.signedIP),
		Version:     parameters.signedversion,
	}.NewSASQueryParameters(cred)
	if err != nil {
		return "", "", err
	}

	queryParams := sasQueryParams.Encode()
	sasURL := fmt.Sprintf("%s/%s", bucketID, queryParams)
	accountID := fmt.Sprintf("https://%s.blob.core.windows.net/", getStorageAccountNameFromContainerURL(bucketID))
	return sasURL, accountID, nil
}
