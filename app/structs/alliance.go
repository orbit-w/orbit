// Code generated from proto files. DO NOT EDIT.

package pb

// AllianceEntity represents the proto message AllianceEntity
type AllianceEntity struct {
	Id int64 `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
}

// DeepCopy creates a deep copy of AllianceEntity
func (s *AllianceEntity) DeepCopy() *AllianceEntity {
	if s == nil {
		return nil
	}

	co := &AllianceEntity{}
	*co = *s

	return co
}
