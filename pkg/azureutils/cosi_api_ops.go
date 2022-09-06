package azureutils

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"project/azure-cosi-driver/pkg/constant"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/to"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

type BucketClassParameters struct {
	bucketUnitType                 constant.BucketUnitType
	createBucket                   bool
	createStorageAccount           bool
	storageAccountName             string
	region                         string
	accessTier                     constant.AccessTier
	SKUName                        constant.SKU
	resourceGroup                  string
	allowBlobAccess                bool
	allowSharedAccessKey           bool
	enableBlobVersioning           bool
	enableBlobDeleteRetention      bool
	blobDeleteRetentionDays        int
	enableContainerDeleteRetention bool
	containerDeleteRetentionDays   int
	//account options
	storageAccountType        string
	kind                      constant.Kind
	tags                      map[string]string
	virtualNetworkResourceIDs []string
	enableHTTPSTrafficOnly    bool
	createPrivateEndpoint     bool
	isHnsEnabled              bool
	enableNfsV3               bool
	enableLargeFileShare      bool
}

/*
signed expiry is done in days in reference to start, while signed start is written as a date
start should be written in ISO 8601 UTC. default will be the immediate time/date
YYYY-MM-DD
YYYY-MM-DDThh:mm<TZDSuffix>
YYYY-MM-DDThh:mm:ss<TZDSuffix>
*/
type BucketAccessClassParameters struct {
	bucketUnitType                   constant.BucketUnitType
	storageAccountName               string
	region                           string
	signedversion                    string
	signedIP                         azblob.IPRange
	validationPeriod                 uint64
	signedProtocol                   azblob.SASProtocol
	enableList                       bool
	enableRead                       bool
	enableWrite                      bool
	enablePermanentDelete            bool
	allowServiceSignedResourceType   bool
	allowContainerSignedResourceType bool
	allowObjectSignedResourceType    bool
	key                              string
}

func CreateBucket(ctx context.Context,
	bucketName string,
	parameters map[string]string,
	cloud *azure.Cloud) (string, error) {
	bucketClassParams, err := parseBucketClassParameters(parameters)
	if err != nil {
		return "", status.Error(codes.Unknown, fmt.Sprintf("Error parsing parameters : %v", err))
	}

	switch bucketClassParams.bucketUnitType {
	case constant.Container:
		return createContainerBucket(ctx, bucketName, bucketClassParams, cloud)
	case constant.StorageAccount:
		return createStorageAccountBucket(ctx, bucketName, bucketClassParams, cloud)
	}
	return "", status.Error(codes.InvalidArgument, "Invalid BucketUnitType")
}

func DeleteBucket(ctx context.Context,
	bucketID string,
	cloud *azure.Cloud) error {
	//determine if the bucket is an account or a blob container
	account, container, blob := parsecontainerurl(bucketID)
	if account == "" {
		return status.Error(codes.InvalidArgument, "Storage Account required")
	}
	if blob != "" {
		return status.Error(codes.InvalidArgument, "Individual Blobs unsupported. Please use Blob Containers or Storage Accounts instead.")
	}

	klog.Infof("DriverDeleteBucket :: Bucket id :: %s", bucketID)
	var err error
	if container == "" { //container not present, deleting storage account
		err = DeleteStorageAccount(ctx, account, cloud)
	} else { //container name present, deleting container
		err = DeleteContainerBucket(ctx, bucketID, cloud)
	}
	return err
}

// creates bucketSASURL and returns (SASURL, accountID, err)
func CreateBucketSASURL(ctx context.Context, bucketID string, parameters map[string]string) (string, string, error) {
	bucketAccessClassParams, err := parseBucketAccessClassParameters(parameters)
	if err != nil {
		return "", "", err
	}

	switch bucketAccessClassParams.bucketUnitType {
	case constant.Container:
		return createContainerSASURL(ctx, bucketID, bucketAccessClassParams)
	case constant.StorageAccount:
		return createAccountSASURL(ctx, bucketID, bucketAccessClassParams)
	}
	return "", "", status.Error(codes.InvalidArgument, "invalid bucket type")
}

func parseBucketClassParameters(parameters map[string]string) (*BucketClassParameters, error) {
	BCParams := &BucketClassParameters{}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case constant.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case constant.Container.String(), "":
				BCParams.bucketUnitType = constant.Container
			case constant.StorageAccount.String():
				BCParams.bucketUnitType = constant.StorageAccount
			default:
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case constant.CreateBucketField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.createBucket = true
			} else {
				BCParams.createBucket = false
			}
		case constant.CreateStorageAccountField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.createStorageAccount = true
			} else {
				BCParams.createStorageAccount = false
			}
		case constant.StorageAccountNameField:
			BCParams.storageAccountName = v
		case constant.RegionField:
			BCParams.region = v
		case constant.AccessTierField:
			switch strings.ToLower(v) {
			case constant.Hot.String():
				BCParams.accessTier = constant.Hot
			case constant.Cool.String():
				BCParams.accessTier = constant.Cool
			case constant.Archive.String():
				BCParams.accessTier = constant.Archive
			default:
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case constant.SKUNameField:
			switch strings.ToLower(v) {
			case strings.ToLower(constant.StandardLRS.String()):
				BCParams.SKUName = constant.StandardLRS
			case strings.ToLower(constant.StandardGRS.String()):
				BCParams.SKUName = constant.StandardGRS
			case strings.ToLower(constant.StandardRAGRS.String()):
				BCParams.SKUName = constant.StandardRAGRS
			case strings.ToLower(constant.PremiumLRS.String()):
				BCParams.SKUName = constant.PremiumLRS
			default:
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case constant.ResourceGroupField:
			BCParams.resourceGroup = v
		case constant.AllowBlobAccessField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.allowBlobAccess = true
			} else {
				BCParams.allowBlobAccess = false
			}
		case constant.AllowSharedAccessKeyField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.allowSharedAccessKey = true
			} else {
				BCParams.allowSharedAccessKey = false
			}
		case constant.EnableBlobVersioningField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableBlobVersioning = true
			} else {
				BCParams.enableBlobVersioning = false
			}
		case constant.EnableBlobDeleteRetentionField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableBlobDeleteRetention = true
			} else {
				BCParams.enableBlobDeleteRetention = false
			}
		case constant.BlobDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.blobDeleteRetentionDays = days
		case constant.EnableContainerDeleteRetentionField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableContainerDeleteRetention = true
			} else {
				BCParams.enableContainerDeleteRetention = false
			}
		case constant.ContainerDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.containerDeleteRetentionDays = days
		case StorageAccountTypeField: //Account Options Variables
			BCParams.storageAccountType = v
		case KindField:
			switch strings.ToLower(v) {
			case strings.ToLower(constant.StorageV2.String()):
				BCParams.kind = constant.StorageV2
			case strings.ToLower(constant.Storage.String()):
				BCParams.kind = constant.Storage
			case strings.ToLower(constant.BlobStorage.String()):
				BCParams.kind = constant.BlobStorage
			case strings.ToLower(constant.BlockBlobStorage.String()):
				BCParams.kind = constant.BlockBlobStorage
			case strings.ToLower(constant.FileStorage.String()):
				BCParams.kind = constant.FileStorage
			default:
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Account Kind %s is unsupported", v))
			}
		case TagsField:
			tags, err := ConvertTagsToMap(v)
			if err != nil {
				return nil, err
			}
			BCParams.tags = tags
		case VNResourceIdsField:
			BCParams.virtualNetworkResourceIDs = strings.Split(v, TagsDelimiter)
		case HTTPSTrafficOnlyField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableHTTPSTrafficOnly = true
			}
		case CreatePrivateEndpointField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.createPrivateEndpoint = true
			}
		case HNSEnabledField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.isHnsEnabled = true
			}
		case EnableNFSV3Field:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableNfsV3 = true
			}
		case EnableLargeFileSharesField:
			if strings.EqualFold(v, TrueValue) {
				BCParams.enableLargeFileShare = true
			}
		}
	}
	return BCParams, nil
}

func parseBucketAccessClassParameters(parameters map[string]string) (*BucketAccessClassParameters, error) {
	//defaults
	// validation period default = one week
	BACParams := &BucketAccessClassParameters{
		validationPeriod:                 604800000,
		signedProtocol:                   azblob.SASProtocolHTTPSandHTTP,
		enableRead:                       true,
		enableList:                       true,
		allowServiceSignedResourceType:   true,
		allowContainerSignedResourceType: true,
		allowObjectSignedResourceType:    true,
	}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case constant.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BACParams.bucketUnitType = constant.Container
			case "":
				BACParams.bucketUnitType = constant.Container
			case "storageaccount":
				BACParams.bucketUnitType = constant.StorageAccount
			default:
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case constant.StorageAccountNameField:
			BACParams.storageAccountName = v
		case constant.RegionField:
			BACParams.region = v
		case constant.SignedVersionField:
			BACParams.signedversion = v
		case constant.SignedProtocolField:
			switch v {
			case string(azblob.SASProtocolHTTPS):
				BACParams.signedProtocol = azblob.SASProtocolHTTPS
			case string(azblob.SASProtocolHTTPSandHTTP):
				BACParams.signedProtocol = azblob.SASProtocolHTTPSandHTTP
			}
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid SAS Protocol %s", v))
		case constant.SignedIPField:
			iplist := strings.Split(v, "-")
			switch len(iplist) {
			case 1:
				start := net.ParseIP(iplist[0])
				if start == nil {
					klog.Warning(fmt.Sprintf("IP %s is an invalid ip, no range will be set", iplist[0]))
				}
				BACParams.signedIP = azblob.IPRange{Start: start}
			case 2:
				start := net.ParseIP(iplist[0])
				end := net.ParseIP(iplist[1])
				if start == nil {
					klog.Warning(fmt.Sprintf("IP %s is an invalid ip, no range will be set", iplist[0]))
				}
				if end == nil {
					klog.Warning(fmt.Sprintf("IP %s is an invalid ip, no end to the range will be set", iplist[0]))
				}
				BACParams.signedIP = azblob.IPRange{Start: start, End: end}
			}
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid IP Range %s, Must be formatted as <ip> or <ip1>-<ip2>", v))
		case constant.ValidationPeriodField:
			msec, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			BACParams.validationPeriod = msec
		case constant.EnableListField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.enableList = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.enableList = false
			}
		case constant.EnableReadField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.enableRead = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.enableRead = false
			}
		case constant.EnableWriteField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.enableWrite = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.enableWrite = false
			}
		case constant.EnablePermanentDeleteField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.enablePermanentDelete = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.enablePermanentDelete = false
			}
		case constant.AllowServiceSignedResourceTypeField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.allowServiceSignedResourceType = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.allowServiceSignedResourceType = false
			}
		case constant.AllowContainerSignedResourceTypeField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.allowObjectSignedResourceType = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.allowObjectSignedResourceType = false
			}
		case constant.AllowObjectSignedResourceTypeField:
			if strings.EqualFold(v, TrueValue) {
				BACParams.allowObjectSignedResourceType = true
			} else if strings.EqualFold(v, FalseValue) {
				BACParams.allowObjectSignedResourceType = false
			}
		case constant.KeyField:
			BACParams.key = v
		}
	}
	return BACParams, nil
}

func getAccountOptions(params *BucketClassParameters) *azure.AccountOptions {
	options := &azure.AccountOptions{
		Name:                      params.storageAccountName,
		ResourceGroup:             params.resourceGroup,
		Location:                  params.region,
		Type:                      params.storageAccountType,
		Kind:                      params.kind.String(),
		Tags:                      params.tags,
		VirtualNetworkResourceIDs: params.virtualNetworkResourceIDs,
		EnableHTTPSTrafficOnly:    params.enableHTTPSTrafficOnly,
		CreatePrivateEndpoint:     params.createPrivateEndpoint,
		IsHnsEnabled:              to.BoolPtr(params.isHnsEnabled),
		EnableNfsV3:               to.BoolPtr(params.enableNfsV3),
		EnableLargeFileShare:      params.enableLargeFileShare,
	}
	return options
}
