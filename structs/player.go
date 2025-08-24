// Code generated from proto files. DO NOT EDIT.

package pb

import (
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	dirty "gitee.com/orbit-w/orbit/lib/module/dirty/dirty_tracker"
	"gitee.com/orbit-w/orbit/lib/module/dirty/xmap"
)

// PlayerEntity represents the proto message PlayerEntity
type PlayerEntity struct {
	Id        int64               `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Asset     *AssetComponent     `json:"asset" bson:"asset" protobuf:"bytes,2,opt,name=asset"`             // proto: AssetComponent
	Item      *ItemComponent      `json:"item" bson:"item" protobuf:"bytes,3,opt,name=item"`                // proto: ItemComponent
	Inventory *InventoryComponent `json:"inventory" bson:"inventory" protobuf:"bytes,4,opt,name=inventory"` // proto: InventoryComponent
}

func (e *PlayerEntity) GetAssetComponent() *AssetComponent {
	if e.Asset == nil {
		e.Asset = NewAssetComponent()
	}
	return e.Asset
}

func (e *PlayerEntity) GetItemComponent() *ItemComponent {
	if e.Item == nil {
		e.Item = NewItemComponent()
	}
	return e.Item
}

func (e *PlayerEntity) GetInventoryComponent() *InventoryComponent {
	if e.Inventory == nil {
		e.Inventory = NewInventoryComponent()
	}
	return e.Inventory
}

func (e *PlayerEntity) BuildMongoUpdate() map[string]any {
	if e == nil {
		return nil
	}
	return mgo_builder.WithBuilderResult(func(builder *mgo_builder.MongoUpdateBuilder) map[string]any {
		e.Asset.BuildMongoUpdate(builder, "asset")
		e.Item.BuildMongoUpdate(builder, "item")
		e.Inventory.BuildMongoUpdate(builder, "inventory")
		return builder.Build()
	})
}

// DeepCopy creates a deep copy of PlayerEntity
func (s *PlayerEntity) DeepCopy() *PlayerEntity {
	if s == nil {
		return nil
	}

	co := &PlayerEntity{}
	*co = *s

	if s.Asset != nil {
		co.Asset = &AssetComponent{}
		s.Asset.DeepCopy(co.Asset)
	}

	if s.Item != nil {
		co.Item = &ItemComponent{}
		s.Item.DeepCopy(co.Item)
	}

	if s.Inventory != nil {
		co.Inventory = &InventoryComponent{}
		s.Inventory.DeepCopy(co.Inventory)
	}

	return co
}

// ItemComponent represents the proto message ItemComponent
type ItemComponent struct {
	dirty.DirtyTracker
	Id    int64               `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Items map[int64]*ItemData `json:"items" bson:"items" protobuf:"bytes,2,opt,name=items"` // proto: map<int64, ItemData>
}

func NewItemComponent() *ItemComponent {
	c := &ItemComponent{}
	return c
}

// Dirty bits for component fields
const (
	ItemComponentDirtyIdBit    int64 = 1 << 0
	ItemComponentDirtyItemsBit int64 = 1 << 1
)

func (c *ItemComponent) GetId() int64 {
	return c.Id
}

func (c *ItemComponent) SetId(v int64) {
	c.Id = v
	c.MarkDirty(ItemComponentDirtyIdBit)
}

type ItemComponentItemsOpsV2 struct {
	xmap.MapAccessor[int64, *ItemData]
}

func (c *ItemComponent) ItemsOpsV2() ItemComponentItemsOpsV2 {
	return ItemComponentItemsOpsV2{xmap.NewMapAccessorWithMarker(&c.Items, c, ItemComponentDirtyItemsBit)}
}

func (c *ItemComponent) Relink() {
	if c == nil {
		return
	}
}

func (c *ItemComponent) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if c == nil {
		return
	}
	{
		path := "id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(ItemComponentDirtyIdBit) {
			builder.Set(path, c.Id)
		}
	}
	{
		path := "items"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(ItemComponentDirtyItemsBit) {
			// Deep copy map to prevent reference sharing
			if c.Items != nil {
				dst := make(map[int64]*ItemData, len(c.Items))
				for k := range c.Items {
					v := c.Items[k]
					if v != nil {
						dst[k] = new(ItemData)
						v.DeepCopy(dst[k])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
}

func (c *ItemComponent) ClearDirtyRecursive() {
	if c == nil {
		return
	}
	c.ClearAllDirty()
}

// DeepCopy creates a deep copy of ItemComponent
func (s *ItemComponent) DeepCopy(co *ItemComponent) {
	if s == nil {
		return
	}

	*co = *s

	if s.Items != nil {
		co.Items = make(map[int64]*ItemData, len(s.Items))
		for k, v := range s.Items {
			if v != nil {
				co.Items[k] = new(ItemData)
				v.DeepCopy(co.Items[k])
			}
		}
	}
}

// AssetComponent represents the proto message AssetComponent
type AssetComponent struct {
	dirty.DirtyTracker
	Id      int64                `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Assets  map[int32]*AssetData `json:"assets" bson:"assets" protobuf:"bytes,2,opt,name=assets"`    // proto: map<int32, AssetData>
	Records map[int32]int32      `json:"records" bson:"records" protobuf:"bytes,3,opt,name=records"` // proto: map<int32, int32>
}

func NewAssetComponent() *AssetComponent {
	c := &AssetComponent{}
	return c
}

// Dirty bits for component fields
const (
	AssetComponentDirtyIdBit      int64 = 1 << 0
	AssetComponentDirtyAssetsBit  int64 = 1 << 1
	AssetComponentDirtyRecordsBit int64 = 1 << 2
)

func (c *AssetComponent) GetId() int64 {
	return c.Id
}

func (c *AssetComponent) SetId(v int64) {
	c.Id = v
	c.MarkDirty(AssetComponentDirtyIdBit)
}

type AssetComponentAssetsOpsV2 struct {
	xmap.MapAccessor[int32, *AssetData]
}

func (c *AssetComponent) AssetsOpsV2() AssetComponentAssetsOpsV2 {
	return AssetComponentAssetsOpsV2{xmap.NewMapAccessorWithMarker(&c.Assets, c, AssetComponentDirtyAssetsBit)}
}

type AssetComponentRecordsOpsV2 struct{ xmap.MapAccessor[int32, int32] }

func (c *AssetComponent) RecordsOpsV2() AssetComponentRecordsOpsV2 {
	return AssetComponentRecordsOpsV2{xmap.NewMapAccessorWithMarker(&c.Records, c, AssetComponentDirtyRecordsBit)}
}

func (c *AssetComponent) Relink() {
	if c == nil {
		return
	}
}

func (c *AssetComponent) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if c == nil {
		return
	}
	{
		path := "id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(AssetComponentDirtyIdBit) {
			builder.Set(path, c.Id)
		}
	}
	{
		path := "assets"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(AssetComponentDirtyAssetsBit) {
			// Deep copy map to prevent reference sharing
			if c.Assets != nil {
				dst := make(map[int32]*AssetData, len(c.Assets))
				for k := range c.Assets {
					v := c.Assets[k]
					if v != nil {
						dst[k] = new(AssetData)
						v.DeepCopy(dst[k])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
	{
		path := "records"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(AssetComponentDirtyRecordsBit) {
			// Deep copy map to prevent reference sharing
			if c.Records != nil {
				dst := make(map[int32]int32, len(c.Records))
				for k := range c.Records {
					dst[k] = c.Records[k]
				}
				builder.Set(path, dst)
			}
		}
	}
}

func (c *AssetComponent) ClearDirtyRecursive() {
	if c == nil {
		return
	}
	c.ClearAllDirty()
}

// DeepCopy creates a deep copy of AssetComponent
func (s *AssetComponent) DeepCopy(co *AssetComponent) {
	if s == nil {
		return
	}

	*co = *s

	if s.Assets != nil {
		co.Assets = make(map[int32]*AssetData, len(s.Assets))
		for k, v := range s.Assets {
			if v != nil {
				co.Assets[k] = new(AssetData)
				v.DeepCopy(co.Assets[k])
			}
		}
	}
	if s.Records != nil {
		co.Records = make(map[int32]int32, len(s.Records))
		for k, v := range s.Records {
			co.Records[k] = v
		}
	}
}

// BagComponent represents the proto message BagComponent
type BagComponent struct {
	dirty.DirtyTracker
	Id    int64                `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Items map[string]*ItemData `json:"items" bson:"items" protobuf:"bytes,2,opt,name=items"` // proto: map<string, ItemData>
}

func NewBagComponent() *BagComponent {
	c := &BagComponent{}
	return c
}

// Dirty bits for component fields
const (
	BagComponentDirtyIdBit    int64 = 1 << 0
	BagComponentDirtyItemsBit int64 = 1 << 1
)

func (c *BagComponent) GetId() int64 {
	return c.Id
}

func (c *BagComponent) SetId(v int64) {
	c.Id = v
	c.MarkDirty(BagComponentDirtyIdBit)
}

type BagComponentItemsOpsV2 struct {
	xmap.MapAccessor[string, *ItemData]
}

func (c *BagComponent) ItemsOpsV2() BagComponentItemsOpsV2 {
	return BagComponentItemsOpsV2{xmap.NewMapAccessorWithMarker(&c.Items, c, BagComponentDirtyItemsBit)}
}

func (c *BagComponent) Relink() {
	if c == nil {
		return
	}
}

func (c *BagComponent) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if c == nil {
		return
	}
	{
		path := "id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(BagComponentDirtyIdBit) {
			builder.Set(path, c.Id)
		}
	}
	{
		path := "items"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(BagComponentDirtyItemsBit) {
			// Deep copy map to prevent reference sharing
			if c.Items != nil {
				dst := make(map[string]*ItemData, len(c.Items))
				for k := range c.Items {
					v := c.Items[k]
					if v != nil {
						dst[k] = new(ItemData)
						v.DeepCopy(dst[k])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
}

func (c *BagComponent) ClearDirtyRecursive() {
	if c == nil {
		return
	}
	c.ClearAllDirty()
}

// DeepCopy creates a deep copy of BagComponent
func (s *BagComponent) DeepCopy(co *BagComponent) {
	if s == nil {
		return
	}

	*co = *s

	if s.Items != nil {
		co.Items = make(map[string]*ItemData, len(s.Items))
		for k, v := range s.Items {
			if v != nil {
				co.Items[k] = new(ItemData)
				v.DeepCopy(co.Items[k])
			}
		}
	}
}

// EquipmentComponent represents the proto message EquipmentComponent
type EquipmentComponent struct {
	dirty.DirtyTracker
	Id         int64                    `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Slots      map[int32]*EquipmentData `json:"slots" bson:"slots" protobuf:"bytes,2,opt,name=slots"`                                    // proto: map<int32, EquipmentData>
	Equipments []*EquipmentData         `json:"equipments,omitempty" bson:"equipments,omitempty" protobuf:"bytes,3,rep,name=equipments"` // proto: repeated EquipmentData
	IdList     []int32                  `json:"idList,omitempty" bson:"idList,omitempty" protobuf:"varint,4,rep,name=idList"`            // proto: repeated int32
	Cache      map[int32]string         `json:"cache" bson:"cache" protobuf:"bytes,5,opt,name=cache"`                                    // proto: map<int32, string>
}

func NewEquipmentComponent() *EquipmentComponent {
	c := &EquipmentComponent{}
	return c
}

// Dirty bits for component fields
const (
	EquipmentComponentDirtyIdBit         int64 = 1 << 0
	EquipmentComponentDirtySlotsBit      int64 = 1 << 1
	EquipmentComponentDirtyEquipmentsBit int64 = 1 << 2
	EquipmentComponentDirtyIdListBit     int64 = 1 << 3
	EquipmentComponentDirtyCacheBit      int64 = 1 << 4
)

func (c *EquipmentComponent) GetId() int64 {
	return c.Id
}

func (c *EquipmentComponent) SetId(v int64) {
	c.Id = v
	c.MarkDirty(EquipmentComponentDirtyIdBit)
}

type EquipmentComponentSlotsOpsV2 struct {
	xmap.MapAccessor[int32, *EquipmentData]
}

func (c *EquipmentComponent) SlotsOpsV2() EquipmentComponentSlotsOpsV2 {
	return EquipmentComponentSlotsOpsV2{xmap.NewMapAccessorWithMarker(&c.Slots, c, EquipmentComponentDirtySlotsBit)}
}

func (c *EquipmentComponent) GetEquipments() []*EquipmentData {
	return c.Equipments
}

func (c *EquipmentComponent) SetEquipments(v []*EquipmentData) {
	c.Equipments = v
	c.MarkDirty(EquipmentComponentDirtyEquipmentsBit)
}

func (c *EquipmentComponent) GetIdList() []int32 {
	return c.IdList
}

func (c *EquipmentComponent) SetIdList(v []int32) {
	c.IdList = v
	c.MarkDirty(EquipmentComponentDirtyIdListBit)
}

type EquipmentComponentCacheOpsV2 struct {
	xmap.MapAccessor[int32, string]
}

func (c *EquipmentComponent) CacheOpsV2() EquipmentComponentCacheOpsV2 {
	return EquipmentComponentCacheOpsV2{xmap.NewMapAccessorWithMarker(&c.Cache, c, EquipmentComponentDirtyCacheBit)}
}

func (c *EquipmentComponent) Relink() {
	if c == nil {
		return
	}
}

func (c *EquipmentComponent) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if c == nil {
		return
	}
	{
		path := "id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(EquipmentComponentDirtyIdBit) {
			builder.Set(path, c.Id)
		}
	}
	{
		path := "slots"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(EquipmentComponentDirtySlotsBit) {
			// Deep copy map to prevent reference sharing
			if c.Slots != nil {
				dst := make(map[int32]*EquipmentData, len(c.Slots))
				for k := range c.Slots {
					v := c.Slots[k]
					if v != nil {
						dst[k] = new(EquipmentData)
						v.DeepCopy(dst[k])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
	{
		path := "equipments"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(EquipmentComponentDirtyEquipmentsBit) {
			// Deep copy slice to prevent reference sharing
			if c.Equipments != nil {
				dst := make([]*EquipmentData, len(c.Equipments))
				for i := range c.Equipments {
					v := c.Equipments[i]
					if v != nil {
						dst[i] = new(EquipmentData)
						v.DeepCopy(dst[i])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
	{
		path := "idList"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(EquipmentComponentDirtyIdListBit) {
			// Deep copy slice to prevent reference sharing
			if c.IdList != nil {
				dst := make([]int32, len(c.IdList))
				copy(dst, c.IdList)
				builder.Set(path, dst)
			}
		}
	}
	{
		path := "cache"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(EquipmentComponentDirtyCacheBit) {
			// Deep copy map to prevent reference sharing
			if c.Cache != nil {
				dst := make(map[int32]string, len(c.Cache))
				for k := range c.Cache {
					dst[k] = c.Cache[k]
				}
				builder.Set(path, dst)
			}
		}
	}
}

func (c *EquipmentComponent) ClearDirtyRecursive() {
	if c == nil {
		return
	}
	c.ClearAllDirty()
}

// DeepCopy creates a deep copy of EquipmentComponent
func (s *EquipmentComponent) DeepCopy(co *EquipmentComponent) {
	if s == nil {
		return
	}

	*co = *s

	if s.Slots != nil {
		co.Slots = make(map[int32]*EquipmentData, len(s.Slots))
		for k, v := range s.Slots {
			if v != nil {
				co.Slots[k] = new(EquipmentData)
				v.DeepCopy(co.Slots[k])
			}
		}
	}
	if s.Equipments != nil {
		co.Equipments = make([]*EquipmentData, len(s.Equipments))
		for i, v := range s.Equipments {
			if v != nil {
				co.Equipments[i] = new(EquipmentData)
				v.DeepCopy(co.Equipments[i])
			}
		}
	}
	if s.IdList != nil {
		co.IdList = make([]int32, len(s.IdList))
		copy(co.IdList, s.IdList)
	}
	if s.Cache != nil {
		co.Cache = make(map[int32]string, len(s.Cache))
		for k, v := range s.Cache {
			co.Cache[k] = v
		}
	}
}

// InventoryComponent represents the proto message InventoryComponent
type InventoryComponent struct {
	dirty.DirtyTracker
	Id        int64                `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Bag       *BagComponent        `json:"bag" bson:"bag" protobuf:"bytes,2,opt,name=bag"`                   // proto: BagComponent
	Equipment *EquipmentComponent  `json:"equipment" bson:"equipment" protobuf:"bytes,3,opt,name=equipment"` // proto: EquipmentComponent
	Stash     map[string]*ItemData `json:"stash" bson:"stash" protobuf:"bytes,4,opt,name=stash"`             // proto: map<string, ItemData>
}

func NewInventoryComponent() *InventoryComponent {
	c := &InventoryComponent{}
	c.Bag = NewBagComponent()
	c.Bag.Link(&c.DirtyTracker, InventoryComponentDirtyBagBit)
	c.Equipment = NewEquipmentComponent()
	c.Equipment.Link(&c.DirtyTracker, InventoryComponentDirtyEquipmentBit)
	return c
}

// Dirty bits for component fields
const (
	InventoryComponentDirtyIdBit        int64 = 1 << 0
	InventoryComponentDirtyBagBit       int64 = 1 << 1
	InventoryComponentDirtyEquipmentBit int64 = 1 << 2
	InventoryComponentDirtyStashBit     int64 = 1 << 3
)

func (c *InventoryComponent) GetId() int64 {
	return c.Id
}

func (c *InventoryComponent) SetId(v int64) {
	c.Id = v
	c.MarkDirty(InventoryComponentDirtyIdBit)
}

func (c *InventoryComponent) GetBagComponent() *BagComponent {
	if c.Bag == nil {
		c.Bag = &BagComponent{}
		c.Bag.Link(&c.DirtyTracker, InventoryComponentDirtyBagBit)
	}
	return c.Bag
}

func (c *InventoryComponent) SetBag(v *BagComponent) {
	c.Bag = v
	if v != nil {
		v.Link(&c.DirtyTracker, InventoryComponentDirtyBagBit)
	}
	c.MarkDirty(InventoryComponentDirtyBagBit)
}

func (c *InventoryComponent) LinkBag() {
	if c.Bag == nil {
		c.Bag = &BagComponent{}
	}
	c.Bag.Link(&c.DirtyTracker, InventoryComponentDirtyBagBit)
}

func (c *InventoryComponent) UnlinkBag() {
	if c.Bag != nil {
		c.Bag.Unlink()
	}
}

func (c *InventoryComponent) GetEquipmentComponent() *EquipmentComponent {
	if c.Equipment == nil {
		c.Equipment = &EquipmentComponent{}
		c.Equipment.Link(&c.DirtyTracker, InventoryComponentDirtyEquipmentBit)
	}
	return c.Equipment
}

func (c *InventoryComponent) SetEquipment(v *EquipmentComponent) {
	c.Equipment = v
	if v != nil {
		v.Link(&c.DirtyTracker, InventoryComponentDirtyEquipmentBit)
	}
	c.MarkDirty(InventoryComponentDirtyEquipmentBit)
}

func (c *InventoryComponent) LinkEquipment() {
	if c.Equipment == nil {
		c.Equipment = &EquipmentComponent{}
	}
	c.Equipment.Link(&c.DirtyTracker, InventoryComponentDirtyEquipmentBit)
}

func (c *InventoryComponent) UnlinkEquipment() {
	if c.Equipment != nil {
		c.Equipment.Unlink()
	}
}

type InventoryComponentStashOpsV2 struct {
	xmap.MapAccessor[string, *ItemData]
}

func (c *InventoryComponent) StashOpsV2() InventoryComponentStashOpsV2 {
	return InventoryComponentStashOpsV2{xmap.NewMapAccessorWithMarker(&c.Stash, c, InventoryComponentDirtyStashBit)}
}

func (c *InventoryComponent) Relink() {
	if c == nil {
		return
	}
	if c.Bag != nil {
		c.Bag.Link(&c.DirtyTracker, InventoryComponentDirtyBagBit)
		c.Bag.Relink()
	}
	if c.Equipment != nil {
		c.Equipment.Link(&c.DirtyTracker, InventoryComponentDirtyEquipmentBit)
		c.Equipment.Relink()
	}
}

func (c *InventoryComponent) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if c == nil {
		return
	}
	{
		path := "id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(InventoryComponentDirtyIdBit) {
			builder.Set(path, c.Id)
		}
	}
	{
		path := "bag"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(InventoryComponentDirtyBagBit) {
			if c.Bag == nil {
				builder.Set(path, nil)
			} else {
				c.Bag.BuildMongoUpdate(builder, path)
			}
		}
	}
	{
		path := "equipment"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(InventoryComponentDirtyEquipmentBit) {
			if c.Equipment == nil {
				builder.Set(path, nil)
			} else {
				c.Equipment.BuildMongoUpdate(builder, path)
			}
		}
	}
	{
		path := "stash"
		if prefix != "" {
			path = prefix + "." + path
		}
		if c.IsDirty(InventoryComponentDirtyStashBit) {
			// Deep copy map to prevent reference sharing
			if c.Stash != nil {
				dst := make(map[string]*ItemData, len(c.Stash))
				for k := range c.Stash {
					v := c.Stash[k]
					if v != nil {
						dst[k] = new(ItemData)
						v.DeepCopy(dst[k])
					}
				}
				builder.Set(path, dst)
			}
		}
	}
}

func (c *InventoryComponent) ClearDirtyRecursive() {
	if c == nil {
		return
	}
	if c.Bag != nil {
		c.Bag.ClearDirtyRecursive()
	}
	if c.Equipment != nil {
		c.Equipment.ClearDirtyRecursive()
	}
	c.ClearAllDirty()
}

// DeepCopy creates a deep copy of InventoryComponent
func (s *InventoryComponent) DeepCopy(co *InventoryComponent) {
	if s == nil {
		return
	}

	*co = *s

	if s.Bag != nil {
		co.Bag = &BagComponent{}
		s.Bag.DeepCopy(co.Bag)
	}
	if s.Equipment != nil {
		co.Equipment = &EquipmentComponent{}
		s.Equipment.DeepCopy(co.Equipment)
	}
	if s.Stash != nil {
		co.Stash = make(map[string]*ItemData, len(s.Stash))
		for k, v := range s.Stash {
			if v != nil {
				co.Stash[k] = new(ItemData)
				v.DeepCopy(co.Stash[k])
			}
		}
	}
}

// AssetData represents the proto message AssetData
type AssetData struct {
	Id     int64 `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	ConfId int32 `json:"conf_id" bson:"conf_id" protobuf:"varint,2,opt,name=conf_id"`
}

func (s *AssetData) DeepCopy(co *AssetData) {
	if s == nil {
		return
	}
	*co = *s
}

// ItemData represents the proto message ItemData
type ItemData struct {
	Id       int64 `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Type     int32 `json:"type" bson:"type" protobuf:"varint,2,opt,name=type"`
	Quality  int32 `json:"quality" bson:"quality" protobuf:"varint,3,opt,name=quality"`
	Quantity int32 `json:"quantity" bson:"quantity" protobuf:"varint,4,opt,name=quantity"`
}

func (s *ItemData) DeepCopy(co *ItemData) {
	if s == nil {
		return
	}
	*co = *s
}

// EquipmentData represents the proto message EquipmentData
type EquipmentData struct {
	Id   int64  `json:"id" bson:"id" protobuf:"varint,1,opt,name=id"`
	Slot string `json:"slot" bson:"slot" protobuf:"bytes,2,opt,name=slot"`
}

func (s *EquipmentData) DeepCopy(co *EquipmentData) {
	if s == nil {
		return
	}
	*co = *s
}
