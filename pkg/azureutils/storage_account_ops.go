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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

func DeleteStorageAccount(
	ctx context.Context,
	account string,
	cloud *azure.Cloud) error {
	SAClient := cloud.StorageAccountClient
	err := SAClient.Delete(ctx, cloud.SubscriptionID, cloud.ResourceGroup, account)
	return err.Error()
}

func createStorageAccountBucket(ctx context.Context,
	bucketName string,
	parameters *BucketClassParameters,
	cloud *azure.Cloud) (string, error) {
	accName, _, err := cloud.EnsureStorageAccount(ctx, getAccountOptions(parameters), "")
	if err != nil {
		return "", status.Error(codes.Internal, fmt.Sprintf("Could not create storage account: %v", err))
	}
	return accName, nil
}
