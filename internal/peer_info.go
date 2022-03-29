package internal

type IndexEntry map[string]string

type ServiceIdentity struct {
	Identity    string      `json:"identity"`     // 业务系统用户唯一标识
	IndexedInfo IndexEntry  `json:"indexed_info"` // 业务系统用户信息（可搜索）
	ExtraInfo   interface{} `json:"extra_info"`   // 业务系统用户额外信息（不可搜索）
}

type PeerInfo struct {
	Protocol        string          `json:"protocol"`        // 连接协议，目前只支持JSON
	ClientID        string          `json:"client_id"`       // 客户端ID
	IP              string          `json:"ip"`              // 客户端IP
	Service         string          `json:"service"`         // 业务系统
	ServiceToken    string          `json:"service_token"`   // 业务系统认证信息
	ServiceIdentity ServiceIdentity `json:"client_identity"` // 业务系统客户端信息
}
