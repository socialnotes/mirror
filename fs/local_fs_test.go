package fs

import (
	"io/ioutil"
	"os"
	"testing"
)

func withTempDir(t *testing.T, f func(d Dir)) {
	dirPath, err := ioutil.TempDir("", "tests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)
	f(Dir(dirPath))
}

func TestDir(t *testing.T) {
	fpath := "test/123/f.txt"
	content := "content"
	withTempDir(t, func(d Dir) {
		f, err := d.Create(fpath)
		if err != nil {
			t.Fatal("failed to create file: ", err)
		}

		_, err = f.Write([]byte("content"))
		if err != nil {
			f.Close()
			t.Fatal("failed to write into the file: ", err)
		}
		f.Close()

		fi, err := d.Stat(fpath)
		if err != nil {
			t.Fatal("failed to get file info: ", err)
		}

		if fi.Size != int64(len(content)) {
			t.Fatalf("expected file to be %d bytes, file was %d", len(content), fi.Size)
		}

		fr, err := d.Open(fpath)
		if err != nil {
			t.Fatal("failed to open file for reading: ", err)
		}
		defer fr.Close()
		c, err := ioutil.ReadAll(fr)
		if err != nil {
			t.Fatal("failed to read from file: ", err)
		}

		if string(c) != content {
			t.Fatal("expected file to contain '%s', found '%s'", content, string(c))
		}
	})
}
