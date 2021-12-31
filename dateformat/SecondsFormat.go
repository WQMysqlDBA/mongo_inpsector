package dateformat

const (
	_Oneday     = 24
	_OneHour    = 3600
	_OneSeconds = 60
)

func ResolveTime(seconds int) (day, hour, min, sec int) {
	d := seconds / (_Oneday * _OneHour)
	h := (seconds - d*_OneHour*_Oneday) / _OneHour
	m := (seconds - d*_Oneday*_OneHour - h*_OneHour) / _OneSeconds
	s := seconds - d*_Oneday*_OneHour - h*_OneHour - m*_OneSeconds
	return d, h, m, s
}
