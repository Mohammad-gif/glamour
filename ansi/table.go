package ansi

import (
	"bytes"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Add default padding to cells.
var cellStyle = lipgloss.NewStyle().Padding(0, 1)

// A TableElement is used to render tables.
type TableElement struct {
	lipgloss *table.Table
	headers  []string
	row      []string
}

// A TableRowElement is used to render a single row in a table.
type TableRowElement struct{}

// A TableHeadElement is used to render a table's head element.
type TableHeadElement struct{}

// A TableCellElement is used to render a single cell in a row.
type TableCellElement struct {
	Text string
	Head bool
}

func (e *TableElement) Render(w io.Writer, ctx RenderContext) error {
	bs := ctx.blockStack

	rules := ctx.options.Styles.Table
	style := bs.With(rules.StylePrimitive)

	renderText(w, ctx.options.ColorProfile, bs.Current().Style.StylePrimitive, rules.BlockPrefix)
	renderText(w, ctx.options.ColorProfile, style, rules.Prefix)
	ctx.table.lipgloss = table.New().StyleFunc(func(row, col int) lipgloss.Style { return cellStyle })
	return nil
}

// setBorders sets the borders for the lipgloss table. It uses the default
// border styles if no custom styles are set.
func (ctx *RenderContext) setBorders() {
	// TODO restore ability to use custom tables
	rules := ctx.options.Styles.Table
	border := lipgloss.NormalBorder()

	if rules.RowSeparator != nil && rules.ColumnSeparator != nil {
		border = lipgloss.Border{
			Top:    *rules.RowSeparator,
			Bottom: *rules.RowSeparator,
			Left:   *rules.ColumnSeparator,
			Right:  *rules.ColumnSeparator,
			Middle: *rules.CenterSeparator,
		}
	}
	ctx.table.lipgloss.Border(border)
	ctx.table.lipgloss.BorderTop(false)
	ctx.table.lipgloss.BorderLeft(false)
	ctx.table.lipgloss.BorderRight(false)
	ctx.table.lipgloss.BorderBottom(false)
}

func (e *TableElement) Finish(w io.Writer, ctx RenderContext) error {
	rules := ctx.options.Styles.Table
	ctx.setBorders()

	ow := ctx.blockStack.Current().Block

	// TODO should prefix, suffix, and margins etc all be handled in the parent writer?
	ow.Write([]byte(ctx.table.lipgloss.Render()))
	renderText(ow, ctx.options.ColorProfile, ctx.blockStack.With(rules.StylePrimitive), rules.Suffix)
	renderText(ow, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, rules.BlockSuffix)

	ctx.table.lipgloss = nil
	return nil
}

func (e *TableRowElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}
	if len(ctx.table.row) == 0 {
		panic(fmt.Sprintf("got an empty row %#v", ctx.table.row))
	}

	ctx.table.lipgloss.Row(ctx.table.row...)
	ctx.table.row = []string{}
	return nil
}

func (e *TableHeadElement) Finish(w io.Writer, ctx RenderContext) error {
	if ctx.table.lipgloss == nil {
		return nil
	}

	ctx.table.lipgloss.Headers(ctx.table.headers...)
	ctx.table.headers = []string{}
	return nil
}

func (e *TableCellElement) Render(w io.Writer, ctx RenderContext) error {
	// Style the text
	var tmp bytes.Buffer
	renderText(&tmp, ctx.options.ColorProfile, ctx.options.Styles.Table.StylePrimitive, e.Text)

	// Append to the current row
	if e.Head {
		ctx.table.headers = append(ctx.table.headers, tmp.String())
	} else {
		ctx.table.row = append(ctx.table.row, tmp.String())
	}

	return nil
}
