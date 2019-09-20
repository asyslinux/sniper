package file

import "os"

// WriteAtPos - if pos<0 store at the end of file
func WriteAtPos(f *os.File, b []byte, pos int64) (seek int64, n int, err error) {
	seek = pos
	if pos < 0 {
		seek, err = f.Seek(0, 2)
		if err != nil {
			return
		}
	}
	n, err = f.WriteAt(b, seek)
	return
}
