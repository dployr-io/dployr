// src/lib/utils/localStorage.ts

/**
 * Save a value to localStorage under the given key.
 * Serializes to JSON before storing.
 */
export function saveToLocalStorage<T>(key: string, value: T): void {
  try {
    const json = JSON.stringify(value);
    localStorage.setItem(key, json);
  } catch (err) {
    console.error(`Unable to save “${key}” to localStorage:`, err);
  }
}

/**
 * Retrieve a value from localStorage by key.
 * Parses the JSON back into the original type, or returns null if missing / invalid.
 */
export function getFromLocalStorage<T>(key: string): T | null {
  try {
    const json = localStorage.getItem(key);
    if (!json) return null;
    return JSON.parse(json) as T;
  } catch (err) {
    console.error(`Unable to read “${key}” from localStorage:`, err);
    return null;
  }
}

