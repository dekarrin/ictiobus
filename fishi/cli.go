package fishi

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dekarrin/rosed"
)

// WarnHandling is a type of handling that should be used for a particular type
// of Warning.
type WarnHandling int

const (
	// WarnHandlingOutput is the default handling of a warning, and indicates
	// it should be output as normal.
	WarnHandlingOutput WarnHandling = iota

	// WarnHandlingSuppress indicates that a warning should be suppressed and
	// that no action should be taken when it is encountered.
	WarnHandlingSuppress

	// WarnHandlingFatal indicates that a warning should be treated as a fatal
	// error, causing immediate termination of the current procedure and
	// possibly the entire program.
	WarnHandlingFatal
)

// String returns the string represtnation of a WarnHandling.
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

	// Output is the writer that output is sent to. By default this will be
	// os.Stderr in newly-created WarnHandlers, but this can be changed by
	// setting Output to a different io.Writer.
	Output io.Writer

	// ErrorPrefix is the string to prepend error messages written to Output
	// with.
	ErrorPrefix string

	// WarnPrefix is the string to prepend error messages written to Output
	// with.
	WarnPrefix string
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

// NewWarnHandler creates a new WarnHandler with default settings.
func NewWarnHandler() *WarnHandler {
	return &WarnHandler{
		h:           map[WarnType]WarnHandling{},
		Output:      os.Stderr,
		ErrorPrefix: "ERROR: ",
		WarnPrefix:  "WARN: ",
	}
}

// NewWarnHandlerFromCLI creates a new WarnHandler by using the provided options
// for setting short-codes of WarnTypes, presumably as provided from CLI flags.
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
