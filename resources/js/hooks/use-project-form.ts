import { useState } from 'react';
import { z } from 'zod';

const formSchema = z.object({
    name: z.string().min(3, 'Name with a minimum of three (3) characters is required'),
    description: z.string().optional(),
});

export function useProjectForm() {
    const [error, setError] = useState<string>('');
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');

    const validateForm = () => {
        const result = formSchema.safeParse({ name, description });

        if (!result.success) {
            const fieldErrors = result.error.flatten().fieldErrors;
            setError(fieldErrors.name?.[0] || fieldErrors.description?.[0] || 'Validation failed');
            return false;
        }

        setError('');
        return true;
    };

    const getFormData = () => {
        if (!validateForm()) return {};
        return { name, description };
    };

    return {
        name,
        description,
        validationError: error,
        setName,
        setDescription,
        getFormData,
    };
}
