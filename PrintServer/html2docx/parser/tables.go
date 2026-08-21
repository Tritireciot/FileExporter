package parser

import (
	"PrintServer/PrintServer/html2docx/css"
	"PrintServer/PrintServer/html2docx/html"
	"fmt"
	"strconv"

	"github.com/mmonterroca/docxgo/v2/domain"
	nethtml "golang.org/x/net/html"
)

func (p *DocumentParser) parseTableTags(node *nethtml.Node, virtualTable *[][]CellData, row *int, col *int, rawStyles css.StyleMap) {
	tagName := html.Tag(node.Data)

	currentRawStyle := rawStyles.Copy()
	p.templateStyles.CombineStyles(node, &currentRawStyle)

	if tagName.IsTable() {
		html.IterateOverChildren(node, func(child *nethtml.Node) {
			p.parseTableTags(child, virtualTable, row, col, currentRawStyle)
		})
	}

	switch tagName {
	case html.TrTag:
		*virtualTable = append(*virtualTable, []CellData{})
		*col = 0
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == nethtml.ElementNode {
				p.parseTableTags(child, virtualTable, row, col, currentRawStyle)
			}
		}
		if *row > 0 {
			previous_row := (*virtualTable)[*row-1]
			last_col := previous_row[len(previous_row)-1].Col
			current_row := (*virtualTable)[*row]
			for last_col >= *col {
				cell := previous_row[*col]
				current_row = append(current_row, CellData{Row: *row, Col: *col, Merge: []int{cell.Merge[0] - 1, cell.Merge[1]}, Skip: true, CellStyle: cell.CellStyle})
				*col += 1
			}
			(*virtualTable)[*row] = current_row
		}
		*row += 1

	case html.ThTag, html.TdTag:
		current_row := (*virtualTable)[*row]
		merge_row, merge_col := 1, 1
		if *row > 0 {
			previous_row := (*virtualTable)[*row-1]
			if len(previous_row) > *col {
				cell := previous_row[*col]
				if cell.Col == *col {
					for cell.Merge[0] > merge_row {
						current_row = append(
							current_row,
							CellData{
								Row:       *row,
								Col:       *col,
								Merge:     []int{cell.Merge[0] - 1, cell.Merge[1]},
								Skip:      true,
								CellStyle: cell.CellStyle,
								TextStyle: cell.TextStyle,
							},
						)
						*col += 1
						if len((*virtualTable)[*row-1]) == *col {
							break
						}
						cell = (*virtualTable)[*row-1][*col]
					}
				}
			}
		}

		textStyle := currentRawStyle.Copy()
		style := BuildStyleModel(currentRawStyle)

		current_row = append(current_row, CellData{Row: *row, Col: *col, Data: node.FirstChild, CellStyle: style, TextStyle: textStyle})

		if rowspan, ok := html.GetAttribute(node, html.RowSpanAttr); ok {
			rowspan_int, _ := strconv.Atoi(rowspan)
			merge_row = rowspan_int
		}
		if colspan, ok := html.GetAttribute(node, html.ColSpanAttr); ok {
			colspan_int, _ := strconv.Atoi(colspan)
			merge_col = colspan_int
		}
		current_row[len(current_row)-1].Merge = []int{merge_row, merge_col}
		for i := 1; i < merge_col; i++ {
			*col += 1
			current_row = append(current_row, CellData{Row: *row, Col: *col, Merge: []int{merge_row, merge_col - i}, Skip: true, CellStyle: style, TextStyle: textStyle})
		}
		(*virtualTable)[*row] = current_row
		*col += 1
	}
}

func (p *DocumentParser) handleTable(node *nethtml.Node, rawStyles css.StyleMap, source Source) {
	var virtualTable [][]CellData
	row, col := 0, 0
	currentRawStyle := rawStyles.Copy()
	p.templateStyles.CombineStyles(node, &currentRawStyle)
	html.IterateOverChildren(node, func(child *nethtml.Node) {
		p.parseTableTags(child, &virtualTable, &row, &col, currentRawStyle)
	})
	table, err := source.AddTable(row, col)
	if err != nil || table == nil {
		fmt.Printf("Ошибка создания таблицы: %v, размеры: %d x %d\n", err, row, col)
		return
	}
	table.SetWidth(domain.TableWidth{Type: domain.WidthPct, Value: 50 * 100})

	for i, table_row := range virtualTable {
		docs_row, _ := table.Row(i)
		for _, cell := range table_row {
			docs_cell, _ := docs_row.Cell(cell.Col)
			cell.CellStyle.ApplyToCell(&docs_cell)
			docs_cell.Merge(cell.Merge[1], cell.Merge[0])

			html.IterateOverSiblings(
				cell.Data, func(cell_subnode *nethtml.Node) {
					p.parseNode(cell_subnode, cell.TextStyle, docs_cell)
				},
			)

			//paragraph, _ := docs_cell.AddParagraph()
			//cell.CellStyle.ApplyToParagraph(&paragraph)

			//p.parseInlineElements(
			//	paragraph,
			//	&nethtml.Node{Type: nethtml.TextNode, Data: cell.Data},
			//	TagStyleState{},
			//	cell.TextStyle,
			//)
		}
	}
}
