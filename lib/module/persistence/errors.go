package persistence

import "errors"

var (
	// ErrInvalidRequest 无效的请求
	ErrInvalidRequest = errors.New("invalid persistence request")
	// ErrInvalidResponseReceiver 响应接收者为空
	ErrInvalidResponseReceiver = errors.New("invalid response receiver")
	// ErrWrapperIsNil Wrapper为nil
	ErrWrapperIsNil = errors.New("wrapper is nil")
	// ErrEmptyCollection 集合名称为空
	ErrEmptyCollection = errors.New("collection name is empty")
	// ErrEmptyDocumentID 文档ID为空
	ErrEmptyDocumentID = errors.New("document id is empty")
	// ErrDBNotInitialized 数据库未初始化
	ErrDBNotInitialized = errors.New("database not initialized")
)
