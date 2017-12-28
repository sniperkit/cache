package formats

func isText(c *Checker) bool {
	if hasRecognizedBOM(c.Header) {
		return true
	}
	seemsLike := ASCIIUS
	checkLen := int64(30)
	if c.ParsedLayout.FileSize < checkLen {
		checkLen = c.ParsedLayout.FileSize
	}

	for pos := int64(0); pos < checkLen; pos++ {
		if pos >= c.ParsedLayout.FileSize {
			break
		}

		if seemsLike == ASCIIUS {
			c := c.Header[pos]
			if c < 32 {
				if c != '\n' && c != '\r' && c != '\t' {
					return false
				}
			}
		}
	}
	return true
}

func parseBOMMark(b []byte) (TextEncoding, int64) {
	if b[0] == 0xff && b[1] == 0xfe && b[2] == 0 && b[3] == 0 {
		return UTF32le, 2
	}
	if b[0] == 0 && b[1] == 0 && b[2] == 0xfe && b[3] == 0xff {
		return UTF32be, 4
	}
	if b[0] == 0xfe && b[1] == 0xff {
		return UTF16be, 2
	}
	if b[0] == 0xff && b[1] == 0xfe {
		return UTF16le, 2
	}
	if b[0] == 0xef && b[1] == 0xbb && b[2] == 0xbf {
		return UTF8, 3
	}
	return None, 0
}

func hasRecognizedBOM(b []byte) bool {
	_, len := parseBOMMark(b)
	return len > 0
}
