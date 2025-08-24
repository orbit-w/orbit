package gostruct

import (
	"fmt"
	"strings"
)

// generateComponentDirtyBits emits bit constants for each field in a Component.
// To avoid collisions across multiple components in one file, constant names
// are prefixed with the struct name: <StructName>Dirty<FieldName>Bit = 1 << n
func generateComponentDirtyBits(s *GoStruct) string {
	if s == nil || len(s.Fields) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("// Dirty bits for component fields\n")
	b.WriteString("const (\n")
	bitIndex := 0
	for _, f := range s.Fields {
		// Skip embedded trackers
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		b.WriteString(fmt.Sprintf("\t%sDirty%sBit int64 = 1 << %d\n", s.Name, f.Name, bitIndex))
		bitIndex++
	}
	b.WriteString(")\n")
	return b.String()
}

// generateComponentAccessors emits getter/setter methods for each field.
// Setters call MarkDirty with the corresponding bit.
func generateComponentAccessors(s *GoStruct) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		// If field is a map, we skip generating direct Get/Set to encourage using OpsV2
		isMapField := f.IsMapType()

		// Getter (for nested component fields, support lazy init + link)
		if isComponentType(f.Type) {
			// Pointer expected for message fields, and method name uses Get<FieldName>Component
			b.WriteString(fmt.Sprintf("func (c *%s) Get%sComponent() %s {\n", s.Name, f.Name, f.Type))
			b.WriteString(fmt.Sprintf("\tif c.%s == nil {\n", f.Name))
			b.WriteString(fmt.Sprintf("\t\tc.%s = &%s{}\n", f.Name, extractTypeName(f.Type)))
			b.WriteString(fmt.Sprintf("\t\tc.%s.Link(&c.DirtyTracker, %sDirty%sBit)\n", f.Name, s.Name, f.Name))
			b.WriteString("\t}\n")
			b.WriteString(fmt.Sprintf("\treturn c.%s\n}\n\n", f.Name))
		} else if !isMapField {
			b.WriteString(fmt.Sprintf("func (c *%s) Get%s() %s {\n\treturn c.%s\n}\n\n", s.Name, f.Name, f.Type, f.Name))
		}
		// Setter
		if !isMapField {
			b.WriteString(fmt.Sprintf("func (c *%s) Set%s(v %s) {\n\tc.%s = v\n", s.Name, f.Name, f.Type, f.Name))
			// If nested component, auto-link child's tracker to parent
			if isComponentType(f.Type) {
				b.WriteString(fmt.Sprintf("\tif v != nil { v.Link(&c.DirtyTracker, %sDirty%sBit) }\n", s.Name, f.Name))
			}
			b.WriteString(fmt.Sprintf("\tc.MarkDirty(%sDirty%sBit)\n}\n\n", s.Name, f.Name))
		}

		// Extra helpers for map fields: generate OpsV2 accessor only
		if isMapField {
			if keyType, valType, ok := splitMapGoType(f.Type); ok {
				// V2: Generic MapAccessor-based accessor (avoids code duplication)
				b.WriteString(fmt.Sprintf("type %s%sOpsV2 struct { xmap.MapAccessor[%s, %s] }\n\n", s.Name, f.Name, keyType, valType))
				b.WriteString(fmt.Sprintf("func (c *%s) %sOpsV2() %s%sOpsV2 {\n", s.Name, f.Name, s.Name, f.Name))
				b.WriteString(fmt.Sprintf("\treturn %s%sOpsV2{ xmap.NewMapAccessorWithMarker(&c.%s, c, %sDirty%sBit) }\n}\n\n", s.Name, f.Name, f.Name, s.Name, f.Name))
			}
		}

		// If this field itself is a nested Component, generate Link/Unlink helpers
		if isComponentType(f.Type) {
			// Link: c.<Field>.Link(&c.DirtyTracker, <StructName>Dirty<Field>Bit)
			b.WriteString(fmt.Sprintf("func (c *%s) Link%s() {\n", s.Name, f.Name))
			// ensure non-nil when pointer field
			if strings.HasPrefix(f.Type, "*") {
				b.WriteString(fmt.Sprintf("\tif c.%s == nil { c.%s = &%s{} }\n", f.Name, f.Name, extractTypeName(f.Type)))
			}
			b.WriteString(fmt.Sprintf("\tc.%s.Link(&c.DirtyTracker, %sDirty%sBit)\n}\n\n", f.Name, s.Name, f.Name))

			// Unlink: c.<Field>.Unlink()
			b.WriteString(fmt.Sprintf("func (c *%s) Unlink%s() {\n", s.Name, f.Name))
			b.WriteString(fmt.Sprintf("\tif c.%s != nil { c.%s.Unlink() }\n}\n\n", f.Name, f.Name))
		}
	}

	// After all field-level methods, emit a Relink() method to restore dirty link relationships
	// after deserialization (e.g., from MongoDB). This recursively links child components and
	// component-valued map entries to this component's tracker with the correct dirty bit.
	b.WriteString(fmt.Sprintf("func (c *%s) Relink() {\n", s.Name))
	b.WriteString("\tif c == nil { return }\n")
	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		// Direct child component pointer
		if isComponentType(f.Type) {
			b.WriteString(fmt.Sprintf("\tif c.%s != nil { c.%s.Link(&c.DirtyTracker, %sDirty%sBit); c.%s.Relink() }\n", f.Name, f.Name, s.Name, f.Name, f.Name))
			continue
		}
		// Map of component pointers: map[K]*XComponent
		if f.IsMapType() {
			if _, valType, ok := splitMapGoType(f.Type); ok {
				if isComponentType(valType) {
					b.WriteString(fmt.Sprintf("\tif c.%s != nil { for _, v := range c.%s { if v != nil { v.Link(&c.DirtyTracker, %sDirty%sBit); v.Relink() } } }\n", f.Name, f.Name, s.Name, f.Name))
				}
			}
		}
	}
	b.WriteString("}\n\n")

	return b.String()
}

// generateComponentClearDirtyRecursiveMethod emits a method that clears this component's
// dirty flags and recursively clears dirty flags on any nested component fields,
// including component values inside maps.
func generateComponentClearDirtyRecursiveMethod(s *GoStruct) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("func (c *%s) ClearDirtyRecursive() {\n", s.Name))
	b.WriteString("\tif c == nil { return }\n")
	// Recurse into direct child components and map component values first
	for _, f := range s.Fields {
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		// Direct child component pointer
		if isComponentType(f.Type) {
			b.WriteString(fmt.Sprintf("\tif c.%s != nil { c.%s.ClearDirtyRecursive() }\n", f.Name, f.Name))
			continue
		}
		// Map of component pointers
		if f.IsMapType() {
			if _, valType, ok := splitMapGoType(f.Type); ok {
				if isComponentType(valType) {
					b.WriteString(fmt.Sprintf("\tif c.%s != nil { for _, v := range c.%s { if v != nil { v.ClearDirtyRecursive() } } }\n", f.Name, f.Name))
				}
			}
		}
	}
	// Finally clear this component's own dirty bits
	b.WriteString("\tc.ClearAllDirty()\n")
	b.WriteString("}\n")
	return b.String()
}

// splitMapGoType parses a Go map type string like "map[K]V" and returns K and V.
func splitMapGoType(typeStr string) (keyType, valType string, ok bool) {
	if !strings.HasPrefix(typeStr, "map[") {
		return "", "", false
	}
	rb := strings.Index(typeStr, "]")
	if rb < 0 {
		return "", "", false
	}
	keyType = strings.TrimSpace(typeStr[len("map["):rb])
	valType = strings.TrimSpace(typeStr[rb+1:])
	if keyType == "" || valType == "" {
		return "", "", false
	}
	return keyType, valType, true
}

// generateEntityOpsForwarders generates convenience forwarder methods on Entity
// which expose <Component>.<MapField>Ops() directly on the Entity for quick access.
// Only fields that are component pointers or embedded components will be considered.
func generateEntityOpsForwarders(entity *GoStruct, index map[string]*GoStruct) string {
	if entity == nil {
		return ""
	}
	var b strings.Builder
	for _, f := range entity.Fields {
		// Identify component-typed field: either *XComponent or XComponent
		if !isComponentType(f.Type) {
			continue
		}
		typeName := extractTypeName(f.Type)
		comp, ok := index[typeName]
		if !ok {
			continue
		}
		// For each map field in component, expose a forwarder
		for _, cf := range comp.Fields {
			if cf.IsMapType() {
				// Forwarder: <FieldName><MapName>Ops()
				// Example: InventoryOps() -> returns UserComponentItemsOps via c.Component.ItemsOps()
				b.WriteString(fmt.Sprintf(
					"func (e *%s) %s%sOps() %s%sOps {\n",
					entity.Name, comp.Name, cf.Name, comp.Name, cf.Name,
				))
				// Access path: if field is pointer, nil-guard then init component
				if strings.HasPrefix(f.Type, "*") {
					b.WriteString(fmt.Sprintf(
						"\tif e.%s == nil { e.%s = New%s() }\n",
						f.Name, f.Name, comp.Name,
					))
					b.WriteString(fmt.Sprintf("\treturn e.%s.%sOps()\n}\n\n", f.Name, cf.Name))
				} else {
					b.WriteString(fmt.Sprintf("\treturn e.%s.%sOps()\n}\n\n", f.Name, cf.Name))
				}
			}
		}
	}
	return b.String()
}

// generateEntitySafeComponentGetters creates lazy-init getters on Entity for each Component field.
// Example:
//
//	func (e *PlayerEntity) GetAsset() *AssetComponent {
//	    if e.Asset == nil { e.Asset = NewAssetComponent() }
//	    return e.Asset
//	}
func generateEntitySafeComponentGetters(entity *GoStruct) string {
	if entity == nil {
		return ""
	}
	var b strings.Builder
	for _, f := range entity.Fields {
		if !isComponentType(f.Type) {
			continue
		}
		typeName := extractTypeName(f.Type)
		// Only generate for pointer fields to allow assignment
		if !strings.HasPrefix(f.Type, "*") {
			continue
		}
		// Method name uses the component type (e.g., GetAssetComponent) rather than field name
		b.WriteString(fmt.Sprintf("func (e *%s) Get%s() %s {\n", entity.Name, typeName, f.Type))
		b.WriteString(fmt.Sprintf("\tif e.%s == nil { e.%s = New%s() }\n", f.Name, f.Name, typeName))
		b.WriteString(fmt.Sprintf("\treturn e.%s\n}\n\n", f.Name))
	}
	return b.String()
}

// generateEntityMongoUpdateAggregator emits an Entity-level aggregator that uses
// mgo_builder.MongoUpdateBuilder to collect updates from all Component fields.
// Only generates the method if the entity has component fields.
func generateEntityMongoUpdateAggregator(entity *GoStruct) string {
	if entity == nil {
		return ""
	}

	// Count component fields first
	componentFields := 0
	for _, f := range entity.Fields {
		typeName := strings.TrimPrefix(f.Type, "*")
		if strings.HasSuffix(typeName, "Component") {
			componentFields++
		}
	}

	// Don't generate the method if there are no component fields
	if componentFields == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("func (e *%s) BuildMongoUpdate() map[string]any {\n", entity.Name))
	b.WriteString("\tif e == nil {\n\t\treturn nil\n\t}\n")
	b.WriteString("\treturn mgo_builder.WithBuilderResult(func(builder *mgo_builder.MongoUpdateBuilder) map[string]any {\n")

	for _, f := range entity.Fields {
		typeName := strings.TrimPrefix(f.Type, "*")
		if !strings.HasSuffix(typeName, "Component") {
			continue
		}
		// Delegate to component's BuildMongoUpdate with field proto name as prefix
		b.WriteString(fmt.Sprintf("\t\te.%s.BuildMongoUpdate(builder, \"%s\")\n", f.Name, f.ProtoName))
	}

	b.WriteString("\t\treturn builder.Build()\n\t})\n}\n")
	return b.String()
}

// generateComponentConstructor emits a New<Component>() constructor that:
// - allocates the component
// - recursively allocates nested component fields using their own New constructors
// - Links each child component's tracker to the parent's tracker using the correct dirty bit
func generateComponentConstructor(s *GoStruct) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("func New%s() *%s {\n", s.Name, s.Name))
	b.WriteString(fmt.Sprintf("\tc := &%s{}\n", s.Name))
	// Initialize and link nested component fields
	for _, f := range s.Fields {
		// Skip tracker/embed and non-component fields
		if f.Name == "DirtyTracker" || strings.HasPrefix(f.Type, "dirty.") {
			continue
		}
		if isComponentType(f.Type) {
			typeName := extractTypeName(f.Type)
			// Use child's New constructor to ensure its own subtree is initialized
			b.WriteString(fmt.Sprintf("\tc.%s = New%s()\n", f.Name, typeName))
			// Link child to this component's tracker with the appropriate dirty bit
			b.WriteString(fmt.Sprintf("\tc.%s.Link(&c.DirtyTracker, %sDirty%sBit)\n", f.Name, s.Name, f.Name))
		}
	}
	b.WriteString("\treturn c\n}\n")
	return b.String()
}

// isStructType checks if a type (potentially with pointer prefix) represents a custom struct
func isStructType(typeName string) bool {
	// Remove pointer prefix if present
	cleanType := strings.TrimPrefix(typeName, "*")

	// Check if it's a basic Go type
	basicTypes := map[string]bool{
		"bool": true, "byte": true, "rune": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "string": true,
		"interface{}": true, "any": true,
	}

	if basicTypes[cleanType] {
		return false
	}

	// Check if it's a package-prefixed type (like dirty.DirtyTracker)
	if strings.Contains(cleanType, ".") {
		return false
	}

	// If it doesn't start with uppercase, it's probably not a struct
	if len(cleanType) == 0 || (cleanType[0] < 'A' || cleanType[0] > 'Z') {
		return false
	}

	return true
}
