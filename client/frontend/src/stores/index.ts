import { writable } from 'svelte/store';
import type { User, Project, Deployment, LogEntry, AppState, Account, ThemeType, ViewMode, Domain } from '../types/index';
import { main } from '../../wailsjs/go/models';

// Theme store
export const currentTheme = writable<ThemeType>('system');
export const isDarkMode = writable<boolean>(false);

// Auth store
export const currentUser = writable<User | null>(null);
export const isAuthenticating = writable<boolean>(false);

// Onboarding stores
export const selectedOptions = writable<string[]>([]);
export const discoveryOptions = writable<string[]>([]);
export const discoveryOther = writable<string>('');
export const appStage = writable<string>('');
export const signInProvider = writable<string>('');
export const currentPage = writable<number>(0);
export const isTransitioning = writable<boolean>(false);

// UI stores
export const viewMode = writable<ViewMode>('grid');
export const sidebarWidth = writable<number>(640);
export const isResizing = writable<boolean>(false);
export const showFilterDropdown = writable<boolean>(false);
export const showProjectDropdown = writable<boolean>(false);
export const showAccountDropdown = writable<boolean>(false);

// App state
export const appState = writable<AppState>({
  selectedSection: 'Deployments',
  selectedProjectId: null,
  selectedDeploymentId: null
});

// Data stores
export const projects = writable<Project[]>([]);
export const domains = writable<Domain[]>([]);
export const deployments = writable<Deployment[]>([]);
export const logs = writable<LogEntry[]>([]);

export const selectedProject = writable<Project | null>(null);
export const selectedDomain = writable<Domain | null>(null);
export const selectedDeployment = writable<Deployment | null>(null);
export const selectedLog = writable<LogEntry | null>(null);

// Account management
export const accounts = writable<Account[]>([
  { name: "zeipo-ai", avatar: "" },
  { name: "username", avatar: "" }
]);
export const selectedAccount = writable<Account>({ name: "zeipo-ai", avatar: "" });