package types

import (
	"encoding/base64"
	"encoding/json"
)

// bucketID is passed in DriverCreateBucket as a struct with subID and resource group for deletion
type BucketID struct {
	SubID         string `json:"subscriptionID"`
	ResourceGroup string `json:"resourceGroup"`
	URL           string `json:"url"`
}

// Marshals bucketID struct into json bytes, then encodes into base64
func (id *BucketID) Encode() (string, error) {
	data, err := json.Marshal(id)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// Decodes base64 string to bucketID pointer struct
func DecodeToBucketID(id string) (*BucketID, error) {
	data, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}

	bID := &BucketID{}
	err = json.Unmarshal(data, bID)
	if err != nil {
		return nil, err
	}
	return bID, nil
}
