import { StrictMode } from 'react';
import { QueryClientProvider } from '@tanstack/react-query';
import { createRoot } from 'react-dom/client';
import { Toast } from '@heroui/react';
import { queryClient } from './api/client';
import App from './App.tsx';
import './index.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <Toast.Provider placement="bottom end" />
      <App />
    </QueryClientProvider>
  </StrictMode>
);
