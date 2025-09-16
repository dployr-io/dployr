<?php

namespace App\Rules;

use Closure;
use Illuminate\Contracts\Validation\ValidationRule;

class RemoteRepo implements ValidationRule
{
    /**
     * Run the validation rule.
     *
     * @param  \Closure(string, ?string=): \Illuminate\Translation\PotentiallyTranslatedString  $fail
     */
    public function validate(string $attribute, mixed $value, Closure $fail): void
    {
        $normalized = strtolower($value);

        // Allow http(s), (www.), github.com/gitlab.com 
        $pattern = '/^(https?:\/\/)?(www\.)?(github\.com|gitlab\.com)\/[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/';
    
        if (!preg_match($pattern, $normalized)) {
            $fail("Remote repository must be a valid github or gitlab url");
        }
    }
}
