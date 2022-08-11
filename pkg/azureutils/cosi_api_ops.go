package azureutils

import (
	"fmt"
	"strconv"
	"strings"

	"project/azure-cosi-driver/pkg/constant"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	bucketClassParams, err := ParseBucketClassParameters(parameters)
	if err != nil {
		return "", status.Error(codes.Unknown, fmt.Sprintf("Error parsing parameters : %v", err))
	}

	if bucketClassParams.bucketUnitType == constant.Container {
		return createContainerBucket(ctx, bucketName, &bucketClassParams, cloud)
	} else {
		return createStorageAccountBucket(ctx, bucketName, &bucketClassParams, cloud)
	}
}

func ParseBucketClassParameters(parameters map[string]string) (BucketClassParameters, error) {
	BCParams := BucketClassParameters{}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case constant.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BCParams.bucketUnitType = constant.Container
			case "":
				BCParams.bucketUnitType = constant.Container
			case "storageaccount":
				BCParams.bucketUnitType = constant.StorageAccount
			default:
				return BucketClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case constant.CreateBucketField:
			if constant.TrueValue == v {
				BCParams.createBucket = true
			} else {
				BCParams.createBucket = false
			}
		case constant.CreateStorageAccountField:
			if constant.TrueValue == v {
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
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
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
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case constant.ResourceGroupField:
			BCParams.resourceGroup = v
		case constant.AllowBlobAccessField:
			if constant.TrueValue == v {
				BCParams.allowBlobAccess = true
			} else {
				BCParams.allowBlobAccess = false
			}
		case constant.AllowSharedAccessKeyField:
			if constant.TrueValue == v {
				BCParams.allowSharedAccessKey = true
			} else {
				BCParams.allowSharedAccessKey = false
			}
		case constant.EnableBlobVersioningField:
			if constant.TrueValue == v {
				BCParams.enableBlobVersioning = true
			} else {
				BCParams.enableBlobVersioning = false
			}
		case constant.EnableBlobDeleteRetentionField:
			if constant.TrueValue == v {
				BCParams.enableBlobDeleteRetention = true
			} else {
				BCParams.enableBlobDeleteRetention = false
			}
		case constant.BlobDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BCParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.blobDeleteRetentionDays = days
		case constant.EnableContainerDeleteRetentionField:
			if constant.TrueValue == v {
				BCParams.enableContainerDeleteRetention = true
			} else {
				BCParams.enableContainerDeleteRetention = false
			}
		case constant.ContainerDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BCParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.containerDeleteRetentionDays = days
		}
	}
	return BCParams, nil
}

func ParseBucketAccessClassParameters(parameters map[string]string) (BucketAccessClassParameters, error) {
	BACParams := BucketAccessClassParameters{}
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
				return BucketAccessClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
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
				return BACParams, status.Error(codes.InvalidArgument, err.Error())
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
