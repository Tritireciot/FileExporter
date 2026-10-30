package template_compiler

import "PrintServer/PrintServer/diagnostic"

type cursor struct {
	source []rune

	index  int
	line   int
	column int
}

func newCursor(source string) cursor {
	return cursor{
		source: []rune(source),
		line:   1,
		column: 1,
	}
}

func newCursorAt(
	source string,
	start diagnostic.Location,
) cursor {
	return cursor{
		source: []rune(source),
		line:   start.Line,
		column: start.Column,
	}
}

func (c *cursor) eof() bool {
	return c.index >= len(c.source)
}

func (c *cursor) current() (rune, bool) {
	if c.eof() {
		return 0, false
	}

	return c.source[c.index], true
}

func (c *cursor) peek(offset int) (rune, bool) {
	index := c.index + offset

	if index < 0 || index >= len(c.source) {
		return 0, false
	}

	return c.source[index], true
}

func (c *cursor) startsWith(value string) bool {
	index := c.index

	for _, expected := range value {
		if index >= len(c.source) {
			return false
		}

		if c.source[index] != expected {
			return false
		}

		index++
	}

	return true
}

func (c *cursor) advance(count int) {
	if count < 0 {
		panic("cursor.advance: negative count")
	}

	if c.index+count > len(c.source) {
		panic("cursor.advance: movement beyond source")
	}

	for i := 0; i < count; i++ {
		current := c.source[c.index]
		c.index++

		if current == '\n' {
			c.line++
			c.column = 1
		} else {
			c.column++
		}
	}
}

func (c *cursor) position() diagnostic.Location {
	return diagnostic.Location{
		Line:   c.line,
		Column: c.column,
	}
}

func (c *cursor) runeIndex() int {
	return c.index
}

func (c *cursor) slice(start, end int) string {
	if start < 0 || end < start || end > len(c.source) {
		panic("cursor.slice: invalid bounds")
	}

	return string(c.source[start:end])
}
