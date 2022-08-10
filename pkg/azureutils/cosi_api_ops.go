package azureutils

import (
	"fmt"
	"strconv"
	"strings"

	cons "project/azure-cosi-driver/pkg/constant"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BucketClassParameters struct {
	bucketUnitType                 cons.BucketUnitType
	createBucket                   bool
	createStorageAccount           bool
	storageAccountName             string
	region                         string
	accessTier                     cons.AccessTier
	SKUName                        cons.SKU
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
	bucketUnitType    cons.BucketUnitType
	region            string
	signedversion     string
	signedPermissions cons.SignedPermissions
	signedExpiry      int
	signedResouceType cons.SignedResourceType
}

func ParseBucketClassParameters(parameters map[string]string) (BucketClassParameters, error) {
	BCParams := BucketClassParameters{}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case cons.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BCParams.bucketUnitType = cons.Container
			case "":
				BCParams.bucketUnitType = cons.Container
			case "storageaccount":
				BCParams.bucketUnitType = cons.StorageAccount
			default:
				return BucketClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case cons.CreateBucketField:
			if cons.TrueValue == v {
				BCParams.createBucket = true
			} else {
				BCParams.createBucket = false
			}
		case cons.CreateStorageAccountField:
			if cons.TrueValue == v {
				BCParams.createStorageAccount = true
			} else {
				BCParams.createStorageAccount = false
			}
		case cons.StorageAccountNameField:
			BCParams.storageAccountName = v
		case cons.RegionField:
			BCParams.region = v
		case cons.AccessTierField:
			switch strings.ToLower(v) {
			case cons.Hot.String():
				BCParams.accessTier = cons.Hot
			case cons.Cool.String():
				BCParams.accessTier = cons.Cool
			case cons.Archive.String():
				BCParams.accessTier = cons.Archive
			default:
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case cons.SKUNameField:
			switch strings.ToLower(v) {
			case strings.ToLower(cons.Standard_LRS.String()):
				BCParams.SKUName = cons.Standard_LRS
			case strings.ToLower(cons.Standard_GRS.String()):
				BCParams.SKUName = cons.Standard_GRS
			case strings.ToLower(cons.Standard_RAGRS.String()):
				BCParams.SKUName = cons.Standard_RAGRS
			case strings.ToLower(cons.Premium_LRS.String()):
				BCParams.SKUName = cons.Premium_LRS
			default:
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case cons.ResourceGroupField:
			BCParams.resourceGroup = v
		case cons.AllowBlobAccessField:
			if cons.TrueValue == v {
				BCParams.allowBlobAccess = true
			} else {
				BCParams.allowBlobAccess = false
			}
		case cons.AllowSharedAccessKeyField:
			if cons.TrueValue == v {
				BCParams.allowSharedAccessKey = true
			} else {
				BCParams.allowSharedAccessKey = false
			}
		case cons.EnableBlobVersioningField:
			if cons.TrueValue == v {
				BCParams.enableBlobVersioning = true
			} else {
				BCParams.enableBlobVersioning = false
			}
		case cons.EnableBlobDeleteRetentionField:
			if cons.TrueValue == v {
				BCParams.enableBlobDeleteRetention = true
			} else {
				BCParams.enableBlobDeleteRetention = false
			}
		case cons.BlobDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BCParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.blobDeleteRetentionDays = days
		case cons.EnableContainerDeleteRetentionField:
			if cons.TrueValue == v {
				BCParams.enableContainerDeleteRetention = true
			} else {
				BCParams.enableContainerDeleteRetention = false
			}
		case cons.ContainerDeleteRetentionDaysField:
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
		case cons.BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BACParams.bucketUnitType = cons.Container
			case "":
				BACParams.bucketUnitType = cons.Container
			case "storageaccount":
				BACParams.bucketUnitType = cons.StorageAccount
			default:
				return BucketAccessClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case cons.RegionField:
			BACParams.region = v
		case cons.SignedVersionField:
			BACParams.signedversion = v
		case cons.SignedPermissionsField:
			switch strings.ToLower(v) {
			case cons.Read.String():
				BACParams.signedPermissions = cons.Read
			case cons.Add.String():
				BACParams.signedPermissions = cons.Add
			case cons.Create.String():
				BACParams.signedPermissions = cons.Create
			case cons.Write.String():
				BACParams.signedPermissions = cons.Write
			case cons.Delete.String():
				BACParams.signedPermissions = cons.Delete
			case cons.ReadWrite.String():
				BACParams.signedPermissions = cons.ReadWrite
			case cons.AddDelete.String():
				BACParams.signedPermissions = cons.AddDelete
			case cons.List.String():
				BACParams.signedPermissions = cons.List
			case cons.DeleteVersion.String():
				BACParams.signedPermissions = cons.Delete
			case cons.PermanentDelete.String():
				BACParams.signedPermissions = cons.PermanentDelete
			case cons.All.String():
				BACParams.signedPermissions = cons.All
			}
		case cons.SignedExpiryField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BACParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BACParams.signedExpiry = days
		case cons.SignedResourceTypeField:
			switch strings.ToLower(v) {
			case cons.TypeObject.String():
				BACParams.signedResouceType = cons.TypeObject
			case cons.TypeContainer.String():
				BACParams.signedResouceType = cons.TypeContainer
			case cons.TypeObjectAndContainer.String():
				BACParams.signedResouceType = cons.TypeObjectAndContainer
			}
		}
	}
	return BACParams, nil
}
