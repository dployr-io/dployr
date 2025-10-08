export const runtimes = ['static', 'go', 'php', 'python', 'node-js', 'ruby', 'dotnet', 'java', 'docker', 'k3s', 'custom'] as const;

export const dnsProviders = ['cloudflare', 'route53', 'namecheap', 'godaddy'] as const;

export const logLevels = ['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL', 'ALERT', 'EMERGENCY', 'NOTICE'] as const;
