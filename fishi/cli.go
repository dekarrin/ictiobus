package fishi

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dekarrin/rosed"
)

type WarnHandling int

const (
	WarnHandlingOutput WarnHandling = iota
	WarnHandlingSuppress
	WarnHandlingFatal
)

func (wh WarnHandling) String() string {
	if wh == WarnHandlingFatal {
		return "FATAL"
	} else if wh == WarnHandlingOutput {
		return "OUTPUT"
	} else if wh == WarnHandlingSuppress {
		return "SUPPRESS"
	} else {
		return fmt.Sprintf("WarnHandling(%d)", int(wh))
	}
}

// WarnHandler handles warnings and can be configured by reading slices of names
// of warnings. The zero-value is not ready to be used; call NewWarnHandler() or
// NewWarnHandlerFromCLI to create one.
type WarnHandler struct {
	h map[WarnType]WarnHandling

	Output      io.Writer
	ErrorPrefix string
	WarnPrefix  string
}

// HandlingType returns the WarnHandling configured for the given warning type.
//
// If the warning type has no handling defined for it, the default of
// WarnHandlingOutput will be returned.
func (wh *WarnHandler) HandlingType(t WarnType) WarnHandling {
	return wh.h[t]
}

// Suppressed sets the handling for a type of warning to be 'suppression'.
// Suppressed warnings are not shown in output and no further action is taken.
func (wh *WarnHandler) Suppressed(wt WarnType) {
	wh.h[wt] = WarnHandlingSuppress
}

// Fatal sets the handling for the given warn type to be 'fatal'. Fatal warnings
// are shown with the Error prefix and will cause handling to return a non-nil
// error containing a message that the warning is treated as a fatal error.
func (wh *WarnHandler) Fatal(wt WarnType) {
	wh.h[wt] = WarnHandlingFatal
}

// Output sets the handling for the given warn type to be 'normal'. Normal
// warnings are shown with the Warn prefix.
func (wh *WarnHandler) Normal(wt WarnType) {
	wh.h[wt] = WarnHandlingOutput
}

// Handle performs whatever handling has been specified for the given warning.
// It returns a non-nil error if and only if the warning is treated as fatal.
func (wh *WarnHandler) Handle(w Warning) (fatal error) {
	return wh.Handlef("%s\n", w)
}

// Handlef is identical to Handle but allows a custom format string to be
// supplied. It returns a non-nil error if and only if the warning is treated as
// fatal.
func (wh *WarnHandler) Handlef(fmtStr string, w Warning) (fatal error) {
	var prefix string

	handleType := wh.h[w.Type]

	switch handleType {
	case WarnHandlingSuppress:
		// do nothing
		return nil
	case WarnHandlingFatal:
		fatal = fmt.Errorf("%q warnings are treated as fatal", w.Type.Short())
		prefix = wh.ErrorPrefix
	case WarnHandlingOutput:
		fatal = nil
		prefix = wh.WarnPrefix
	}

	msg := prefix + w.Message
	// okay, not suppressed, so output the warning
	if strings.Contains(msg, "\n") {
		// rosed will help us here;

		// indent all except the first line
		msg = rosed.Edit(msg).
			LinesFrom(1).
			IndentOpts(len(prefix), rosed.Options{IndentStr: " "}).
			String()
	}

	fmt.Fprintf(wh.Output, fmtStr, msg)
	return fatal
}

func NewWarnHandler() *WarnHandler {
	return &WarnHandler{
		h:           map[WarnType]WarnHandling{},
		Output:      os.Stderr,
		ErrorPrefix: "ERROR: ",
		WarnPrefix:  "WARN: ",
	}
}

func NewWarnHandlerFromCLI(suppressions, fatals []string) (*WarnHandler, error) {
	handler := NewWarnHandler()

	for _, wt := range WarnTypeAll() {
		handler.h[wt] = WarnHandlingOutput
	}

	var allFatal bool

	// do fatals first, then do suppressions
	for i := range fatals {
		warnType := fatals[i]

		if strings.ToLower(warnType) == "all" {
			allFatal = true
			for k := range handler.h {
				handler.h[k] = WarnHandlingFatal
			}
		} else {
			wt, err := ParseShortWarnType(warnType)
			if err != nil {
				return nil, err
			}

			handler.h[wt] = WarnHandlingFatal
		}
	}

	// now do suppressions, and give an error if the user tries to both suppress
	// and make fatal all flags.
	for i := range suppressions {
		warnType := suppressions[i]

		if strings.ToLower(warnType) == "all" {
			if allFatal {
				return nil, fmt.Errorf("cannot suppress all warns while also treating all as fatal")
			}

			for k := range handler.h {
				if handler.h[k] != WarnHandlingFatal {
					handler.h[k] = WarnHandlingSuppress
				}
			}
		} else {
			wt, err := ParseShortWarnType(warnType)
			if err != nil {
				return nil, err
			}

			if handler.h[wt] != WarnHandlingFatal {
				// fatal takes precednece
				handler.h[wt] = WarnHandlingSuppress
			}
		}
	}

	return handler, nil
}
