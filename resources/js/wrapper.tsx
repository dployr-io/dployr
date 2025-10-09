import { useFlashToast } from '@/hooks/use-flash-toast';
import { ReactNode } from 'react';
import { Toaster } from 'sonner';

export default function Wrapper({ children }: { children: ReactNode }) {
    useFlashToast(); // auto-handles flash props on navigation

    return (
        <>
            {children}
            <Toaster />
        </>
    );
}
