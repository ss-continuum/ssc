package logbytes

import "fmt"

func lines(b []byte) []string {
	var lines []string
	var line string

	for i := 0; i < len(b); i++ {
		for k := 0; k < 16; k++ {
			if k == 8 {
				line += " "
			}
			if i+k >= len(b) {
				line += " "
				continue
			}
			c := b[i+k]

			if c < 32 || c > 126 {
				c = '.'
			}
			line = fmt.Sprintf("%s%c", line, c)
		}

		line += " |"
		for k := 0; k < 16; k++ {
			if k == 8 {
				line += " "
			}
			if i+k >= len(b) {
				line += "   "
				continue
			}

			line = fmt.Sprintf("%s %02x", line, b[i+k])
		}

		lines = append(lines, line)
		line = ""
		i += 16
	}
	return lines
}

func Log(b []byte) {
	data := lines(b)

	for _, datum := range data {
		fmt.Println(datum)
	}
}

func LogPrefix(b []byte, prefix string) {
	data := lines(b)

	for _, datum := range data {
		fmt.Printf("%s %s\n", prefix, datum)
	}
}
