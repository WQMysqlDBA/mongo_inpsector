package Output

type logPrint interface {
	printmsg
	writer
	trun
	simplewrite
}

type printmsg interface {
	printmessage()
}
type writer interface {
	writeinspectorfile(f string)
}

type trun interface {
	truncateFile(f string)
}
type simplewrite interface {
	writefile(f string)
}
