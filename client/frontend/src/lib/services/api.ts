import { get } from 'svelte/store';
import { 
  SignIn, 
  GetCurrentUser, 
  SignOut, 
  GetDeployments, 
  GetProjects, 
  GetLogs, 
  GetDomains,
  NewConsole
} from '../../../wailsjs/go/main/App.js';
import { types } from '../../../wailsjs/go/models';

export const authService = {
  async signIn(provider: string) {
    return await SignIn(provider.toLowerCase());
  },
  
  async getCurrentUser() {
    const user = await GetCurrentUser();
    // Return null if no user, otherwise create proper User instance
    return user ? types.User.createFrom(user) : null;
  },
  
  async signOut() {
    return await SignOut();
  }
};

export const dataService = {
  async getDeployments() {
    const deploymentData = await GetDeployments();
    return deploymentData.map((d: any) => types.Deployment.createFrom(d));
  },
  
  async getProjects() {
    const projectData = await GetProjects();
    return projectData.map((p: any) => types.Project.createFrom(p));
  },

  async getLogs() {
    const logData = await GetLogs();
    return logData.map((p: any) => types.LogEntry.createFrom(p));
  },

  async getDomains() {
    const domainsData = await GetDomains();
    return domainsData.map((p: any) => types.Domain.createFrom(p));
  },

  async newConsole() {
    const consoleData = await NewConsole();
    return types.Console.createFrom(consoleData);
  }
};
