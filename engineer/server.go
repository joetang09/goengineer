package engineer

var (
	svrBox = []Server{}
)

type Server interface {
	Serve()
	Stop()
}

func RegisterServer(svr Server) {
	svrBox = append(svrBox, svr)
}
