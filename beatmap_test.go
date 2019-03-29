package osu

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func testSingle(filename string) func(*testing.T) {
	return func(t *testing.T) {
		f, err := os.Open("./test/" + filename)
		if err != nil {
			t.Errorf("failed to locate file '%s'", filename)
		}

		beatmap, err := ParseBeatmap(f)
		if err != nil {
			t.Errorf("failed to parse file '%s': %+v", filename, err)
			return
		}

		var buf bytes.Buffer
		err = beatmap.Serialize(&buf)
		if err != nil {
			t.Errorf("failed to serialize: %+v", err)
			return
		}

		t.Errorf("unserialized: %+v", buf.String())
	}
}

func TestSerialization(t *testing.T) {
	files, err := ioutil.ReadDir("./test")
	if err != nil {
		log.Fatal(err)
	}
	if testing.Short() {
		// shuffle to get random beatmaps
		rand.Shuffle(len(files), func(i, j int) {
			files[i], files[j] = files[j], files[i]
		})
	}

	for i, file := range files {
		if !strings.HasSuffix(file.Name(), ".osu") {
			continue
		}
		if i > 5 && testing.Short() {
			break
		}

		t.Run(fmt.Sprintf("test%d", i), testSingle(file.Name()))
	}
}
