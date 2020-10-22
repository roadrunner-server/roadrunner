// +build debug

package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

type stack struct {
	callers []uintptr
	// TODO(adg): add time of creation
}

func (e *Error) populateStack() {
	e.callers = callers()

	e2, ok := e.Err.(*Error)
	if !ok {
		return
	}

	i := 0

	ok = false
	for ; i < len(e.callers) && i < len(e2.callers); i++ {
		// check for similar
		if e.callers[len(e.callers)-1-i] != e2.callers[len(e2.callers)-1-i] {
			break
		}
		ok = true
	}

	if ok { //we have common PCs
		e2Head := e2.callers[:len(e2.callers)-i]
		eTail := e.callers

		e.callers = make([]uintptr, len(e2Head)+len(eTail))

		copy(e.callers, e2Head)
		copy(e.callers[len(e2Head):], eTail)

		e2.callers = nil
	}
}

// frame returns the nth frame, with the frame at top of stack being 0.
func frame(callers []uintptr, n int) runtime.Frame {
	frames := runtime.CallersFrames(callers)
	var f runtime.Frame
	for i := len(callers) - 1; i >= n; i-- {
		var ok bool
		f, ok = frames.Next()
		if !ok {
			break
		}
	}
	return f
}

func (e *Error) printStack(b *bytes.Buffer) {
	c := callers()

	var prev string
	var diff bool
	for i := 0; i < len(e.callers); i++ {
		pc := e.callers[len(e.callers)-i-1] // get current PC
		fn := runtime.FuncForPC(pc)         // get function by pc
		name := fn.Name()

		if !diff && i < len(c) {
			ppc := c[len(c)-i-1]
			pname := runtime.FuncForPC(ppc).Name()
			if name == pname {
				continue
			}
			diff = true
		}

		if name == prev {
			continue
		}

		trim := 0
		for {
			j := strings.IndexAny(name[trim:], "./")
			if j < 0 {
				break
			}
			if !strings.HasPrefix(prev, name[:j+trim]) {
				break
			}
			trim += j + 1 // skip over the separator
		}

		// Do the printing.
		appendStrToBuf(b, Separator)
		file, line := fn.FileLine(pc)
		fmt.Fprintf(b, "%v:%d: ", file, line)
		if trim > 0 {
			b.WriteString("...")
		}
		b.WriteString(name[trim:])

		prev = name
	}
}

func callers() []uintptr {
	var stk [64]uintptr
	const skip = 4
	n := runtime.Callers(skip, stk[:])
	return stk[:n]
}
