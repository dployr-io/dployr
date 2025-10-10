import type { Project } from '@/types';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

export function useProjects() {
    const { data, isLoading } = useQuery<Project[]>({
        queryKey: ['projects'],
        queryFn: async () => {
            const response = await axios.get('/projects/fetch');
            return response.data;
        },
        staleTime: 60 * 1000, // Every minute
    });

    const defaultProject: Project | null = (() => {
        const storedProjectId = localStorage.getItem('current_project');
        if (isLoading) return null;
        if (storedProjectId && data!.length > 0) {
            const savedProject = data!.find((p) => p.id === storedProjectId);
            if (savedProject) return savedProject;
        }
        return data!.length > 0 ? data![0] : null;
    })();

    return {
        projects: data,
        isLoading,
        defaultProject,
    };
}
