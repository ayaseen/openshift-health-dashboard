// app/dashboard/src/index.js
import React from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import Dashboard from './Dashboard';

// Initialize the React app
const root = createRoot(document.getElementById('root'));
root.render(<Dashboard />);