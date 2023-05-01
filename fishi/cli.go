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

// WarnHandler handles warnings and can be configured by reading slices of names
// of warnings. The zero-value is a WarnHandler ready to be used.
type WarnHandler struct {
	h map[WarnType]WarnHandling

	Output      io.Writer
	ErrorPrefix string
	WarnPrefix  string
}

// Suppressed sets the handling for a type of warning to be 'suppression'.
// Suppressed warnings are not shown in output and no further action is taken.
func (wh *WarnHandler) Suppressed(wt WarnType) {
	if wh.h == nil {
		wh.h = map[WarnType]WarnHandling{}
	}

	wh.h[wt] = WarnHandlingSuppress
}

// Fatal sets the handling for the given warn type to be 'fatal'. Fatal warnings
// are shown with the Error prefix and will cause handling to return a non-nil
// error containing a message that the warning is treated as a fatal error.
func (wh *WarnHandler) Fatal(wt WarnType) {
	if wh.h == nil {
		wh.h = map[WarnType]WarnHandling{}
	}

	wh.h[wt] = WarnHandlingFatal
}

// Output sets the handling for the given warn type to be 'normal'. Normal
// warnings are shown with the Warn prefix.
func (wh *WarnHandler) Normal(wt WarnType) {
	if wh.h == nil {
		wh.h = map[WarnType]WarnHandling{}
	}

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

	handleType := handler[w.Type]

	switch handleType {
	case WarnHandlingSuppress:
		// do nothing
		return false
	case WarnHandlingFatal:
		fatal = true
		prefix = errorPrefix
	case WarnHandlingOutput:
		fatal = false
		prefix = warnPrefix
	}

	msg := w.Message
	// okay, not suppressed, so output the warning
	if strings.Contains(msg, "\n") {
		// rosed will help us here;

		// indent all except the first line
		msg = rosed.Edit(prefix+msg).
			LinesFrom(1).
			IndentOpts(len(prefix), rosed.Options{IndentStr: " "}).
			String()
	}

	fmt.Fprintf(os.Stderr, fmtStr, msg)
	return fatal
}

func NewWarnHandlerFromCLI(suppressions, fatals []string) (*WarnHandler, error) {
	handler := WarnHandler{}

	for _, wt := range fishi.WarnTypeAll() {
		handling[wt] = WarnHandlingOutput
	}

	var allFatal bool

	// do fatals first, then do suppressions
	for i := range *flagWarnFatal {
		warnType := (*flagWarnFatal)[i]

		if strings.ToLower(warnType) == "all" {
			allFatal = true
			for k := range handling {
				handling[k] = WarnHandlingFatal
			}
		} else {
			wt, err := fishi.ParseShortWarnType(warnType)
			if err != nil {
				return nil, err
			}

			handling[wt] = WarnHandlingFatal
		}
	}

	// now do suppressions, and give an error if the user tries to both suppress
	// and make fatal all flags.
	for i := range *flagWarnSuppress {
		warnType := (*flagWarnSuppress)[i]

		if strings.ToLower(warnType) == "all" {
			if allFatal {
				return nil, fmt.Errorf("cannot suppress all warns while also treating all as fatal")
			}

			for k := range handling {
				if handling[k] != WarnHandlingFatal {
					handling[k] = WarnHandlingSuppress
				}
			}
		} else {
			wt, err := fishi.ParseShortWarnType(warnType)
			if err != nil {
				return nil, err
			}

			if handling[wt] != WarnHandlingFatal {
				// fatal takes precednece
				handling[wt] = WarnHandlingSuppress
			}
		}
	}

	return handling, nil
}
