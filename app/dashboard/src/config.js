// dashboard-temp/src/config.js
// Fixed configuration to ensure proper API endpoints

const DashboardConfig = {
    // Version
    version: 'v0.1.2',

    // API Configuration
    api: {
        // Base URL for API endpoints (empty string means use relative URLs)
        baseUrl: '',

        // Endpoint paths - make sure these exactly match your backend routes
        endpoints: {
            reports: '/api/reports',
            reportDetails: '/api/reports/{id}',
            downloadReport: '/api/reports/{id}/download/{format}',
            runHealthCheck: '/api/reports/cluster-health/run',
            operatorStatus: '/api/status',
            health: '/api/health'
        },

        // Request timeout in milliseconds (increased for better reliability)
        timeout: 60000,

        // Auto-refresh interval in milliseconds (2 minutes to reduce load)
        refreshInterval: 120000,

        // Max retry attempts for API calls
        maxRetries: 2,

        // Retry delay in milliseconds
        retryDelay: 1000
    },

    // Dashboard UI Configuration
    ui: {
        // Default tab to show when dashboard loads
        defaultTab: 'overview',

        // Enable/disable real-time updates
        enableRealTimeUpdates: true,

        // Enable debug mode
        debug: true, // Set to true to see more console output

        // Colors for status indicators
        colors: {
            // Status colors
            status: {
                healthy: '#10B981', // green
                warning: '#F59E0B', // amber
                critical: '#EF4444', // red
                unknown: '#6B7280'  // gray
            },

            // Chart colors
            chart: {
                passing: '#10B981', // green
                warning: '#F59E0B', // amber
                critical: '#EF4444', // red
                total: '#3B82F6'    // blue
            },

            // Background colors
            background: {
                passing: '#ECFDF5',
                warning: '#FFFBEB',
                critical: '#FEF2F2',
                total: '#EFF6FF'
            }
        },

        // Category names in preferred display order
        categories: [
            'Cluster Config',
            'Applications',
            'Security',
            'Storage',
            'Networking',
            'Performance',
            'Op-Ready'
        ],

        // Animation settings
        animations: {
            // Chart animation duration
            chartDuration: 500,
            // Enable/disable animations
            enabled: true
        },

        // Timeouts
        timeouts: {
            // How long to show success messages
            success: 3000,
            // How long to show error messages
            error: 5000
        }
    },

    // Export Configuration
    export: {
        // Available export formats
        formats: ['adoc', 'pdf', 'zip'],

        // Default format
        defaultFormat: 'pdf'
    },

    // Environment settings
    env: {
        // Current environment (treat as development for debugging)
        current: 'development',
        // Is production
        isProduction: false,
        // Is development
        isDevelopment: true,
        // Enable mock data for development
        useMockData: false
    }
};

export default DashboardConfig;