package paniccases

func f() {
	panic("oops") // want "использование panic запрещено"
}
