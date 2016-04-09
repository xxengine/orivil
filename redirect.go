package orivil

type redirect struct {
	url  string
	code int
}

func RedirectCode(url string, code int) {
	panic(&redirect{url: url, code: code})
}

func Redirect(url string) {
	panic(&redirect{url: url, code: 302})
}
