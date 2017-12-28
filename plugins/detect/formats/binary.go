package formats

func isSQLITE3(b []byte) bool {
	s := string(b[0:16])
	return s == "SQLite format 3"
}
