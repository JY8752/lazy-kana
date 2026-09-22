package main

import "strings"

var gojuonRows = []string{
	"あいうえお", "かきくけこ", "さしすせそ", "たちつてと", "なにぬねの",
	"はひふへほ", "まみむめも", "やゆよ", "らりるれろ", "わをん",
}

var gojuon = []rune(strings.Join(gojuonRows, ""))
