package servicezone_behavior

import "fmt"

var (
	ErrEntityRefNil              = fmt.Errorf("entity ref is nil")
	ErrEntityFactoryNotFound     = fmt.Errorf("entity factory not found")
	ErrEntityDataFactoryNotFound = fmt.Errorf("entity data factory not found")
	ErrEntityDataUnmarshalFailed = fmt.Errorf("entity data unmarshal failed")
	ErrEntityLoadFailed          = fmt.Errorf("entity load failed")
)
