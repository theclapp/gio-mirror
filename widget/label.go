// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"fmt"
	"image"
	"image/color"
	"unicode/utf8"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"

	"golang.org/x/image/math/fixed"
)

// Label is a widget for laying out and drawing text.
type Label struct {
	// Alignment specify the text alignment.
	Alignment text.Alignment
	// MaxLines limits the number of lines. Zero means no limit.
	MaxLines int
}

type Point image.Point

type segmentIterator struct {
	Lines     []text.Line
	Clip      image.Rectangle
	Alignment text.Alignment
	Width     int
	Offset    image.Point
	StartSel  Point
	EndSel    Point

	pos    Point       // current position
	line   text.Line   // current line
	layout text.Layout // current line's Layout

	// pixel positions
	off            fixed.Point26_6
	x, y, prevDesc fixed.Int26_6
}

const inf = 1e6

func (l *segmentIterator) Next() (text.Layout, image.Point, bool, int, layout.Dimensions, bool) {
	for l.pos.Y < len(l.Lines) {
		if l.pos.X == 0 {
			l.line = l.Lines[l.pos.Y]

			// Calculate X & Y pixel coordinates of left edge of line
			l.x = align(l.Alignment, l.line.Width, l.Width) + fixed.I(l.Offset.X)
			l.y += l.prevDesc + l.line.Ascent
			l.prevDesc = l.line.Descent
			// Align baseline and line start to the pixel grid.
			l.off = fixed.Point26_6{X: fixed.I(l.x.Floor()), Y: fixed.I(l.y.Ceil())}
			l.y = l.off.Y
			l.off.Y += fixed.I(l.Offset.Y)
			if (l.off.Y + l.line.Bounds.Min.Y).Floor() > l.Clip.Max.Y {
				break
			}

			if (l.off.Y + l.line.Bounds.Max.Y).Ceil() < l.Clip.Min.Y {
				// This line is outside/before the clip area; go on to the next line.
				l.pos.Y++
				continue
			}

			l.layout = l.line.Layout
			for len(l.layout.Advances) > 0 {
				_, n := utf8.DecodeRuneInString(l.layout.Text)
				adv := l.layout.Advances[0]
				if (l.off.X + adv + l.line.Bounds.Max.X - l.line.Width).Ceil() >= l.Clip.Min.X {
					break
				}
				l.off.X += adv
				l.layout.Text = l.layout.Text[n:]
				l.layout.Advances = l.layout.Advances[1:]
				l.pos.X++
			}
		}

		selected := l.inSelection()
		endx := l.off.X
		rune := 0
		nextLine := true
		retLayout := l.layout
		for n := range l.layout.Text {
			selChanged := selected != l.inSelection()
			if (endx+l.line.Bounds.Min.X).Floor() > l.Clip.Max.X || selChanged {
				retLayout.Advances = l.layout.Advances[:rune]
				retLayout.Text = l.layout.Text[:n]
				if selChanged {
					// Save the rest of the line
					l.layout.Advances = l.layout.Advances[rune:]
					l.layout.Text = l.layout.Text[n:]
					nextLine = false
				}
				break
			}
			endx += l.layout.Advances[rune]
			rune++
			l.pos.X++
		}
		offf := image.Point{X: l.off.X.Floor(), Y: l.off.Y.Floor()}

		// Calculate the width & height if the returned text.
		var d fixed.Int26_6
		for _, adv := range retLayout.Advances {
			d += adv
		}
		dims := layout.Dimensions{
			Size: image.Point{
				X: d.Ceil(),
				Y: -(l.line.Ascent + l.line.Descent).Ceil(),
			},
		}

		if nextLine {
			l.pos.Y++
			l.pos.X = 0
		} else {
			l.off.X = endx
		}

		return retLayout, offf, selected, l.line.Descent.Ceil(), dims, true
	}
	return text.Layout{}, image.Point{}, false, 0, layout.Dimensions{}, false
}

func (l *segmentIterator) inSelection() bool {
	return l.StartSel.LessOrEqual(l.pos) &&
		l.pos.Less(l.EndSel)
}

func (p1 Point) LessOrEqual(p2 Point) bool {
	return p1.Y < p2.Y || (p1.Y == p2.Y && p1.X <= p2.X)
}

func (p1 Point) Less(p2 Point) bool {
	return p1.Y < p2.Y || (p1.Y == p2.Y && p1.X < p2.X)
}

func (l Label) Layout(gtx layout.Context, s text.Shaper, font text.Font, size unit.Value, txt string) layout.Dimensions {
	cs := gtx.Constraints
	textSize := fixed.I(gtx.Px(size))
	lines := s.LayoutString(font, textSize, cs.Max.X, txt)
	if max := l.MaxLines; max > 0 && len(lines) > max {
		lines = lines[:max]
	}
	dims := linesDimens(lines)
	dims.Size = cs.Constrain(dims.Size)
	cl := textPadding(lines)
	cl.Max = cl.Max.Add(dims.Size)
	it := segmentIterator{
		Lines:     lines,
		Clip:      cl,
		Alignment: l.Alignment,
		Width:     dims.Size.X,
	}
	for {
		l, off, selected, yOffs, dims, ok := it.Next()
		if !ok {
			break
		}
		stack := op.Push(gtx.Ops)
		op.Offset(layout.FPt(off)).Add(gtx.Ops)
		if selected {
			drawHighlight(gtx, yOffs, dims)
		}
		s.Shape(font, textSize, l).Add(gtx.Ops)
		clip.Rect(cl.Sub(off)).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Pop()
	}
	return dims
}

func textPadding(lines []text.Line) (padding image.Rectangle) {
	if len(lines) == 0 {
		return
	}
	first := lines[0]
	if d := first.Ascent + first.Bounds.Min.Y; d < 0 {
		padding.Min.Y = d.Ceil()
	}
	last := lines[len(lines)-1]
	if d := last.Bounds.Max.Y - last.Descent; d > 0 {
		padding.Max.Y = d.Ceil()
	}
	if d := first.Bounds.Min.X; d < 0 {
		padding.Min.X = d.Ceil()
	}
	if d := first.Bounds.Max.X - first.Width; d > 0 {
		padding.Max.X = d.Ceil()
	}
	return
}

func linesDimens(lines []text.Line) layout.Dimensions {
	var width fixed.Int26_6
	var h int
	var baseline int
	if len(lines) > 0 {
		baseline = lines[0].Ascent.Ceil()
		var prevDesc fixed.Int26_6
		for _, l := range lines {
			h += (prevDesc + l.Ascent).Ceil()
			prevDesc = l.Descent
			if l.Width > width {
				width = l.Width
			}
		}
		h += lines[len(lines)-1].Descent.Ceil()
	}
	w := width.Ceil()
	return layout.Dimensions{
		Size: image.Point{
			X: w,
			Y: h,
		},
		Baseline: h - baseline,
	}
}

func align(align text.Alignment, width fixed.Int26_6, maxWidth int) fixed.Int26_6 {
	mw := fixed.I(maxWidth)
	switch align {
	case text.Middle:
		return fixed.I(((mw - width) / 2).Floor())
	case text.End:
		return fixed.I((mw - width).Floor())
	case text.Start:
		return 0
	default:
		panic(fmt.Errorf("unknown alignment %v", align))
	}
}

func drawHighlight(gtx layout.Context, yOffs int, dims layout.Dimensions) {
	stack := op.Push(gtx.Ops)
	op.Offset(f32.Point{Y: float32(yOffs)}).Add(gtx.Ops)
	paint.FillShape(gtx.Ops,
		color.NRGBA{B: 0xff, A: 0x40},
		clip.UniformRRect(
			f32.Rectangle{
				Max: layout.FPt(dims.Size),
			},
			0).Op(gtx.Ops))
	stack.Pop()
}
