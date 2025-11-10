package forbidden

import (
	"log"
	"os"
)

func f() {
	os.Exit(1)         // want "вызов os.Exit вне функции main пакета main"
	log.Fatal("fatal") // want "вызов log.Fatal вне функции main пакета main"
}
