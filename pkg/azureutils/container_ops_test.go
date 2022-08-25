package azureutils

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"project/azure-cosi-driver/pkg/constant"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

func TestCreateContainerBucket(t *testing.T) {
	tests := []struct {
		testName    string
		url         string
		params      *BucketClassParameters
		expectedErr error
	}{
		{
			testName: "Could not ensure account exists",
			url:      "",
			params:   &BucketClassParameters{storageAccountName: constant.InvalidAccount},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("Could not ensure storage account %s exists: %v", constant.InvalidAccount,
				fmt.Errorf("could not get storage key for storage account %s: %w", constant.InvalidAccount,
					retry.GetError(&http.Response{}, fmt.Errorf("Invalid Account")).Error()))),
		},
	}

	ctrl := gomock.NewController(t)
	cloud := azure.GetTestCloud(ctrl)
	keyList := make([]storage.AccountKey, 0)
	keyList = append(keyList, storage.AccountKey{KeyName: to.StringPtr(constant.ValidAccount), Value: to.StringPtr("val")})
	cloud.StorageAccountClient = NewMockSAClient(context.Background(), ctrl, "", "", "", &keyList)

	for _, test := range tests {
		_, err := createContainerBucket(context.Background(), test.url, test.params, cloud)
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("\nTestCase: %s\nExpected Error: %v\nActual Error: %v", test.testName, test.expectedErr, err)
		}
	}
}

func TestDeleteContainerBucket(t *testing.T) {
	tests := []struct {
		testName    string
		url         string
		clientNil   bool
		expectedErr error
	}{
		{
			testName:    "Storage Account Client is nil",
			url:         constant.ValidContainerURL,
			clientNil:   true,
			expectedErr: fmt.Errorf("StorageAccountClient is nil"),
		},
		{
			testName:    "Invalid Credentials/Key",
			url:         constant.ValidContainerURL,
			clientNil:   false,
			expectedErr: fmt.Errorf("Error deleting container %s in storage account %s : %v", constant.ValidContainer, constant.ValidAccount, fmt.Errorf("Invalid credentials with error : %s", "illegal base64 data at input byte 0")),
		},
	}

	ctrl := gomock.NewController(t)
	cloud := azure.GetTestCloud(ctrl)

	for _, test := range tests {
		if !test.clientNil {
			keyList := make([]storage.AccountKey, 0)
			keyList = append(keyList, storage.AccountKey{KeyName: to.StringPtr(constant.ValidAccount), Value: to.StringPtr("val")})
			cloud.StorageAccountClient = NewMockSAClient(context.Background(), ctrl, "", "", "", &keyList)
		} else {
			cloud.StorageAccountClient = nil
		}

		err := DeleteContainerBucket(context.Background(), test.url, cloud)
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("\nTestCase: %s\nExpected Error: %v\nActual Error: %v", test.testName, test.expectedErr, err)
		}
	}
}

func TestParseContainerURL(t *testing.T) {
	tests := []struct {
		testName              string
		url                   string
		expectedAccountName   string
		expectedContainerName string
		expectedBlobName      string
	}{
		{
			testName:              "Valid Blob URL",
			url:                   constant.ValidBlobURL,
			expectedAccountName:   constant.ValidAccount,
			expectedContainerName: constant.ValidContainer,
			expectedBlobName:      constant.ValidBlob,
		},
	}
	for _, test := range tests {
		acc, con, blob := parsecontainerurl(test.url)
		if !reflect.DeepEqual(acc, test.expectedAccountName) {
			t.Errorf("\nTestCase: %s\nExpected Account: %v\nActual Account: %v", test.testName, test.expectedAccountName, acc)
		}
		if !reflect.DeepEqual(con, test.expectedContainerName) {
			t.Errorf("\nTestCase: %s\nExpected Container: %v\nActual Container: %v", test.testName, test.expectedContainerName, con)
		}
		if !reflect.DeepEqual(blob, test.expectedBlobName) {
			t.Errorf("\nTestCase: %s\nExpected Blob: %v\nActual Blob: %v", test.testName, test.expectedBlobName, blob)
		}
	}
}

func TestGetStorageAccountNameFromContainerURL(t *testing.T) {
	tests := []struct {
		testName              string
		url                   string
		expectedContainerName string
	}{
		{
			testName:              "Valid URL",
			url:                   constant.ValidContainerURL,
			expectedContainerName: constant.ValidContainer,
		},
	}
	for _, test := range tests {
		con := getContainerNameFromContainerURL(test.url)
		if !reflect.DeepEqual(con, test.expectedContainerName) {
			t.Errorf("\nTestCase: %s\nExpected Container: %v\nActual Container: %v", test.testName, test.expectedContainerName, con)
		}
	}
}

func TestGetContainerNameFromContainerURL(t *testing.T) {
	tests := []struct {
		testName              string
		url                   string
		expectedContainerName string
	}{
		{
			testName:              "Valid URL",
			url:                   constant.ValidContainerURL,
			expectedContainerName: constant.ValidContainer,
		},
	}
	for _, test := range tests {
		con := getContainerNameFromContainerURL(test.url)
		if !reflect.DeepEqual(con, test.expectedContainerName) {
			t.Errorf("\nTestCase: %s\nExpected Container: %v\nActual Container: %v", test.testName, test.expectedContainerName, con)
		}
	}
}

func TestCreateContainerURL(t *testing.T) {
	tests := []struct {
		testName    string
		account     string
		key         string
		container   string
		expectedURL string
		expectedErr error
	}{
		{
			testName:    "Invalid Credentials/Key",
			account:     constant.ValidAccount,
			key:         "key",
			container:   constant.ValidContainer,
			expectedURL: constant.ValidContainerURL,
			expectedErr: fmt.Errorf("Invalid credentials with error : %s", "illegal base64 data at input byte 0"),
		},
		{
			testName:    "Valid URL",
			account:     constant.ValidAccount,
			key:         base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}),
			container:   constant.ValidContainer,
			expectedURL: constant.ValidContainerURL,
			expectedErr: nil,
		},
	}
	for _, test := range tests {
		url, err := createContainerURL(test.account, test.key, test.container)
		urlStr := url.String()
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("\nTestCase: %s\nExpected Error: %v\nActual Error: %v", test.testName, test.expectedErr, err)
		}
		if err == nil && !reflect.DeepEqual(urlStr, test.expectedURL) {
			t.Errorf("\nTestCase: %s\nExpected URL: %v\nActual URL: %v", test.testName, test.expectedURL, urlStr)
		}
	}
}

func TestDeleteAzureContainer(t *testing.T) {
	tests := []struct {
		testName    string
		account     string
		key         string
		container   string
		expectedErr error
	}{
		{
			testName:    "Invalid Credentials/Key (not encoded)",
			account:     constant.ValidAccount,
			key:         "key",
			container:   constant.ValidContainer,
			expectedErr: fmt.Errorf("Invalid credentials with error : %s", "illegal base64 data at input byte 0"),
		},
	}
	for _, test := range tests {
		err := deleteAzureContainer(context.Background(), test.account, test.key, test.container)
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("\nTestCase: %s\nExpected Error: %v\nActual Error: %v", test.testName, test.expectedErr, err)
		}
	}
}

func TestCreateAzureContainer(t *testing.T) {
	tests := []struct {
		testName    string
		account     string
		key         string
		container   string
		expectedURL string
		expectedErr error
	}{
		{
			testName:    "Empty Storage Account",
			account:     "",
			key:         "key",
			container:   constant.ValidContainer,
			expectedURL: "",
			expectedErr: fmt.Errorf("Invalid storage account or access key"),
		},
		{
			testName:    "Empty Access Key",
			account:     constant.ValidAccount,
			key:         "",
			container:   constant.ValidContainer,
			expectedURL: constant.ValidContainerURL,
			expectedErr: fmt.Errorf("Invalid storage account or access key"),
		},
		{
			testName:    "Invalid Credentials/Key (not encoded)",
			account:     constant.ValidAccount,
			key:         "key",
			container:   constant.ValidContainer,
			expectedURL: "",
			expectedErr: fmt.Errorf("Invalid credentials with error : %s", "illegal base64 data at input byte 0"),
		},
	}
	params := make(map[string]string)
	for _, test := range tests {
		url, err := createAzureContainer(context.Background(), test.account, test.key, test.container, params)
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("\nTestCase: %s\nExpected Error: %v\nActual Error: %v", test.testName, test.expectedErr, err)
		}
		if err == nil && !reflect.DeepEqual(url, test.expectedURL) {
			t.Errorf("\nTestCase: %s\nExpected URL: %v\nActual URL: %v", test.testName, test.expectedURL, url)
		}
	}
}