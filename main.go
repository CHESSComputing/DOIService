package main

// Go implementation of FOXDEN DOI service
//
// Copyright (c) 2025 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	_ "expvar"         // to be used for monitoring, see https://github.com/divan/expvarmon
	_ "net/http/pprof" // profiler, see https://golang.org/pkg/net/http/pprof/

	srvConfig "github.com/CHESSComputing/golib/config"
)

func main() {
	srvConfig.Init()
	Server()
}
