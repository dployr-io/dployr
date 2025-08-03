import type { User } from "src/types";
import { addToast } from "../../../src/stores/toastStore";
import {
  SignIn,
  GetDeployments,
  GetProjects,
  GetLogs,
  GetDomains,
  NewConsole,
  VerifyMagicCode,
  CreateProject,
} from "../../../wailsjs/go/main/App.js";
import { models } from "../../../wailsjs/go/models";
import { getFromLocalStorage } from "../../../src/utils/localStorage";
import type { get } from "svelte/store";

// Authentication initialization interfaces
export interface AuthInitializationResult {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  host: string | null;
}

export const authService = {
  async signIn(host: string, email: string, name: string) {
    return await SignIn(host, email, name);
  },

  async verifyMagicCode(host: string, email: string, code: string) {
    return await VerifyMagicCode(host, email, code);
  },

  async getCurrentUser() {
    return getFromLocalStorage<User>("user");
  },

  async getToken() {
    return getFromLocalStorage<string>("token");
  },

  async getHost() {
    return getFromLocalStorage<string>("host");
  },

  async signOut() {
    // return await SignOut();
  },

  /**
   * Initialize authentication state from localStorage
   * Restores user session data and handles errors gracefully
   */
  async initializeAuthState(): Promise<AuthInitializationResult> {
    try {
      // Attempt to restore authentication data from localStorage
      const user = getFromLocalStorage<User>("user");
      const token = getFromLocalStorage<string>("token");
      const host = getFromLocalStorage<string>("host");

      // Validate that we have the minimum required data for authentication
      const isAuthenticated = !!(user && token && host);

      // If we have partial data but not complete authentication, log a warning
      if ((user || token || host) && !isAuthenticated) {
        console.warn("Incomplete authentication data found in localStorage. Some data may be corrupted.");
      }

      // Validate user object structure if it exists
      if (user && typeof user === 'object') {
        // Basic validation - ensure user has expected properties
        if (!user.id && !user.email && !user.name) {
          console.warn("User object appears to be corrupted, clearing authentication data");
          return {
            isAuthenticated: false,
            user: null,
            token: null,
            host: null
          };
        }
      }

      // Validate token format if it exists
      if (token && typeof token !== 'string') {
        console.warn("Token appears to be corrupted, clearing authentication data");
        return {
          isAuthenticated: false,
          user: null,
          token: null,
          host: null
        };
      }

      // Validate host format if it exists
      if (host && typeof host !== 'string') {
        console.warn("Host appears to be corrupted, clearing authentication data");
        return {
          isAuthenticated: false,
          user: null,
          token: null,
          host: null
        };
      }

      return {
        isAuthenticated,
        user,
        token,
        host
      };

    } catch (error) {
      // Handle any errors during initialization
      console.error("Error during authentication initialization:", error);

      // Return unauthenticated state on error
      return {
        isAuthenticated: false,
        user: null,
        token: null,
        host: null
      };
    }
  },
};

export const dataService = {
  async getDeployments(host: string, token: string) {
    const deploymentData = await GetDeployments(host, token);
    return deploymentData.map((d: any) => models.Deployment.createFrom(d));
  },

  async getProjects(host: string, token: string) {
    const projectData = await GetProjects(host, token);
    return projectData.map((p: any) => models.Project.createFrom(p));
  },

  async getLogs(host: string, token: string) {
    const logData = await GetLogs(host, token);
    return logData.map((p: any) => models.LogEntry.createFrom(p));
  },

  async getDomains(host: string, token: string) {
    const domainsData = await GetDomains(host, token);
    return domainsData.map((p: any) => models.Domain.createFrom(p));
  },
};

export const consoleService = {
  async newConsole() {
    const consoleData = await NewConsole();
    return models.Console.createFrom(consoleData);
  },
};

export const projectService = {
  async createProject(
    host: string,
    token: string,
    payload: Record<string, string>
  ) {
    return await CreateProject(host, token, payload);
  },
};
