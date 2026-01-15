http://{{.Domain}}, https://{{.Domain}} {
	reverse_proxy {{.App.Upstream}} {
		header_up Host {upstream_hostport}
		header_up X-Real-IP {remote_host}
		header_up X-Forwarded-For {remote_host}
		header_up X-Forwarded-Proto {scheme}
	}
	
	log {
		output file {{.LogFile}}
		format json
	}
	
	encode gzip
}