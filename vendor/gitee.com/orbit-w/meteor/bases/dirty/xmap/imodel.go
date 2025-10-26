package xmap

// FactsAccessor 是用于跟踪 Facts 变化的接口
type FactsAccessor interface {
	TrackSet()
}
