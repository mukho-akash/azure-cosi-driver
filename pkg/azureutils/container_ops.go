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

	"github.com/Azure/azure-storage-blob-go/azblob"
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
	bucketId string,
	cloud *azure.Cloud) error {
	// Get storage account name from bucketId
	storageAccountName := getStorageAccountNameFromContainerUrl(bucketId)
	// Get access keys for the storage account
	accessKey, err := cloud.GetStorageAccesskey(ctx, cloud.SubscriptionID, storageAccountName, cloud.ResourceGroup)
	if err != nil {
		return err
	}

	containerName := getContainerNameFromContainerUrl(bucketId)
	err = deleteAzureContainer(ctx, storageAccountName, accessKey, containerName)
	if err != nil {
		return fmt.Errorf("Error deleting container %s in storage account %s : %v", containerName, storageAccountName, err)
	}

	// Now, we check and delete the storage account if its empty
	return nil
}

func getStorageAccountNameFromContainerUrl(containerUrl string) string {
	storageAccountName, _, _ := parsecontainerurl(containerUrl)
	return storageAccountName
}

func getContainerNameFromContainerUrl(containerUrl string) string {
	_, containerName, _ := parsecontainerurl(containerUrl)
	return containerName
}

func deleteAzureContainer(
	ctx context.Context,
	storageAccount,
	accessKey,
	containerName string) error {
	containerUrl, err := createContainerUrl(storageAccount, accessKey, containerName)

	if err != nil {
		return err
	}

	_, err = containerUrl.Delete(ctx, azblob.ContainerAccessConditions{})
	return err
}

func createContainerUrl(
	storageAccount string,
	accessKey string,
	containerName string) (azblob.ContainerURL, error) {
	// Create credentials
	credential, err := azblob.NewSharedKeyCredential(storageAccount, accessKey)
	if err != nil {
		return azblob.ContainerURL{}, fmt.Errorf("Invalid credentials with error : %v", err)
	}

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

func parsecontainerurl(containerUrl string) (string, string, string) {
	matches := storageAccountRE.FindStringSubmatch(containerUrl)
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

	containerURL, err := createContainerUrl(storageAccount, accessKey, containerName)
	if err != nil {
		return "", err
	}

	// Lets create a container with the containerURL
	_, err = containerURL.Create(ctx, parameters, azblob.PublicAccessNone)
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok {
			if serr.ServiceCode() == azblob.ServiceCodeBlobAlreadyExists {
				return containerURL.String(), nil
			}
		}
		return "", fmt.Errorf("Error creating container from containterURL : %s, Error : %v", containerURL.String(), err)
	}

	return containerURL.String(), nil
}
