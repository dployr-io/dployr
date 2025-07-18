import type { User } from 'src/types';
import { addToast } from '../../../src/stores/toastStore';
import { 
  SignIn, 
  GetDeployments, 
  GetProjects, 
  GetLogs, 
  GetDomains,
  NewConsole,
  VerifyMagicCode,
  CreateProject,
} from '../../../wailsjs/go/main/App.js';
import { models } from '../../../wailsjs/go/models';
import { getFromLocalStorage } from '../../../src/utils/localStorage';
import type { get } from 'svelte/store';


export const authService = {
  async signIn(host: string, email: string, name: string, password: string, privateKey: string) {
   return await SignIn(host, email, name, password, privateKey);
  },

  async verifyMagicCode(host: string, email: string, code: string) {
    return await VerifyMagicCode(host, email, code);
  },
  
  async getCurrentUser() {
    return await getFromLocalStorage<User>('user');
  },

  async getToken() {
    return await getFromLocalStorage<string>('token');
  },

  async getHost() {
    return await getFromLocalStorage<string>('host');
  },
  
  async signOut() {
    // return await SignOut();
  }
};

export const dataService = {
  async getDeployments() {
    const deploymentData = await GetDeployments();
    return deploymentData.map((d: any) => models.Deployment.createFrom(d));
  },
  
  async getProjects(host: string, token: string) {
    const projectData = await GetProjects(host, token);
    return projectData.map((p: any) => models.Project.createFrom(p));
  },

  async getLogs() {
    const logData = await GetLogs();
    return logData.map((p: any) => models.LogEntry.createFrom(p));
  },

  async getDomains() {
    const domainsData = await GetDomains();
    return domainsData.map((p: any) => models.Domain.createFrom(p));
  },
};

export const consoleService = {
  async newConsole() {
    const consoleData = await NewConsole();
    return models.Console.createFrom(consoleData);
  }
};

export const projectService = {
  async createProject(host: string, token: string, payload: Record<string, string>) {
    return await CreateProject(host, token, payload);
  },
};
