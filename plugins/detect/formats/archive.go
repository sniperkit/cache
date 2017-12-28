package formats

/*
	Refs:
	- https://github.com/martinlindhe/formats/blob/master/parse/archive/arc_lzma.go
*/

func isGZIP(b []byte) bool {
	if b[0] != 0x1f || b[1] != 0x8b {
		return false
	}
	return true
}

func isZIP(b []byte) bool {
	if b[0] != 'P' || b[1] != 'K' || b[2] != 3 || b[3] != 4 {
		return false
	}
	return true
}

func isISO(b []byte) bool {
	pos := 0x8000
	if b[pos] != 1 || b[pos+1] != 'C' || b[pos+2] != 'D' {
		return false
	}
	return true
}

func isDEB(b []byte) bool {
	s := string(b[0:21])
	return s == "!<arch>\n"+"debian-binary"
}

func isBZIP2(b []byte) bool {
	if b[0] != 'B' || b[1] != 'Z' {
		return false
	}
	if b[2] != 'h' {
		// only huffman encoding is used in the format
		return false
	}
	return true
}

func isXZ(b []byte) bool {
	if b[0] != 0xfd || b[1] != '7' || b[2] != 'z' || b[3] != 'X' ||
		b[4] != 'Z' || b[5] != 0x00 {
		return false
	}
	return true
}

func isZLib(b []byte) bool {
	// XXX only matches zlib streams without dictionary.. this dont always work
	if b[0] != 0x78 {
		return false
	}
	if b[1] != 0x01 && b[1] != 0x9c && b[1] != 0xda {
		// compression level
		return false
	}
	return true
}

func isXAR(b []byte) bool {
	if b[0] != 'x' || b[1] != 'a' || b[2] != 'r' || b[3] != '!' {
		return false
	}
	return true
}

func isRAR(b []byte) bool {
	if b[0] != 'R' || b[1] != 'a' || b[2] != 'r' || b[3] != '!' {
		return false
	}
	// RAR 4.x signature
	//if (ReadByte() != 0x1A || ReadByte() != 0x07 || ReadByte() != 0x00)
	//    return false;
	// RAR 5.0 signature
	//if (ReadByte() != 0x1A || ReadByte() != 0x07 || ReadByte() != 0x01 || ReadByte() != 0x00)
	//    return false;
	return true
}

func isLZMA(b []byte) bool {
	// XXX not proper magic , need other check
	if b[0] != 0x5d || b[1] != 0x00 {
		return false
	}
	return true
}

func isLZH(b []byte) bool {
	if b[2] != '-' || b[3] != 'l' {
		return false
	}
	if b[4] == 'h' || b[4] == 'z' {
		return true
	}
	return false
}
