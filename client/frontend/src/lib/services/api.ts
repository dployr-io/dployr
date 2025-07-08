import { SignIn, GetCurrentUser, SignOut, GetDeployments, GetProjects } from '../../../wailsjs/go/main/App.js';
import { main } from '../../../wailsjs/go/models';

export const authService = {
  async signIn(provider: string) {
    return await SignIn(provider.toLowerCase());
  },
  
  async getCurrentUser() {
    const user = await GetCurrentUser();
    // Return null if no user, otherwise create proper User instance
    return user ? main.User.createFrom(user) : null;
  },
  
  async signOut() {
    return await SignOut();
  }
};

export const dataService = {
  async getDeployments() {
    const deploymentData = await GetDeployments();
    return deploymentData.map((d: any) => main.Deployment.createFrom(d));
  },
  
  async getProjects() {
    const projectData = await GetProjects();
    return projectData.map((p: any) => main.Project.createFrom(p));
  }
};