package azureutils

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	//BucketClassParameterFields
	BucketUnitTypeField                 = "bucketunittype"
	CreateBucketField                   = "createbucket"
	CreateStorageAccountField           = "createstorageaccount"
	StorageAccountNameField             = "storageaccountname"
	ContainerNameField                  = "containername"
	RegionField                         = "region"
	AccessTierField                     = "accesstier"
	SKUNameField                        = "skuname"
	ResourceGroupField                  = "resourcegroup"
	AllowBlobAccessField                = "allowblobaccess"
	allowSharedAccessKeyField           = "allowsharedaccesskey"
	EnableBlobVersioningField           = "enableblobversioning"
	EnableBlobDeleteRetentionField      = "enableblobdeleteretention"
	BlobDeleteRetentionDaysField        = "blobdeleteretentiondays"
	EnableContainerDeleteRetentionField = "enablecontainerdeleteretention"
	ContainerDeleteRetentionDaysField   = "containerdeleteretentiondays"

	//BucketAccessClassFields
	SignedVersionField      = "signedversion"
	SignedPermissionsField  = "signedpermissions"
	SignedExpiryField       = "signedexpiry"
	SignedResourceTypeField = "signedresourcetype"
)

//defining enums
type BucketUnitType int
type AccessTier int
type SKU int
type SignedResourceType int
type SignedPermissions int

const (
	Container BucketUnitType = iota
	StorageAccount
)

const (
	Hot AccessTier = iota
	Cool
	Archive
)

const (
	Standard_LRS SKU = iota
	Standard_GRS
	Standard_RAGRS
	Premium_LRS
)

const (
	TypeObject SignedResourceType = iota
	TypeContainer
	TypeObjectAndContainer
)

const (
	Read SignedPermissions = iota
	Add
	Create
	Write
	Delete
	ReadWrite
	AddDelete
	List
	DeleteVersion
	PermanentDelete
	All
)

type BucketClassParameters struct {
	bucketUnitType                 BucketUnitType
	createBucket                   bool
	createStorageAccount           bool
	storageAccountName             string
	region                         string
	accessTier                     AccessTier
	SKUName                        SKU
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
	bucketUnitType    BucketUnitType
	region            string
	signedversion     string
	signedPermissions SignedPermissions
	signedExpiry      int
	signedResouceType SignedResourceType
}

func ParseBucketClassParameters(parameters map[string]string) (BucketClassParameters, error) {
	BCParams := BucketClassParameters{}
	for k, v := range parameters {
		switch strings.ToLower(k) {
		case BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BCParams.bucketUnitType = Container
			case "":
				BCParams.bucketUnitType = Container
			case "storageaccount":
				BCParams.bucketUnitType = StorageAccount
			default:
				return BucketClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case CreateBucketField:
			if TrueValue == v {
				BCParams.createBucket = true
			} else {
				BCParams.createBucket = false
			}
		case CreateStorageAccountField:
			if TrueValue == v {
				BCParams.createStorageAccount = true
			} else {
				BCParams.createStorageAccount = false
			}
		case StorageAccountNameField:
			BCParams.storageAccountName = v
		case RegionField:
			BCParams.region = v
		case AccessTierField:
			switch strings.ToLower(v) {
			case Hot.String():
				BCParams.accessTier = Hot
			case Cool.String():
				BCParams.accessTier = Cool
			case Archive.String():
				BCParams.accessTier = Archive
			default:
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case SKUNameField:
			switch strings.ToLower(v) {
			case strings.ToLower(Standard_LRS.String()):
				BCParams.SKUName = Standard_LRS
			case strings.ToLower(Standard_GRS.String()):
				BCParams.SKUName = Standard_GRS
			case strings.ToLower(Standard_RAGRS.String()):
				BCParams.SKUName = Standard_RAGRS
			case strings.ToLower(Premium_LRS.String()):
				BCParams.SKUName = Premium_LRS
			default:
				return BCParams, status.Error(codes.InvalidArgument, fmt.Sprintf("Access Tier %s is unsupported", v))
			}
		case ResourceGroupField:
			BCParams.resourceGroup = v
		case AllowBlobAccessField:
			if TrueValue == v {
				BCParams.allowBlobAccess = true
			} else {
				BCParams.allowBlobAccess = false
			}
		case allowSharedAccessKeyField:
			if TrueValue == v {
				BCParams.allowSharedAccessKey = true
			} else {
				BCParams.allowSharedAccessKey = false
			}
		case EnableBlobVersioningField:
			if TrueValue == v {
				BCParams.enableBlobVersioning = true
			} else {
				BCParams.enableBlobVersioning = false
			}
		case EnableBlobDeleteRetentionField:
			if TrueValue == v {
				BCParams.enableBlobDeleteRetention = true
			} else {
				BCParams.enableBlobDeleteRetention = false
			}
		case BlobDeleteRetentionDaysField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BCParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BCParams.blobDeleteRetentionDays = days
		case EnableContainerDeleteRetentionField:
			if TrueValue == v {
				BCParams.enableContainerDeleteRetention = true
			} else {
				BCParams.enableContainerDeleteRetention = false
			}
		case ContainerDeleteRetentionDaysField:
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
		case BucketUnitTypeField:
			//determine unit type and set to container as default if blank
			switch strings.ToLower(v) {
			case "container":
				BACParams.bucketUnitType = Container
			case "":
				BACParams.bucketUnitType = Container
			case "storageaccount":
				BACParams.bucketUnitType = StorageAccount
			default:
				return BucketAccessClassParameters{}, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid BucketUnitType %s", v))
			}
		case RegionField:
			BACParams.region = v
		case SignedVersionField:
			BACParams.signedversion = v
		case SignedPermissionsField:
			switch strings.ToLower(v) {
			case Read.String():
				BACParams.signedPermissions = Read
			case Add.String():
				BACParams.signedPermissions = Add
			case Create.String():
				BACParams.signedPermissions = Create
			case Write.String():
				BACParams.signedPermissions = Write
			case Delete.String():
				BACParams.signedPermissions = Delete
			case ReadWrite.String():
				BACParams.signedPermissions = ReadWrite
			case AddDelete.String():
				BACParams.signedPermissions = AddDelete
			case List.String():
				BACParams.signedPermissions = List
			case DeleteVersion.String():
				BACParams.signedPermissions = Delete
			case PermanentDelete.String():
				BACParams.signedPermissions = PermanentDelete
			case All.String():
				BACParams.signedPermissions = All
			}
		case SignedExpiryField:
			days, err := strconv.Atoi(v)
			if err != nil {
				return BACParams, status.Error(codes.InvalidArgument, err.Error())
			}
			BACParams.signedExpiry = days
		case SignedResourceTypeField:
			switch strings.ToLower(v) {
			case TypeObject.String():
				BACParams.signedResouceType = TypeObject
			case TypeContainer.String():
				BACParams.signedResouceType = TypeContainer
			case TypeObjectAndContainer.String():
				BACParams.signedResouceType = TypeObjectAndContainer
			}
		}
	}
	return BACParams, nil
}

func (s SKU) String() string {
	switch s {
	case Standard_LRS:
		return "Standard_LRS"
	case Standard_GRS:
		return "Standard_GRS"
	case Standard_RAGRS:
		return "Standard_RAGRS"
	case Premium_LRS:
		return "Premium_LRS"
	}
	return "unknown"
}

func (a AccessTier) String() string {
	switch a {
	case Hot:
		return "hot"
	case Cool:
		return "cool"
	case Archive:
		return "archive"
	}
	return "unknown"
}

func (s SignedResourceType) String() string {
	switch s {
	case TypeObject:
		return "object"
	case TypeContainer:
		return "container"
	case TypeObjectAndContainer:
		return "objectandcontainer"
	}
	return "unknown"
}

func (s SignedPermissions) String() string {
	switch s {
	case Read:
		return "read"
	case Add:
		return "add"
	case Create:
		return "create"
	case Write:
		return "write"
	case Delete:
		return "delete"
	case ReadWrite:
		return "readwrite"
	case AddDelete:
		return "adddelete"
	case List:
		return "list"
	case DeleteVersion:
		return "deleteversion"
	case PermanentDelete:
		return "permanentdelete"
	case All:
		return "all"
	}
	return "unknown"
}
