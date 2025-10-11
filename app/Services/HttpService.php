<?php

namespace App\Services;

class HttpService
{
    /**
     * Make an HTTP request with the given method, URL, headers, and context.
     *
     * @param  string  $method  The HTTP method (GET, POST, PUT, PATCH, DELETE).
     * @param  string  $url  The URL to send the request to.
     * @param  array  $headers  An associative array of headers to include in the request.
     * @param  string  $context  A string describing the context of the request (for error messages).
     * @return array The JSON-decoded response body.
     *
     * @throws \RuntimeException If the HTTP request fails.
     * @throws \InvalidArgumentException If the HTTP method is unsupported.
     */
    public static function makeRequest(string $method, string $url, array $headers = [], ?string $context = 'request')
    {
        $defaultHeaders = [
            'Accept' => 'application/json',
        ];

        $allHeaders = array_merge($defaultHeaders, $headers);

        $http = \Http::withHeaders($allHeaders);

        switch (strtolower($method)) {
            case 'get':
                $response = $http->get($url);
                break;
            case 'post':
                $response = $http->post($url);
                break;
            case 'put':
                $response = $http->put($url);
                break;
            case 'patch':
                $response = $http->patch($url);
                break;
            case 'delete':
                $response = $http->delete($url);
                break;
            default:
                throw new \InvalidArgumentException("Unsupported HTTP method: {$method}");
        }

        if (! $response->successful()) {
            $msg = $response->json('message') ?? $response->body();
            throw new \RuntimeException("Http request failed on {$context}: {$url} ({$msg})", 1);
        }

        return $response->json();
    }
}
