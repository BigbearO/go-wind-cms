package serviceid

// NewDiscoveryName 构建服务发现名称
func NewGrpcDiscoveryName(serviceName string) string {
	return ProjectName + "/" + serviceName + ".grpc"
}

// MakeDiscoveryAddress 构建服务发现地址
func MakeGrpcDiscoveryAddress(serviceName string) string {
	return "discovery:///" + serviceName + ".grpc"
}
