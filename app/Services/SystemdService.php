<?php

namespace App\Services;

class SystemdService
{
    public function newService(string $name, $workingDir, $runCmd)
    {
        $serviceContent = <<<EOF
[Unit]
Description=$name
After=network.target

[Service]
Type=simple
User=dployr
WorkingDirectory=$workingDir
ExecStart=/bin/bash -lc '{$runCmd}'
EnvironmentFile=-$workingDir/.env
EnvironmentFile=-$workingDir/.env.secrets
Restart=always
StandardOutput=append:/home/dployr/logs/{$name}.log
StandardError=inherit

[Install]
WantedBy=multi-user.target
EOF;

        $config = escapeshellarg($serviceContent);

        $result = Cmd::execute("echo $config | sudo tee /etc/systemd/system/{$name}.service > /dev/null");
        if (! $result->successful) {
            throw new \RuntimeException("Failed to create systemd file {$name}.service");
        }

        $result = Cmd::execute('sudo systemctl daemon-reload');
        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }

        $result = Cmd::execute("sudo systemctl enable {$name}");
        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }

        $result = Cmd::execute("sudo systemctl start {$name}");
        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }
    }
}