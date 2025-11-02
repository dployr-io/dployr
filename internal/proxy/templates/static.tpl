{{.Domain}} {
    root * {{.App.Root}}
    file_server
    encode gzip
    
    log {
        output file {{.LogFile}}
        format json
    }
}