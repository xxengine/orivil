<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Orivil Server</title>
</head>
<body>
<h1>Hello {{.say}}</h1>

{{if .getSession}}
<p>Get Last Session: {{.getSession}}</p>
{{end}}

{{if .setSession}}
<p>Set Session: {{.setSession}}</p>
{{end}}

</body>
</html>