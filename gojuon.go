package main

import "strings"

var gojuonRows = []string{
	"あいうえお", "かきくけこ", "さしすせそ", "たちつてと", "なにぬねの",
	"はひふへほ", "まみむめも", "やゆよ", "らりるれろ", "わをん",
}

var gojuon = []rune(strings.Join(gojuonRows, ""))

var katakanaRows = []string{
	"アイウエオ", "カキクケコ", "サシスセソ", "タチツテト", "ナニヌネノ",
	"ハヒフヘホ", "マミムメモ", "ヤユヨ", "ラリルレロ", "ワヲン",
}

var katakana = []rune(strings.Join(katakanaRows, ""))
