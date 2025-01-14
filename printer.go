package synapsecleaner

import (
	"fmt"

	"github.com/morikuni/aec"
	"golang.org/x/term"
)

type Line interface {
	Id() string
	String() string
}

type LinesPrinter struct {
	positions map[string]uint

	footerHeight  int
	footerPrinted bool
	termHeight    int
}

func NewLinesPrinter() *LinesPrinter {
	return &LinesPrinter{
		positions: make(map[string]uint),
	}
}

func NewLinesPrinterWithFooter(height int) (*LinesPrinter, error) {
	lp := NewLinesPrinter()

	_, termHeight, err := term.GetSize(0)
	if err != nil {
		return nil, err
	}

	lp.footerHeight = height
	lp.footerPrinted = false
	lp.termHeight = termHeight

	return lp, nil
}

func (lp *LinesPrinter) Print(line Line) {
	position, ok := lp.positions[line.Id()]
	if !ok {
		position = uint(len(lp.positions))
		lp.positions[line.Id()] = position
		lp.print("\n")

		// If the footer was already printed, and we are at the bottom of the term,
		// we make a new space for the footer.
		// This way the last line is still above the footer.
		if lp.footerHeight > 0 && lp.footerPrinted && len(lp.positions) > lp.termHeight-lp.footerHeight {
			lp.cleanFooter()
		}
	}

	diff := uint(len(lp.positions)) - position
	lp.cursorUp(diff)
	lp.cleanLine()
	lp.print(line.String())

	// If the footer was already printed, and we are at the bottom of the term,
	// the last lines is above it. If we are not at the bottom of the term, or if
	// the footer was not printed yet, lines can still grow.
	if lp.footerHeight > 0 && lp.footerPrinted && diff > uint(lp.termHeight-lp.footerHeight) {
		diff = uint(lp.termHeight - lp.footerHeight)
	}

	lp.cursorDown(diff)
}

func (lp *LinesPrinter) PrintFooter(footer string) {
	// clean space for the footer
	lp.cleanLine()
	for i := 1; i < lp.footerHeight; i++ {
		fmt.Println()
		lp.cleanLine()
	}

	fmt.Print(aec.PreviousLine(uint(lp.footerHeight - 1)))

	// print footer
	fmt.Print(footer)
	fmt.Print(aec.PreviousLine(uint(lp.footerHeight - 1)))

	lp.footerPrinted = true
}

func (lp *LinesPrinter) Exit() {
	if lp.footerHeight > 0 && lp.footerPrinted {
		fmt.Print(aec.NextLine(uint(lp.footerHeight)))
	}
}

func (lp *LinesPrinter) cleanLine() {
	fmt.Print(aec.EraseLine(aec.EraseModes.All))
}

func (lp *LinesPrinter) cleanFooter() {
	lp.cleanLine()
	for i := 1; i < lp.footerHeight; i++ {
		fmt.Println()
		lp.cleanLine()
	}

	fmt.Print(aec.PreviousLine(uint(lp.footerHeight - 1)))
}

func (lp *LinesPrinter) cursorUp(diff uint) {
	fmt.Print(aec.PreviousLine(diff))
}

func (lp *LinesPrinter) cursorDown(diff uint) {
	fmt.Print(aec.NextLine(diff))
}

func (lp *LinesPrinter) print(a ...any) {
	fmt.Print(a...)
	//time.Sleep(100 * time.Millisecond)
}
