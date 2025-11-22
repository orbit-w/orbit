package blueprint_gen

import (
	"fmt"
	"strings"
)

func (g *GoStructGenerator) GenerateImport(obj MMEObjectBase, packageName string) string {
	sb := strings.Builder{}

	// 文件头部
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	sb.WriteString("import (\n")

	sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
	sb.WriteString("\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n")
	sb.WriteString("\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n")
	sb.WriteString("\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n")
	sb.WriteString("\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n")
	sb.WriteString("\t\"google.golang.org/protobuf/proto\"\n")

	if obj.GetObjectType() == ObjectTypeEntity {
		sb.WriteString("\t\"go.mongodb.org/mongo-driver/v2/bson\"\n")
	}

	var (
		mapsImport bool // 是否已经导入了maps包
		xmapImport bool // 是否已经导入了xmap包
	)

	// 检查xmap的Value类型是 MMEObject 或 Message, 需要使用xmapwrapper
	if ok, _ := hasXMapValueIsMMEObjectOrMessage(obj.GetFields()); ok {
		sb.WriteString("\txmapwrapper \"gitee.com/orbit-w/orbit/lib/module/xmapwrapper\"\n")
		if !xmapImport {
			xmapImport = true
			sb.WriteString("\t\"gitee.com/orbit-w/meteor/bases/container/xmap\"\n")
		}

	}

	//TODO: 检查map的Value类型是 MMEObject 或 Message, 需要特殊处理
	if ok, field := hasMapValueIsMMEObjectOrMessage(obj.GetFields()); ok {
		panic(fmt.Sprintf("map的Value类型是 MMEObject 或 Message, 需要特殊处理: %+v", field.Name))
	}

	// 检查map的Value类型是基础类型, 则需要导入maps包
	if ok, _ := hasMapValueIsBaseType(obj.GetFields()); ok {
		mapsImport = true
		sb.WriteString("\t\"maps\"\n")
	}

	// 检查xmap的Value类型是基础类型, 则需要导入maps包和xmap包
	if ok, _ := hasXMapValueIsBaseType(obj.GetFields()); ok {
		if !mapsImport {
			mapsImport = true
			sb.WriteString("\t\"maps\"\n")
		}
		if !xmapImport {
			xmapImport = true
			sb.WriteString("\t\"gitee.com/orbit-w/meteor/bases/container/xmap\"\n")
		}
	}
	sb.WriteString(")\n\n")

	return sb.String()
}
