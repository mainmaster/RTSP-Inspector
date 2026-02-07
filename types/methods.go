package types

type RTSPMethod string

const (
	MethodOptions  RTSPMethod = "OPTIONS"
	MethodDescribe RTSPMethod = "DESCRIBE"
	MethodSetup    RTSPMethod = "SETUP"
	MethodPlay     RTSPMethod = "PLAY"
	MethodTeardown RTSPMethod = "TEARDOWN"
)
