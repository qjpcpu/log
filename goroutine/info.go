// Copyright 2016 Huan Du. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goroutine

import (
	"unsafe"

	runtime1_5 "third/goroutine/hack/go1_5/runtime"
	runtime1_5_1 "third/goroutine/hack/go1_5_1/runtime"
	runtime1_5_2 "third/goroutine/hack/go1_5_2/runtime"
	runtime1_5_3 "third/goroutine/hack/go1_5_3/runtime"
	runtime1_5_4 "third/goroutine/hack/go1_5_4/runtime"
	runtime1_6 "third/goroutine/hack/go1_6/runtime"
	runtime1_6_1 "third/goroutine/hack/go1_6_1/runtime"
	runtime1_6_2 "third/goroutine/hack/go1_6_2/runtime"
	runtime1_6_3 "third/goroutine/hack/go1_6_3/runtime"
	runtime1_7 "third/goroutine/hack/go1_7/runtime"
	runtime1_7_1 "third/goroutine/hack/go1_7_1/runtime"
	runtime1_7_2 "third/goroutine/hack/go1_7_2/runtime"
	runtime1_7_3 "third/goroutine/hack/go1_7_3/runtime"
)

func getg() unsafe.Pointer

// GoroutineId return id of current goroutine.
// It's guaranteed to be unique globally during app's life time.
func GoroutineId() int64 {
	gp := getg()

	switch goVersionCode() {
	case _GO_VERSION1_5:
		return (*runtime1_5.Goroutine)(gp).Goid()
	case _GO_VERSION1_5_1:
		return (*runtime1_5_1.Goroutine)(gp).Goid()
	case _GO_VERSION1_5_2:
		return (*runtime1_5_2.Goroutine)(gp).Goid()
	case _GO_VERSION1_5_3:
		return (*runtime1_5_3.Goroutine)(gp).Goid()
	case _GO_VERSION1_5_4:
		return (*runtime1_5_4.Goroutine)(gp).Goid()
	case _GO_VERSION1_6:
		return (*runtime1_6.Goroutine)(gp).Goid()
	case _GO_VERSION1_6_1:
		return (*runtime1_6_1.Goroutine)(gp).Goid()
	case _GO_VERSION1_6_2:
		return (*runtime1_6_2.Goroutine)(gp).Goid()
	case _GO_VERSION1_6_3:
		return (*runtime1_6_3.Goroutine)(gp).Goid()
	case _GO_VERSION1_7:
		return (*runtime1_7.Goroutine)(gp).Goid()
	case _GO_VERSION1_7_1:
		return (*runtime1_7_1.Goroutine)(gp).Goid()
	case _GO_VERSION1_7_2:
		return (*runtime1_7_2.Goroutine)(gp).Goid()
	case _GO_VERSION1_7_3:
		return (*runtime1_7_3.Goroutine)(gp).Goid()

	default:
		panic("unsupported go version " + goVersion().String())
	}
}
