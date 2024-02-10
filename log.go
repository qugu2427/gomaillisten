package main

type LogLevel uint8

const (
	Success LogLevel = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

type LogHandler = func(LogLevel, string)
