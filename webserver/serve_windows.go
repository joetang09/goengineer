package webserver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (WebServer) Serve() {

	svrs = svrs[:0]

	svrs = append(svrs, buildAppSrv())

	if config.Pprof {

		addrDiv := strings.Split(config.Addr, ":")
		p, _ := strconv.Atoi(addrDiv[1])
		svrs = append(svrs, buildPPROFSrv(p+1))

	}

	sc := make(chan struct{}, len(svrs))
	for _, s := range svrs {
		go func(s *http.Server) {
			if err := s.ListenAndServe(); err != nil {
				fmt.Printf("Server Listen : %s \n", err)
				sc <- struct{}{}
			}
		}(s)
	}
	<-sc

}
