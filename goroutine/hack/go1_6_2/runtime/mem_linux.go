// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"third/goroutine/hack/go1_6_2/runtime/internal/sys"
)

const (
	_PAGE_SIZE = sys.PhysPageSize
	_EACCES    = 13
)
