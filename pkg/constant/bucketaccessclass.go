package constant

const (
	//BucketAccessClassFields
	SignedVersionField      = "signedversion"
	SignedPermissionsField  = "signedpermissions"
	SignedExpiryField       = "signedexpiry"
	SignedResourceTypeField = "signedresourcetype"
)

type SignedResourceType int
type SignedPermissions int

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
