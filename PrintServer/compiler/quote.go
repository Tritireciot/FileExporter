package template_compiler

import "PrintServer/PrintServer/diagnostic"

func scanConditionString(c *cursor) (raw string, diag *diagnostic.Diagnostic) {
	quote, ok := c.current()
    if !ok || (quote != '"' && quote != '\'') {
        panic("scanConditionString: курсор должен стоять на кавычке")
    }

	start := c.position()
    startIndex := c.runeIndex()
    c.advance(1)

	for !c.eof() {
		current, _ := c.current()

        if current == '\n' {
            break
        }

        if current == '\\' {
            c.advance(1)
            if c.eof() {
                break
            }
            next, _ := c.current()
            if next == '\n' {
                break
            }
            c.advance(1)
            continue
        }

        c.advance(1)
        if current == quote {
            return c.slice(startIndex, c.runeIndex()), nil
        }
	}

	return "", diagnostic.NewTranspile(
        diagnostic.CodeUnclosedString,
        "Строковый литерал в условии не закрыт",
    ).
        WithHint("Закройте строку такой же кавычкой").
        WithSpan(start, c.position())
}