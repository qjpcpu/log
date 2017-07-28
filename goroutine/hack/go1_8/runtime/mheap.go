// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.go for overview.

package runtime

import (
	"github.com/qjpcpu/log/goroutine/hack/go1_8/runtime/internal/sys"
	"unsafe"
)

// minPhysPageSize is a lower-bound on the physical page size. The
// true physical page size may be larger than this. In contrast,
// sys.PhysPageSize is an upper-bound on the physical page size.
const minPhysPageSize = 4096

// Main malloc heap.
// The heap itself is the "free[]" and "large" arrays,
// but all the other global data is here too.
//
// mheap must not be heap-allocated because it contains mSpanLists,
// which must not be heap-allocated.
//
//go:notinheap
type mheap struct {
	lock      mutex
	free      [_MaxMHeapList]mSpanList
	freelarge mSpanList
	busy      [_MaxMHeapList]mSpanList
	busylarge mSpanList
	sweepgen  uint32
	sweepdone uint32

	allspans []*mspan

	spans []*mspan

	sweepSpans [2]gcSweepBuf

	_ uint32

	pagesInUse        uint64
	spanBytesAlloc    uint64
	pagesSwept        uint64
	sweepPagesPerByte float64

	largefree  uint64
	nlargefree uint64
	nsmallfree [_NumSizeClasses]uint64

	bitmap         uintptr
	bitmap_mapped  uintptr
	arena_start    uintptr
	arena_used     uintptr
	arena_end      uintptr
	arena_reserved bool

	central [_NumSizeClasses]struct {
		mcentral mcentral
		pad      [sys.CacheLineSize]byte
	}

	spanalloc             fixalloc
	cachealloc            fixalloc
	specialfinalizeralloc fixalloc
	specialprofilealloc   fixalloc
	speciallock           mutex
}

// An MSpan representing actual memory has state _MSpanInUse,
// _MSpanStack, or _MSpanFree. Transitions between these states are
// constrained as follows:
//
// * A span may transition from free to in-use or stack during any GC
//   phase.
//
// * During sweeping (gcphase == _GCoff), a span may transition from
//   in-use to free (as a result of sweeping) or stack to free (as a
//   result of stacks being freed).
//
// * During GC (gcphase != _GCoff), a span *must not* transition from
//   stack or in-use to free. Because concurrent GC may read a pointer
//   and then look up its span, the span state must be monotonic.
type mSpanState uint8

const (
	_MSpanDead mSpanState = iota
	_MSpanInUse
	_MSpanStack
	_MSpanFree
)

// mSpanList heads a linked list of spans.
//
//go:notinheap
type mSpanList struct {
	first *mspan
	last  *mspan
}

//go:notinheap
type mspan struct {
	next *mspan
	prev *mspan
	list *mSpanList

	startAddr     uintptr
	npages        uintptr
	stackfreelist gclinkptr

	freeindex uintptr

	nelems uintptr

	allocCache uint64

	allocBits  *uint8
	gcmarkBits *uint8

	sweepgen    uint32
	divMul      uint16
	baseMask    uint16
	allocCount  uint16
	sizeclass   uint8
	incache     bool
	state       mSpanState
	needzero    uint8
	divShift    uint8
	divShift2   uint8
	elemsize    uintptr
	unusedsince int64
	npreleased  uintptr
	limit       uintptr
	speciallock mutex
	specials    *special
}

const (
	_KindSpecialFinalizer = 1
	_KindSpecialProfile   = 2
)

//go:notinheap
type special struct {
	next   *special
	offset uint16
	kind   byte
}

// The described object has a finalizer set for it.
//
// specialfinalizer is allocated from non-GC'd memory, so any heap
// pointers must be specially handled.
//
//go:notinheap
type specialfinalizer struct {
	special special
	fn      *funcval
	nret    uintptr
	fint    *_type
	ot      *ptrtype
}

// The described object is being heap profiled.
//
//go:notinheap
type specialprofile struct {
	special special
	b       *bucket
}

const gcBitsChunkBytes = uintptr(64 << 10)
const gcBitsHeaderBytes = unsafe.Sizeof(gcBitsHeader{})

type gcBitsHeader struct {
	free uintptr
	next uintptr
}

//go:notinheap
type gcBits struct {
	free uintptr
	next *gcBits
	bits [gcBitsChunkBytes - gcBitsHeaderBytes]uint8
}
