package color

import "fmt"

const (
	Reset = iota
	Bold
	Fuzzy
	Italic
	Underscore
	Blink
	FastBlink
	Reverse
	Concealed
	Strikethrough
)

const (
	Black = iota + 30
	Red
	Green
	Yellow
	Blue
	Pink
	Cyan
	Gray

	White   = 97
	Unknown = 999
)

var colorMap = map[int]string{
	Bold:    "bold",
	Black:   "black",
	Red:     "red",
	Green:   "green",
	Yellow:  "yellow",
	Blue:    "blue",
	Pink:    "pink",
	Cyan:    "cyan",
	Gray:    "gray",
	White:   "white",
	Unknown: "unknown",
}

func SetColor(text string, conf, bg, color int) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, conf, bg, color, text, 0x1B)
}

func BoldText(s string) string {
	return SetColor(s, 0, 0, Bold)
}

func BlackText(s string) string {
	return SetColor(s, 0, 0, Black)
}

func RedText(s string) string {
	return SetColor(s, 0, 0, Red)
}

func GreenText(s string) string {
	return SetColor(s, 0, 0, Green)
}

func YellowText(s string) string {
	return SetColor(s, 0, 0, Yellow)
}

func BlueText(s string) string {
	return SetColor(s, 0, 0, Blue)
}

func PinkText(s string) string {
	return SetColor(s, 0, 0, Pink)
}

func CyanText(s string) string {
	return SetColor(s, 0, 0, Cyan)
}

func GrayText(s string) string {
	return SetColor(s, 0, 0, Gray)
}

func WhiteText(s string) string {
	return SetColor(s, 0, 0, White)
}

func PrintBold(s string) {
	println(BoldText(s))
}

func PrintBlack(s string) {
	println(BlackText(s))
}

func PrintRed(s string) {
	println(RedText(s))
}

func PrintGreen(s string) {
	println(GreenText(s))
}

func PrintYellow(s string) {
	println(YellowText(s))
}

func PrintBlue(s string) {
	println(BlueText(s))
}

func PrintPink(s string) {
	println(PinkText(s))
}

func PrintCyan(s string) {
	println(CyanText(s))
}

func PrintGray(s string) {
	println(GrayText(s))
}

func PrintWhite(s string) {
	println(WhiteText(s))
}

func CodeReason(code int) string {
	v, ok := colorMap[code]
	if !ok {
		v = colorMap[Unknown]
	}
	return v
}

func ColorToCode(s string) int {
	for k := range colorMap {
		if colorMap[k] == s {
			return k
		}
	}
	return Unknown
}
