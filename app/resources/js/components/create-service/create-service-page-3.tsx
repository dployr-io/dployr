import React, { useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';

interface Remote {
  id: string;
  name: string;
  repository: string;
  provider: string;
}

interface FormData {
  name: string;
  source: string;
  runtime: string;
  workingDir?: string | null;
  runCmd?: string | null;
  port?: number | null;
  domain?: string | null;
  dnsProvider?: string | null;
  remote?: Remote | null;
}

interface Props {
  formData: FormData;
}

export function CreateServicePage3({ formData }: Props = { 
  formData: {
    name: "my-awesome-app",
    source: "remote",
    runtime: "node-js",
    workingDir: "src",
    runCmd: "npm run start",
    port: 3000,
    domain: "myapp.example.com",
    dnsProvider: "cloudflare",
    remote: {
      id: "123",
      name: "john-doe",
      repository: "my-website",
      provider: "github"
    }
  }
}) {
  const [format, setFormat] = useState<'yaml' | 'json'>('yaml');

  const config = useMemo(() => {
    // Convert form data to clean config object
    const cleanConfig: any = {
      name: formData.name,
      source: formData.source,
      runtime: formData.runtime,
    };

    if (formData.workingDir) cleanConfig.workingDir = formData.workingDir;
    if (formData.runCmd) cleanConfig.runCmd = formData.runCmd;
    if (formData.port) cleanConfig.port = formData.port;
    if (formData.domain) cleanConfig.domain = formData.domain;
    if (formData.dnsProvider) cleanConfig.dnsProvider = formData.dnsProvider;
    if (formData.remote) cleanConfig.remote = formData.remote;

    return cleanConfig;
  }, [formData]);

  const yamlConfig = useMemo(() => {
    const toYaml = (obj: any, indent = 0): string => {
      const spaces = '  '.repeat(indent);
      let yaml = '';

      for (const [key, value] of Object.entries(obj)) {
        if (value === null || value === undefined) continue;
        
        yaml += `${spaces}${key}:`;
        
        if (typeof value === 'object' && value !== null) {
          yaml += '\n' + toYaml(value, indent + 1);
        } else if (typeof value === 'string') {
          yaml += ` "${value}"\n`;
        } else {
          yaml += ` ${value}\n`;
        }
      }
      return yaml;
    };

    return toYaml(config);
  }, [config]);

  const jsonConfig = useMemo(() => {
    return JSON.stringify(config, null, 2);
  }, [config]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(format === 'yaml' ? yamlConfig : jsonConfig);
    } catch (err) {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = format === 'yaml' ? yamlConfig : jsonConfig;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
  };

  return (
    <div className="grid items-start gap-6">
      <div className="rounded-lg bg-muted p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Configuration Blueprint</h3>
          <div className="flex items-center gap-2">
            <div className="flex bg-background rounded-md p-1">
              <Button
                variant={format === 'yaml' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setFormat('yaml')}
                className="h-7 px-3"
              >
                YAML
              </Button>
              <Button
                variant={format === 'json' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setFormat('json')}
                className="h-7 px-3"
              >
                JSON
              </Button>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleCopy}
              className="h-7 px-3"
            >
              Copy
            </Button>
          </div>
        </div>
        
        <div className="bg-background rounded border">
          <pre className="p-4 text-sm font-mono overflow-x-auto whitespace-pre-wrap">
            <code>
              {format === 'yaml' ? yamlConfig : jsonConfig}
            </code>
          </pre>
        </div>
        
        <p className="text-xs text-muted-foreground mt-2">
          This configuration can be saved as <code className="bg-background px-1 rounded text-xs">service.{format}</code> and used for deployment.
        </p>
      </div>
    </div>
  );
}
