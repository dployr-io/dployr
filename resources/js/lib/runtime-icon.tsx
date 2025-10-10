import type { Runtime } from '@/types';
import { BiCustomize } from 'react-icons/bi';
import { FaDocker, FaJava, FaNodeJs, FaPython } from 'react-icons/fa';
import { FaGolang, FaServer } from 'react-icons/fa6';
import { SiDotnet, SiK3S, SiPhp, SiRubyonrails } from 'react-icons/si';

export const getRuntimeIcon = (runtime: Runtime) => {
    switch (runtime) {
        case 'static':
            return <FaServer size={14} />;
        case 'go':
            return <FaGolang size={22} />;
        case 'php':
            return <SiPhp size={20} />;
        case 'python':
            return <FaPython size={16} />;
        case 'node-js':
            return <FaNodeJs size={16} />;
        case 'ruby':
            return <SiRubyonrails size={18} />;
        case 'dotnet':
            return <SiDotnet size={22} />;
        case 'java':
            return <FaJava size={20} />;
        case 'docker':
            return <FaDocker size={16} />;
        case 'k3s':
            return <SiK3S />;
        case 'custom':
            return <BiCustomize size={16} />;
        default:
            return <BiCustomize size={16} />;
    }
};
