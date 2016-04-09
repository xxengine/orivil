package i18n

import (
	"github.com/orivil/event"
	"github.com/orivil/helper"
	"github.com/orivil/orivil"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	// listener name
	LsnI18n = "i18n.Listener"

	// i18n dir name
	i18nDir = "i18n"
)

// Listener listening the server event "EvtConfigProvider", if config option
// "auto_generate_files" is "true", it will generate not exist i18n
// view files
type Listener struct{}

func (this *Listener) RegisterService(s *orivil.Server) {}

func (this *Listener) RegisterRoute(s *orivil.Server) {}

func (this *Listener) RegisterMiddle(s *orivil.Server) {}

func (this *Listener) BootProvider(s *orivil.Server) {}

// when event "EvtConfigProvider" triggered, this function will be run
func (this *Listener) ConfigServer(s *orivil.Server) {
	if Config.Auto_generate_files {
		// auto generate I18n view files
		actions := s.RContainer.GetActions()
		for bundle, _ := range actions {
			bundleDir := filepath.Join(orivil.DirBundle, bundle)
			generateViewFiles(Config.DefaultLang, bundleDir, Config.Languages)
		}

		// auto generate I18n messages
		msgCacheDir := filepath.Join(orivil.DirConfig, "i18n_msgs")
		var langs []string
		for _, lang := range Config.Languages {
			langs = append(langs, lang)
		}
		orivil.I18n.Init(msgCacheDir, Config.Languages[Config.DefaultLang], langs)
		orivil.UpdateI18nConfig()
	}
}

// implement event.Listener interface to subscribe event
func (this *Listener) GetSubscribe() (name string, subscribes []event.Subscribe) {
	name = LsnI18n
	p := 500
	subscribes = []event.Subscribe{
		{
			Name:     orivil.EvtConfigProvider,
			Priority: p,
		},
	}
	return
}

// generateViewFiles 生成不存在的 view 文件
func generateViewFiles(defalutLang, bundleDir string, langs map[string]string) {
	for lang, shortName := range langs {
		if lang != defalutLang {
			generateFiles(bundleDir, shortName)
		}
	}
}

func generateFiles(bundleDir, lang string) {

	viewDir := filepath.Join(bundleDir, "view")
	if !helper.IsExist(viewDir) {
		return
	}
	i18nAbsDir := filepath.Join(viewDir, i18nDir)
	dstDir := filepath.Join(i18nAbsDir, lang)
	allSubDirs, err := helper.GetAllSubDirs(viewDir)
	if err != nil {
		panic(err)
	}

	// 收集所有需要拷贝的目录以及目标目录
	var srcDirs = make([]string, 1)
	var dstDirs = make([]string, 1)
	srcDirs[0] = viewDir
	dstDirs[0] = dstDir
	for _, dir := range allSubDirs {
		if !strings.HasPrefix(dir, i18nAbsDir) {
			// 排除 i18n 下的目录
			srcDirs = append(srcDirs, dir)
			dstSubDir := filepath.Join(dstDir, strings.TrimPrefix(dir, viewDir))
			dstDirs = append(dstDirs, dstSubDir)
		}
	}

	for index, srcDir := range srcDirs {
		dstDir = dstDirs[index]
		err := copyDir(srcDir, dstDir)
		if err != nil {
			panic(err)
		}
	}
}

func copyFile(srcName, dstName string) (err error) {

	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	fileInfo, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, fileInfo.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	// if target file exist, do not copy it
	if fileInfo, err := dst.Stat(); err != nil {
		return err
	} else if fileInfo.Size() > 0 {
		return nil
	}

	_, err = io.Copy(dst, src)
	return err
}

func copyDir(srcDir, dstDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	dirInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dstDir, dirInfo.Mode()); err != nil {
		return err
	}

	for _, file := range files {
		fileName := file.Name()
		srcFile := filepath.Join(srcDir, fileName)
		dstFile := filepath.Join(dstDir, fileName)
		if !file.IsDir() {
			if err := copyFile(srcFile, dstFile); err != nil {
				return err
			}
		}
	}
	return nil
}
