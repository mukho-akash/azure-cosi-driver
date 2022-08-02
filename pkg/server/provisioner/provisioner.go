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

package provisionerserver

import (
	"context"
	"fmt"
	"project/azure-cosi-driver/pkg/azureutils"
	"reflect"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	spec "sigs.k8s.io/container-object-storage-interface-spec"
)

type bucketDetails struct {
	bucketId   string
	parameters map[string]string
}

type provisioner struct {
	spec.UnimplementedProvisionerServer

	bucketsLock       sync.RWMutex
	nameToBucketMap   map[string]*bucketDetails
	bucketIdToNameMap map[string]string
	cloud             *azure.Cloud
}

var _ spec.ProvisionerServer = &provisioner{}

func NewProvisionerServer(
	kubeconfig,
	cloudConfigSecretName,
	cloudConfigSecretNamespace string) (spec.ProvisionerServer, error) {
	kubeClient, err := azureutils.GetKubeClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	klog.Infof("Kubeclient : %+v", kubeClient)

	azCloud, err := azureutils.GetAzureCloudProvider(kubeClient, cloudConfigSecretName, cloudConfigSecretNamespace)
	if err != nil {
		return nil, err
	}

	return &provisioner{
		nameToBucketMap:   make(map[string]*bucketDetails),
		bucketsLock:       sync.RWMutex{},
		bucketIdToNameMap: make(map[string]string),
		cloud:             azCloud,
	}, nil
}

func (pr *provisioner) DriverCreateBucket(
	ctx context.Context,
	req *spec.DriverCreateBucketRequest) (*spec.DriverCreateBucketResponse, error) {
	protocol := req.GetProtocol()
	if protocol == nil {
		return nil, status.Error(codes.InvalidArgument, "Protocol is nil")
	}

	azureBlob := protocol.GetAzureBlob()
	if azureBlob == nil {
		return nil, status.Error(codes.InvalidArgument, "Azure blob protocol is missing")
	}

	bucketName := req.GetName()
	parameters := req.GetParameters()
	if parameters == nil {
		parameters = make(map[string]string)
	}

	// Check if a bucket with these set of values exist in the namesToBucketMap
	pr.bucketsLock.RLock()
	currBucket, exists := pr.nameToBucketMap[bucketName]
	pr.bucketsLock.RUnlock()

	if exists {
		bucketParams := currBucket.parameters
		if bucketParams == nil {
			bucketParams = make(map[string]string)
		}
		if reflect.DeepEqual(bucketParams, parameters) {
			return &spec.ProvisionerCreateBucketResponse{
				BucketId: currBucket.bucketId,
			}, nil
		}

		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Bucket %s exists with different parameters", bucketName))
	}

	storageAccountName := azureBlob.StorageAccount

	bucketId, err := azureutils.CreateBucket(ctx, storageAccountName, bucketName, parameters, pr.cloud)
	if err != nil {
		return nil, err
	}

	// Insert the bucket into the namesToBucketMap
	pr.bucketsLock.RLock()
	pr.nameToBucketMap[bucketName] = &bucketDetails{
		bucketId:   bucketId,
		parameters: parameters,
	}
	pr.bucketIdToNameMap[bucketId] = bucketName
	pr.bucketsLock.RUnlock()

	klog.Infof("DriverCreateBucket :: Bucket id :: %s", bucketId)

	return &spec.DriverCreateBucketResponse{
		BucketId: bucketId,
	}, nil
}

func (pr *provisioner) DriverDeleteBucket(
	ctx context.Context,
	req *spec.DriverDeleteBucketRequest) (*spec.DriverDeleteBucketResponse, error) {
	bucketId := req.GetBucketId()
	klog.Infof("DriverDeleteBucket :: Bucket id :: %s", bucketId)
	err := azureutils.DeleteBucket(ctx, bucketId, pr.cloud)

	if err != nil {
		return nil, err
	}

	if bucketName, ok := pr.bucketIdToNameMap[bucketId]; ok {
		// Remove from the namesToBucketMap
		pr.bucketsLock.RLock()
		delete(pr.nameToBucketMap, bucketName)
		delete(pr.bucketIdToNameMap, bucketId)
		pr.bucketsLock.RUnlock()
	}

	return &spec.DriverDeleteBucketResponse{}, nil
}

func (pr *provisioner) DriverGrantBucketAccess(
	ctx context.Context,
	req *spec.DriverGrantBucketAccessRequest) (*spec.DriverGrantBucketAccessResponse, error) {
	return &spec.DriverGrantBucketAccessResponse{}, nil
}

func (pr *provisioner) DriverRevokeBucketAccess(
	ctx context.Context,
	req *spec.DriverRevokeBucketAccessRequest) (*spec.DriverRevokeBucketAccessResponse, error) {
	return &spec.DriverRevokeBucketAccessResponse{}, nil
}

func (pr *provisioner) DriverGetInfo(
	ctx context.Context,
	req *spec.DriverGetInfoRequest) (*spec.DriverGetInfoResponse, error) {
	return &spec.DriverGetInfoRequest{}, nil
}
