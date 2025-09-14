package dispatch

import (
	"fmt"
	"reflect"
	"strings"

	"gitee.com/orbit-w/orbit/app/proto/pb"
	"gitee.com/orbit-w/orbit/lib/base/protoid"
	"github.com/gogo/protobuf/proto"
)

// MethodInfo 存储方法的反射信息
type MethodInfo struct {
	Method      reflect.Method
	Controller  reflect.Value
	RequestType reflect.Type
}

// ReflectionRouter 反射路由器，实现了 pb.RequestHandler 接口
type ReflectionRouter struct {
	// methodMap 存储协议ID到方法信息的映射
	methodMap map[uint32]*MethodInfo
}

// NewReflectionRouter 创建新的反射路由器
func NewReflectionRouter() *ReflectionRouter {
	router := &ReflectionRouter{
		methodMap: make(map[uint32]*MethodInfo),
	}
	return router
}

// RegisterController 注册控制器，通过反射自动发现所有Handle开头的方法
func (r *ReflectionRouter) RegisterController(controller any) error {
	controllerValue := reflect.ValueOf(controller)
	controllerType := reflect.TypeOf(controller)

	// 遍历控制器的所有方法
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)

		// 只处理以"Handle"开头的方法
		if !strings.HasPrefix(method.Name, "Handle") {
			continue
		}

		// 验证方法签名: func (receiver) HandleXXX(req *pb.Request_XXX) proto.Message
		if err := r.validateMethodSignature(method); err != nil {
			return fmt.Errorf("method %s signature invalid: %w", method.Name, err)
		}

		// 获取请求类型
		requestType := method.Type.In(1) // 第0个参数是receiver，第1个是请求参数

		// 根据请求类型名称获取协议ID
		requestTypeName := r.getRequestTypeName(requestType)

		pid := protoid.HashProtoMessage(requestTypeName)

		// 注册方法信息
		r.methodMap[pid] = &MethodInfo{
			Method:      method,
			Controller:  controllerValue,
			RequestType: requestType.Elem(), // 去掉指针类型
		}

		fmt.Printf("Registered route: Name=%s, Method=%s, RequestType=%s\n",
			requestTypeName, method.Name, requestTypeName)
	}

	return nil
}

// validateMethodSignature 验证方法签名
func (r *ReflectionRouter) validateMethodSignature(method reflect.Method) error {
	methodType := method.Type

	// 检查参数数量: receiver + request = 2
	if methodType.NumIn() != 2 {
		return fmt.Errorf("method should have exactly 1 parameter (plus receiver), got %d", methodType.NumIn()-1)
	}

	// 检查返回值数量: 应该返回1个值
	if methodType.NumOut() != 1 {
		return fmt.Errorf("method should return exactly 1 value, got %d", methodType.NumOut())
	}

	// 检查请求参数类型: 应该是指向pb.Request_XXX的指针
	requestType := methodType.In(1)
	if requestType.Kind() != reflect.Ptr {
		return fmt.Errorf("request parameter should be a pointer")
	}

	// 检查返回值类型: 应该实现proto.Message接口
	returnType := methodType.Out(0)
	protoMessageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	if !returnType.Implements(protoMessageType) {
		return fmt.Errorf("return type should implement proto.Message interface")
	}

	return nil
}

// getRequestTypeName 从请求类型获取消息名称
func (r *ReflectionRouter) getRequestTypeName(requestType reflect.Type) string {
	// 去掉指针，获取实际类型名称
	typeName := requestType.Elem().Name()

	// 例如: Request_SearchBook -> Request_SearchBook
	return typeName
}

// Dispatch 根据协议ID分发请求到对应的方法
func (r *ReflectionRouter) Dispatch(pid uint32, data []byte) (proto.Message, uint32, error) {
	req, err := pb.UnmarshalRequest(pid, data)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal request failed: %w", err)
	}

	// 查找对应的方法信息
	methodInfo, ok := r.methodMap[pid]
	if !ok {
		return nil, 0, fmt.Errorf("no handler found for protocol ID: %d", pid)
	}

	// 调用处理方法
	args := []reflect.Value{methodInfo.Controller, reflect.ValueOf(req)}
	results := methodInfo.Method.Func.Call(args)

	// 获取返回结果
	if len(results) != 1 {
		return nil, 0, fmt.Errorf("method returned unexpected number of values: %d", len(results))
	}

	response := results[0].Interface().(proto.Message)

	responsePid := pb.GetResponsePID(response)
	return response, responsePid, nil
}
