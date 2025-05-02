// dashboard-temp/src/api-service.js
// Enhanced with debugging and more robust error handling

import config from './config';
import Logger from './utils/logger';

/**
 * Enhanced fetch with debugging and consistent error handling
 * @param {string} url - The URL to fetch from
 * @param {Object} options - Fetch options
 * @param {number} retries - Number of retries (default: 2)
 * @returns {Promise<any>} - Response data
 */
async function enhancedFetch(url, options = {}, retries = 2) {
    // Add console logging for immediate visibility during debugging
    console.log(`API Request: ${options.method || 'GET'} ${url}`);

    try {
        const startTime = Date.now();
        Logger.apiRequest(options.method || 'GET', url);

        // Add basic headers and timeout
        const requestOptions = {
            ...options,
            headers: {
                'Accept': 'application/json',
                'Cache-Control': 'no-cache',
                ...(options.headers || {})
            }
        };

        // Add controller for timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), config.api.timeout);
        requestOptions.signal = controller.signal;

        // Perform fetch
        console.log("Fetch options:", JSON.stringify(requestOptions));
        const response = await fetch(url, requestOptions);
        clearTimeout(timeoutId);

        const timing = Date.now() - startTime;
        Logger.apiResponse(options.method || 'GET', url, response.status, timing);
        console.log(`API Response: ${response.status} - ${response.statusText} (${timing}ms)`);

        if (!response.ok) {
            console.error(`API error: ${response.status} - ${response.statusText}`);
            // Try to get error details from response
            let errorDetails;
            try {
                errorDetails = await response.text();
                console.error("Error details:", errorDetails);
            } catch (e) {
                console.error("Could not read error details:", e);
            }

            throw new Error(`API error: ${response.status} - ${errorDetails || response.statusText}`);
        }

        // Try to parse as JSON, fall back to text if not JSON
        const contentType = response.headers.get('content-type');
        let responseData;

        if (contentType && contentType.includes('application/json')) {
            responseData = await response.json();
            console.log("Response data (JSON):", responseData);
        } else {
            responseData = await response.text();
            console.log("Response data (Text):", responseData.substring(0, 200) + (responseData.length > 200 ? '...' : ''));
            // Try to parse as JSON anyway if it looks like JSON
            if (responseData.trim().startsWith('{') || responseData.trim().startsWith('[')) {
                try {
                    responseData = JSON.parse(responseData);
                    console.log("Parsed text response as JSON");
                } catch (e) {
                    console.log("Could not parse text as JSON, returning as text");
                }
            }
        }

        return responseData;
    } catch (error) {
        // Handle specific errors
        if (error.name === 'AbortError') {
            console.error(`Request timeout: ${url}`);
            Logger.error(`Request timeout: ${url}`);
            throw new Error(`Request timeout: ${url}`);
        }

        // Add detailed logging
        console.error(`API error for ${url}:`, error);

        // Retry logic
        if (retries > 0) {
            console.log(`Retrying request to ${url}, ${retries} retries left`);
            Logger.warn(`Retrying request to ${url}, ${retries} retries left`);
            // Add exponential backoff
            const backoffDelay = config.api.retryDelay * Math.pow(2, config.api.maxRetries - retries);
            await new Promise(resolve => setTimeout(resolve, backoffDelay));
            return enhancedFetch(url, options, retries - 1);
        }

        Logger.apiError(options.method || 'GET', url, error);
        throw error;
    }
}

/**
 * Fetch all health check reports with improved error handling
 * @returns {Promise<Array>} Array of report objects
 */
export async function fetchReports() {
    try {
        // Make sure URL is properly formed
        const endpoint = config.api.baseUrl + config.api.endpoints.reports;
        console.log('Fetching reports from:', endpoint);
        Logger.info('Fetching reports from:', endpoint);

        return await enhancedFetch(endpoint);
    } catch (error) {
        Logger.error('Error fetching reports:', error);
        console.error('Error fetching reports:', error);
        // Return empty array instead of throwing to allow UI to handle gracefully
        return [];
    }
}

/**
 * Fetch detailed information for a specific report
 * @param {string} reportId - The ID of the report to fetch
 * @returns {Promise<Object>} Report details
 */
export async function fetchReportDetails(reportId) {
    try {
        const endpoint = config.api.baseUrl + config.api.endpoints.reportDetails.replace('{id}', reportId);
        console.log('Fetching report details from:', endpoint);
        Logger.info('Fetching report details from:', endpoint);

        return await enhancedFetch(endpoint);
    } catch (error) {
        Logger.error(`Error fetching report ${reportId}:`, error);
        console.error(`Error fetching report ${reportId}:`, error);
        throw error;
    }
}

/**
 * Triggers a new health check run
 * @returns {Promise<Object>} Response data
 */
export async function runHealthCheck() {
    try {
        const endpoint = config.api.baseUrl + config.api.endpoints.runHealthCheck;
        console.log('Triggering health check run:', endpoint);
        Logger.info('Triggering health check run:', endpoint);

        // For POST requests, make sure to send an empty body at minimum
        return await enhancedFetch(endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({}) // Empty JSON object for POST
        });
    } catch (error) {
        Logger.error('Error running health check:', error);
        console.error('Error running health check:', error);
        throw error;
    }
}

/**
 * Downloads a report in the specified format
 * @param {string} reportId - The ID of the report to download
 * @param {string} format - The format to download (adoc, pdf, zip)
 */
export function downloadReport(reportId, format) {
    try {
        const endpoint = config.api.baseUrl + config.api.endpoints.downloadReport
            .replace('{id}', reportId)
            .replace('{format}', format);

        console.log('Downloading report from:', endpoint);
        Logger.info('Downloading report from:', endpoint);

        // Use direct location change for downloads
        window.location.href = endpoint;
        return true;
    } catch (error) {
        Logger.error(`Error initiating download for report ${reportId} in format ${format}:`, error);
        console.error(`Error initiating download for report ${reportId} in format ${format}:`, error);
        throw error;
    }
}

/**
 * Downloads the latest report in the specified format
 * @param {string} format - The format to download (adoc, pdf, zip)
 * @returns {Promise<boolean>} Success status
 */
export async function downloadLatestReport(format) {
    try {
        console.log('Downloading latest report in format:', format);
        Logger.info('Downloading latest report in format:', format);

        const reports = await fetchReports();
        console.log('Fetched reports for download:', reports);

        if (!reports || reports.length === 0) {
            const error = new Error('No reports available');
            console.error(error);
            throw error;
        }

        // Sort by timestamp descending
        reports.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));

        // Get the latest report
        const latestReport = reports[0];
        console.log('Latest report:', latestReport);

        // Check if the format is available
        if (!latestReport.formats || !latestReport.formats.includes(format)) {
            const error = new Error(`Format ${format} not available for the latest report`);
            console.error(error);
            throw error;
        }

        // Download the report
        downloadReport(latestReport.id, format);
        return true;
    } catch (error) {
        Logger.error('Error downloading latest report:', error);
        console.error('Error downloading latest report:', error);
        throw error;
    }
}