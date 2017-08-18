package log

import (
	"os"
	"testing"
)

func TestFiles(t *testing.T) {
	os.Remove("./logs")
	w := NewFileWriter(FileOptions{
		ByType: ByDay,
		Dir:    "./logs",
	})
	Std.SetOutput(w)
	Std.SetFlags(Std.Flags() | ^Llongcolor | ^Lshortcolor)
	Info("test")
	Debug("ssss")
	Warn("ssss")
	w.Close()
}
