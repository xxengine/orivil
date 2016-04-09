# Orivil i18n bundle

-------------------


------------------

**如何使用:**

-------------------

* 将 i18n 目录放置于`项目目录/bundle`下
* 将 i18n.yml 配置文件放置于`项目目录/config`下
* 在 i18n.yml 文件中配置网站需要使用的语言
* 使用 i18n.MidDataSender 中间件将 i18n.yml 中的配置数据发送至模板文件
 
```
func (this *Controller) SetMiddle(bag *middle.Bag) {

    // 配置发送 i18n 数据的中间件
	bag.Set(i18n.MidDataSender).OnlyActions("Index")
}
```

* 在模板中接收数据

```
<ul>
    {{/* 显示当前语言 */}}
    <li>{{$.currentLang}}</li>
    {{range $lang, $shortName := .langs}}
       
        {{/* 如果不是当前语言则加上超链接 */}}
        {{if ne $lang $.currentLang}}
        
            {{/* 该地址可以用 Ajax 发送请求, 请求成功后再刷新页面就可以切换语言了 */}}
            <li><a href="/setlang/{{$lang}}">{{$lang}}</a></li>
        {{end}}
    {{end}}
</ul>
```

* 注册 i18n bundle 到你的 Orivil 项目中
* 启动 server 就可以在各个 bundle 的 view 文件目录中生成对应语言的模板文件
* 翻译对应语言的模板文件