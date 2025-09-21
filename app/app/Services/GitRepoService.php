<?php

namespace App\Services;

use App\Services\RemoteProviderServices\GithubService;
use App\DTOs\ApiResponse;
use App\Enums\RemoteType;

class GitRepoService extends GithubService
{
    /**
     * Format url to a standard format.
     * e.g http://foo.bar -> https://foo.bar |
     *     https://www.foo.bar -> https://foo.bar |
     *     foo.bar -> https://foo.bar
     * 
     * @param string $url
     * @return string|null
     */
    protected static function formatUrl(string $url): ?string
    {
        $url = strtolower(trim($url));

        // Maintain http/https
        if (!preg_match('#^https?://#', $url)) {
            $url = 'https://' . ltrim($url, '/');
        }

        // Remove www. after scheme
        $url = preg_replace('#^https?://www\.#', 'https://', $url);

        // Validate URL
        if (!filter_var($url, FILTER_VALIDATE_URL)) {
            return null; // invalid URL
        }

        $url = rtrim($url, '/');

        return $url;
    }

    
    /**
     * Extract info about a remote repository
     * 
     * @param string $url
     * @return array{name: string, remote: mixed|string, repository: array|null}
     */
    public static function parse(string $url): ?array
    {
        $formattedUrl = self::formatUrl($url);

        $parts = parse_url($formattedUrl);

        if (!isset($parts['host'], $parts['path'])) {
            return null;
        }

        $segments = explode('/', ltrim($parts['path'], '/'));

        if (count($segments) < 2) {
            return null;
        }

        return [
            'name'       => $segments[0],
            'repository' => preg_replace('/\.git$/', '', $segments[1]),
            'provider'     => $parts['host'],
        ];
    }    

    protected function search(string $name, string $repository, string $provider): ApiResponse
    {
        if (self::getRemoteType($provider) == RemoteType::Github) 
        {
            return parent::search($name, $repository, $provider);   
        }

        return new ApiResponse(false, [], "Something went wrong");
    }

    public function searchRepo(string $url): ApiResponse
    {
        $formattedUrl = self::formatUrl($url);

        $parsed = self::parse($formattedUrl);

        return $this->search($parsed['name'], $parsed['repository'], $parsed['provider']);
    }

    public function getLatestCommitMessage(string $name, string $repository, string $provider): array
    {
        if (self::getRemoteType($provider) == RemoteType::Github) 
        {
            return parent::getLatestCommitMessage($name, $repository, $provider);   
        }
        
        return [];
    }
}
