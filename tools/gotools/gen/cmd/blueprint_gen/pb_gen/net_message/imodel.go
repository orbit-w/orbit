package net_message

type NetWallFile interface {
	GetPackageName() string
	GetRequests() []*NetMessage
	GetNotifies() []*NetMessage
}
