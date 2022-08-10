package constant

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
	AllowSharedAccessKeyField           = "allowsharedaccesskey"
	EnableBlobVersioningField           = "enableblobversioning"
	EnableBlobDeleteRetentionField      = "enableblobdeleteretention"
	BlobDeleteRetentionDaysField        = "blobdeleteretentiondays"
	EnableContainerDeleteRetentionField = "enablecontainerdeleteretention"
	ContainerDeleteRetentionDaysField   = "containerdeleteretentiondays"
)

type BucketUnitType int
type AccessTier int
type SKU int

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
