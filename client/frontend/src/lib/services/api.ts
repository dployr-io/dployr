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
} from '../../../wailsjs/go/main/App.js';
import { types } from '../../../wailsjs/go/models';
import { getFromLocalStorage } from '../../../src/utils/localStorage';


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
  
  async signOut() {
    // return await SignOut();
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
