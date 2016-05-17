# Orivil Web Framework

Fast & Simple & Powerful Go Web Framework. Inspired by [Symfony](http://symfony.com/) and [Laravel](https://laravel.com/).


## Version

```
v1.0
```

## Overview

* Use service container to manage services.
* Fantastic view file compiler.
* Semantic management of middleware.
* Automatic generate controller routes.
* Could cache view file, could cache services, support memory session.
* Made up of components, every user could be a contributor.
* Automatic generate I18n files, including view file and configuration file.

## Install

```
go get -v gopkg.in/orivil/orivil.v1
```

## Run Example

```
cd $GOPATH/src/gopkg.in/orivil/orivil.v1/example_server/base_server

go run main.go
```

## Test Example

```
Browser visit: http://localhost:8080
```

## ApacheBench Test
#### Env

* OS: ubuntu 14.04 LTS
* CPU: Intel® Core™ i7-2600K
* GO: go1.6 linux/amd64
* DEBUG: false

#### Without Session

> ab -c 1000 -n 100000 http://localhost:8080/
>
```
Requests per second:    18693.95 [#/sec] (mean)
```

#### With Memory Session

> ab -c 1000 -n 100000 http://localhost:8080/set-session/ooorivil
>
```
Requests per second:    18150.50 [#/sec] (mean)
```

## Community

* [orivil.com](http://orivil.com/forum)
* [Gitter](https://gitter.im/orivil/orivil?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
* QQ群: 416628342

## Contributors

https://github.com/orivil/orivil/graphs/contributors


## License

Released under the [MIT License](https://github.com/orivil/orivil/blob/master/LICENSE).