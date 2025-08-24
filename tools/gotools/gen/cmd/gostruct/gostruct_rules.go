package gostruct

import (
	"fmt"
	"strings"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// validateFieldByRole 对单个字段进行基于角色的约束校验。
// - Entity：允许标量 id；消息类型字段必须为 Component；map 的 value 若为消息类型则必须为 Component
// - Component：禁止引用 Entity（无论是普通消息字段还是 map 的 value）
// - Data：禁止引用 Component/Entity（无论是普通消息字段还是 map 的 value）
func validateFieldByRole(
	ctx ProtoContext,
	role messageRole,
	fq string,
	f *gogodesc.FieldDescriptorProto,
	roleByFQName map[string]messageRole,
	nameByFQName map[string]string,
) error {
	// Entity 允许非消息类型的 id 字段
	if role == roleEntity && f.GetName() == "id" && f.GetType() != gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
		return nil
	}

	// Map 字段：当 value 为消息类型时，按角色规则进行限制
	if isMapField(f) {
		_, valueField := getMapKeyValueTypes(ctx, f)
		if valueField != nil && valueField.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
			tname := trimLeadingDot(valueField.GetTypeName())
			vRole := roleByFQName[tname]
			switch role {
			case roleEntity:
				if vRole != roleComponent {
					return fmt.Errorf("entity %s: map value type %s must be Component", fq, nameByFQName[tname])
				}
			case roleComponent:
				if vRole == roleEntity {
					return fmt.Errorf("component %s: map value type must not reference Entity (%s)", fq, nameByFQName[tname])
				}
			case roleData:
				if vRole == roleComponent || vRole == roleEntity {
					return fmt.Errorf("data %s: map value type must not reference Component/Entity (%s)", fq, nameByFQName[tname])
				}
			}
		}
		return nil
	}

	// 非 map 的消息类型字段
	if f.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
		tname := trimLeadingDot(f.GetTypeName())
		fRole := roleByFQName[tname]
		switch role {
		case roleEntity:
			if fRole != roleComponent {
				return fmt.Errorf("entity %s: field %s must be Component, got %s", fq, f.GetName(), nameByFQName[tname])
			}
		case roleComponent:
			if fRole == roleEntity {
				return fmt.Errorf("component %s: field %s must not reference Entity (%s)", fq, f.GetName(), nameByFQName[tname])
			}
		case roleData:
			if fRole == roleComponent || fRole == roleEntity {
				return fmt.Errorf("data %s: field %s must not reference Component/Entity (%s)", fq, f.GetName(), nameByFQName[tname])
			}
		}
	}
	return nil
}

type messageRole int

const (
	roleUnknown messageRole = iota
	roleEntity
	roleComponent
	roleData
)

func classifyRoleByName(name string) messageRole {
	switch {
	case strings.HasSuffix(name, "Entity"):
		return roleEntity
	case strings.HasSuffix(name, "Component"):
		return roleComponent
	case strings.HasSuffix(name, "Data"):
		return roleData
	default:
		return roleUnknown
	}
}

// ValidateMessageRoles enforces the rules:
// 1) Entity: collection of Components; allow optional scalar field named "id"; others must be Components
// 2) Component: supports incremental changes; must NOT embed/point to Entity
// 3) Data: plain data; must NOT embed/point to Component or Entity
func ValidateMessageRoles(ctx ProtoContext) error {
	fds := ctx.GetFileDescriptorSet()
	if fds == nil {
		return fmt.Errorf("descriptor set is nil")
	}

	// Build message index for map entry resolution and a role map of fully qualified names
	buildMessageIndex(ctx)

	roleByFQName := make(map[string]messageRole)
	nameByFQName := make(map[string]string)
	for _, file := range fds.File {
		pkg := file.GetPackage()
		for _, m := range file.GetMessageType() {
			fq := m.GetName()
			if pkg != "" {
				fq = pkg + "." + fq
			}
			roleByFQName[fq] = classifyRoleByName(m.GetName())
			nameByFQName[fq] = m.GetName()
		}
	}

	var firstErr error

	for _, file := range fds.File {
		if isSystemProtoFile(file.GetName()) {
			continue
		}
		pkg := file.GetPackage()
		for _, m := range file.GetMessageType() {
			fq := m.GetName()
			if pkg != "" {
				fq = pkg + "." + fq
			}
			role := roleByFQName[fq]
			if role == roleUnknown {
				if firstErr == nil {
					firstErr = fmt.Errorf("message %s: must be suffixed with one of [Entity|Component|Data]", fq)
				}
				continue
			}

			// Validate fields based on role
			for _, f := range m.GetField() {
				if err := validateFieldByRole(ctx, role, fq, f, roleByFQName, nameByFQName); err != nil && firstErr == nil {
					firstErr = err
				}
			}
		}
	}

	return firstErr
}
