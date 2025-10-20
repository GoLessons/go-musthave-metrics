package dotforbidden

import (
	. "log"
	. "os"
)

func f() {
	Exit(1)    // want "вызов os.Exit вне функции main пакета main"
	Fatal("x") // want "вызов log.Fatal вне функции main пакета main"
}
