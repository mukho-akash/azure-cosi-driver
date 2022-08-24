package azureutils

import (
	"encoding/base64"
	"fmt"
	"project/azure-cosi-driver/pkg/constant"
	"reflect"
	"testing"
)

func TestCreateContainerBucket(t *testing.T) {

}

func TestDeleteContainerBucket(t *testing.T) {

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

func TestCreateAzureContainer(t *testing.T) {

}
