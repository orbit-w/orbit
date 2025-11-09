package servicezone

type ZoneType int32

const (
	ZoneTypeDungeon ZoneType = iota
	ZoneTypeCity
	ZoneTypeWild
	ZoneTypeArena
)

type ServiceZone struct {
	ID   string
	Type ZoneType
}
