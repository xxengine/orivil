package orivil

const SvcI18nFilter = "orivil.I18nFilter"

type I18nFilter interface {

	FilterMsg(src string) (dst string)

	ViewSubDir() (dir string)
}