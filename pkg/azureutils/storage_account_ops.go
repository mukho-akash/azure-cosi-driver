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
	"strings"

	"project/azure-cosi-driver/pkg/constant"

	"github.com/Azure/go-autorest/autorest/to"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

func DeleteStorageAccount(
	ctx context.Context,
	account string,
	cloud *azure.Cloud) error {
	SAClient := cloud.StorageAccountClient
	SAClient.Delete(ctx, cloud.SubscriptionID, cloud.ResourceGroup, account)
	return nil
}

func parseParametersForStorageAccount(
	parameters map[string]string,
	cloud *azure.Cloud) (*azure.AccountOptions, error) {
	if parameters == nil {
		parameters = make(map[string]string)
	}

	var (
		enableHTTPSTrafficOnly bool
		createPrivateEndpoint  bool
		isHnsEnabled           bool
		enableNfsV3            bool
		enableLargeFileShares  bool
		accountType            string
		kind                   string
		nvResourceIdsStr       string
		customTags             string
		location               string
	)

	for key, val := range parameters {
		switch strings.ToLower(key) {
		case constant.StorageAccountTypeField:
			accountType = val
		case constant.LocationField:
			location = val
		case constant.KindField:
			kind = val
		case constant.TagsField:
			customTags = val
		case constant.VNResourceIdsField:
			nvResourceIdsStr = val
		case constant.HTTPSTrafficOnlyField:
			if strings.EqualFold(val, constant.TrueValue) {
				enableHTTPSTrafficOnly = true
			}
		case constant.CreatePrivateEndpointField:
			if strings.EqualFold(val, constant.TrueValue) {
				createPrivateEndpoint = true
			}
		case constant.HNSEnabledField:
			if strings.EqualFold(val, constant.TrueValue) {
				isHnsEnabled = true
			}
		case constant.EnableNFSV3Field:
			if strings.EqualFold(val, constant.TrueValue) {
				enableNfsV3 = true
			}
		case constant.EnableLargeFileSharesField:
			if strings.EqualFold(val, constant.TrueValue) {
				enableLargeFileShares = true
			}
		}
	}

	if accountType == "" {
		accountType = consts.DefaultStorageAccountType
	}

	tags, err := constant.ConvertTagsToMap(customTags)
	if err != nil {
		return nil, err
	}

	return &azure.AccountOptions{
		Type:                      accountType,
		Kind:                      kind,
		Location:                  location,
		EnableHTTPSTrafficOnly:    enableHTTPSTrafficOnly,
		IsHnsEnabled:              to.BoolPtr(isHnsEnabled),
		EnableLargeFileShare:      enableLargeFileShares,
		EnableNfsV3:               to.BoolPtr(enableNfsV3),
		CreatePrivateEndpoint:     createPrivateEndpoint,
		VirtualNetworkResourceIDs: strings.Split(nvResourceIdsStr, constant.TagsDelimiter),
		Tags:                      tags,
	}, nil
}
