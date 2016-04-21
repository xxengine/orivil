package orivil

type redirect struct {
	url  string
	code int
}

type end struct {}

func Return() {
	panic(&end{})
}

func RedirectCode(url string, code int) {
	panic(&redirect{url: url, code: code})
}

func Redirect(url string) {
	panic(&redirect{url: url, code: 302})
}
