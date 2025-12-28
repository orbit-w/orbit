package zone_meta

import (
	"time"

	"gitee.com/orbit-w/orbit/core/cluster"
)

const (
	NodeRefreshTime = 15 * time.Minute
)

func (x *ZoneMeta) SetNode(node *cluster.Node) {
	x.node = node
	x.nodeRefreshTime = time.Now().Unix()
}

func (x *ZoneMeta) GetNode() *cluster.Node {
	return x.node
}

func (x *ZoneMeta) IsNodeExpired() bool {
	return time.Now().Unix()-x.nodeRefreshTime > int64(NodeRefreshTime.Seconds())
}

func (x *ZoneMeta) NodeInvalid() bool {
	return x.node == nil
}
