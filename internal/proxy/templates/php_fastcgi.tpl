{{.Domain}} {
	{{if .App.Root}}
	root * {{.App.Root}}
	{{else}}
	root * /var/www/{{.Domain}}
	{{end}}
	
	php_fastcgi {{.App.Upstream}} {
		index index.php
	}
	
	file_server
	
	header {
		X-Content-Type-Options nosniff
		X-Frame-Options DENY
		X-XSS-Protection "1; mode=block"
	}
	
	log {
		output file {{.LogFile}}
		format json
	}
	
	encode gzip
}