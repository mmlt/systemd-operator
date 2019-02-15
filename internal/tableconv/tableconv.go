package tableconv

import (
	"bufio"
	"io"
	"strings"
)

type Table struct {
	Header Header
	Rows   [][]string
}

type Header []Pos

type Pos struct {
	Text string
	Col  int
}

// Scan reads a text formatted table.
// Prereqs
// - text table has headers
// - only rows that have at least minCols are added to the output.
func Scan(r io.Reader, minCols int) *Table {
	scanner := bufio.NewScanner(r)

	// scan Header
	ok := scanner.Scan()
	if !ok {
		return &Table{}
	}
	hdr := scanHeader([]byte(scanner.Text()))

	answer := &Table{Header: hdr}
	for scanner.Scan() {
		ln := []byte(scanner.Text())
		if len(ln) == 0 {
			// stop on first empty line
			break
		}

		strt := 0
		var row []string
		for _, p := range hdr {
			if p.Col+1 >= len(ln) {
				break
			}

			var s []byte
			if p.Col == -1 {
				s = ln[strt:]
			} else {
				s = ln[strt:p.Col]
			}

			row = append(row, strings.TrimSpace(string(s)))
			strt = p.Col
		}
		if len(row) >= minCols {
			answer.Rows = append(answer.Rows, row)
		}
	}

	return answer
}

// ColNamesToIndices returns a map of columns names to indices.
func (t *Table) ColNamesToIndices() map[string]int {
	res := make(map[string]int, len(t.Header))
	for k,v := range t.Header {
		res[v.Text] = k
	}
	return res
}

func scanHeader(s []byte) Header {
	var isSpace bool
	var answer Header

	start := 0
	for i:=start; i<len(s); i++ {
		if isSpace && !isWhitepace(s[i]) {
			// right side of column found
			t := strings.TrimSpace(string(s[start:i]))
			answer = append(answer, Pos{Text: t, Col: i-1})
			start = i
		}
		isSpace = isWhitepace(s[i])
	}
	answer = append(answer, Pos{Text: strings.TrimSpace(string(s[start:])), Col: -1})

	return answer
}

func isWhitepace(b byte) bool {
	return b == 32
}


func FormatText(table *Table, columns []int, sep string, hdr bool, output io.Writer) {
	last := len(columns)

	if hdr {
		// output Header
		for i, ci := range columns {
			s := table.Header[ci].Text
			io.WriteString(output, s)
			if i < last-1 {
				io.WriteString(output, sep)
			}
		}
		io.WriteString(output, "\n")
	}

	// output Rows
	for _,row := range table.Rows {
		for i,ci := range columns {
			s := row[ci]
			io.WriteString(output, s)
			if i < last-1 {
				io.WriteString(output, sep)
			}
		}
		io.WriteString(output, "\n")
	}
}

func columnIndices(t *Table, c string) []int {
	// get map of column names
	var colMap map[string]bool
	if c == "" {
		colMap = header2Map(t.Header)
	} else {
		colMap = string2Map(c)
	}
	// map column names to indices
	var answer []int
	for i,c := range t.Header {
		if colMap[c.Text] {
			answer = append(answer, i)
		}
	}

	return answer
}

func string2Map(s string) map[string]bool {
	answer := make(map[string]bool, 3)
	for _,i := range strings.Split(s,",") {
		answer[i] = true
	}
	return answer
}

func header2Map(h Header) map[string]bool {
	answer := make(map[string]bool, 3)
	for _,c := range h {
		answer[c.Text] = true
	}
	return answer
}