package azureutils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"project/azure-cosi-driver/pkg/constant"

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
	enableHttpsTrafficOnly    bool
	createPrivateEndpoint     bool
	isHnsEnabled              bool
	enableNfsV3               bool
	enableLargeFileShare      bool
}

type BucketAccessClassParameters struct {
	bucketUnitType    constant.BucketUnitType
	region            string
	signedversion     string
	signedPermissions constant.SignedPermissions
	signedExpiry      int
	signedResouceType constant.SignedResourceType
}

func CreateBucket(ctx context.Context,
	bucketName string,
	parameters map[string]string,
	cloud *azure.Cloud) (string, error) {
	bucketClassParams, err := parseBucketClassParameters(parameters)
	if err != nil {
		return "", status.Error(codes.Unknown, fmt.Sprintf("Error parsing parameters : %v", err))
	}

	if bucketClassParams.bucketUnitType == constant.Container {
		return createContainerBucket(ctx, bucketName, bucketClassParams, cloud)
	} else {
		return createStorageAccountBucket(ctx, bucketName, bucketClassParams, cloud)
	}
}

func DeleteBucket(ctx context.Context,
	bucketId string,
	cloud *azure.Cloud) error {
	//determine if the bucket is an account or a blob container
	account, container, blob := parsecontainerurl(bucketId)
	if account == "" {
		return status.Error(codes.InvalidArgument, "Storage Account required")
	}
	if blob != "" {
		return status.Error(codes.InvalidArgument, "Individual Blobs unsupported. Please use Blob Containers or Storage Accounts instead.")
	}

	klog.Infof("DriverDeleteBucket :: Bucket id :: %s", bucketId)
	var err error
	if container == "" { //container not present, deleting storage account
		err = DeleteStorageAccount(ctx, account, cloud)
	} else { //container name present, deleting container
		err = DeleteContainerBucket(ctx, bucketId, cloud)
	}
	return err
}

func parseBucketClassParameters(parameters map[string]string) (*BucketClassParameters, error) {
	BCParams := &BucketClassParameters{}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case constant.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container", "":
				BCParams.bucketUnitType = constant.Container
			case "storageaccount":
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
			case strings.ToLower(constant.Standard_LRS.String()):
				BCParams.SKUName = constant.Standard_LRS
			case strings.ToLower(constant.Standard_GRS.String()):
				BCParams.SKUName = constant.Standard_GRS
			case strings.ToLower(constant.Standard_RAGRS.String()):
				BCParams.SKUName = constant.Standard_RAGRS
			case strings.ToLower(constant.Premium_LRS.String()):
				BCParams.SKUName = constant.Premium_LRS
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
				BCParams.enableHttpsTrafficOnly = true
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
	if BCParams.bucketUnitType == constant.StorageAccount {
		return BCParams, nil
	} else {
		return BCParams, nil
	}
}

func parseBucketAccessClassParameters(parameters map[string]string) (*BucketAccessClassParameters, error) {
	BACParams := &BucketAccessClassParameters{}
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
		case constant.RegionField:
			BACParams.region = v
		case constant.SignedVersionField:
			BACParams.signedversion = v
		case constant.SignedPermissionsField:
			switch strings.ToLower(v) {
			case constant.Read.String():
				BACParams.signedPermissions = constant.Read
			case constant.Add.String():
				BACParams.signedPermissions = constant.Add
			case constant.Create.String():
				BACParams.signedPermissions = constant.Create
			case constant.Write.String():
				BACParams.signedPermissions = constant.Write
			case constant.Delete.String():
				BACParams.signedPermissions = constant.Delete
			case constant.ReadWrite.String():
				BACParams.signedPermissions = constant.ReadWrite
			case constant.AddDelete.String():
				BACParams.signedPermissions = constant.AddDelete
			case constant.List.String():
				BACParams.signedPermissions = constant.List
			case constant.DeleteVersion.String():
				BACParams.signedPermissions = constant.Delete
			case constant.PermanentDelete.String():
				BACParams.signedPermissions = constant.PermanentDelete
			case constant.All.String():
				BACParams.signedPermissions = constant.All
			}
		case constant.SignedExpiryField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			BACParams.signedExpiry = days
		case constant.SignedResourceTypeField:
			switch strings.ToLower(v) {
			case constant.TypeObject.String():
				BACParams.signedResouceType = constant.TypeObject
			case constant.TypeContainer.String():
				BACParams.signedResouceType = constant.TypeContainer
			case constant.TypeObjectAndContainer.String():
				BACParams.signedResouceType = constant.TypeObjectAndContainer
			}
		}
	}
	return BACParams, nil
}

func getAccountOptions(params *BucketClassParameters) *azure.AccountOptions {
	options := &azure.AccountOptions{
		Type:                      params.storageAccountType,
		Kind:                      params.kind.String(),
		Tags:                      params.tags,
		VirtualNetworkResourceIDs: params.virtualNetworkResourceIDs,
		EnableHTTPSTrafficOnly:    params.enableHttpsTrafficOnly,
		CreatePrivateEndpoint:     params.createPrivateEndpoint,
		IsHnsEnabled:              to.BoolPtr(params.isHnsEnabled),
		EnableNfsV3:               to.BoolPtr(params.enableNfsV3),
		EnableLargeFileShare:      params.enableLargeFileShare,
	}
	return options
}
