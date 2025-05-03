// app/dashboard/src/Dashboard.js
import React, { useState } from 'react';
import UploadSection from './components/UploadSection';
import TabNavigation from './components/TabNavigation';
import OverviewTab from './components/OverviewTab';
import ExecutiveSummaryTab from './components/ExecutiveSummaryTab';
import RemediationTab from './components/RemediationTab';

const Dashboard = () => {
    const [activeTab, setActiveTab] = useState('overview');
    const [file, setFile] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [reportData, setReportData] = useState(null);

    // Function to handle file selection
    const handleFileChange = (e) => {
        const selectedFile = e.target.files[0];
        if (selectedFile && (selectedFile.name.endsWith('.adoc') || selectedFile.name.endsWith('.asciidoc'))) {
            setFile(selectedFile);
            setError(null);
        } else {
            setFile(null);
            setError('Please select an AsciiDoc (.adoc/.asciidoc) file');
        }
    };

    // Function to upload and analyze report using the server API
    const handleUpload = async () => {
        if (!file) {
            setError('Please select a file first');
            return;
        }

        setLoading(true);
        setError(null);

        try {
            // Create form data to send to server
            const formData = new FormData();
            formData.append('report', file);

            // Send the file to the server for parsing
            const response = await fetch('/api/parse-report', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`Server returned ${response.status}: ${response.statusText}`);
            }

            // Parse the JSON response
            const data = await response.json();

            // Format the overall score to be an integer if it's a whole number
            // or with 1 decimal place if it has a fractional part
            if (data.overallScore) {
                const score = parseFloat(data.overallScore);
                data.overallScore = Number.isInteger(score) ? Math.round(score) : parseFloat(score.toFixed(1));
            }

            // Check for empty arrays and initialize them if needed
            data.itemsRequired = data.itemsRequired || [];
            data.itemsRecommended = data.itemsRecommended || [];
            data.itemsAdvisory = data.itemsAdvisory || [];

            // Set the report data and switch to overview tab
            setReportData(data);
            setActiveTab('overview');

            console.log(`Successfully analyzed report: ${file.name}`);
        } catch (err) {
            console.error('Error processing file:', err);
            setError(`Failed to process the file: ${err.message || 'Unknown error'}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                {/* Header */}
                <div className="bg-white shadow rounded-lg p-4 mb-6">
                    <div className="flex md:flex-row flex-col justify-between items-center">
                        <div>
                            <h1 className="text-2xl font-bold text-gray-800">OpenShift Health Check Dashboard</h1>
                            <p className="text-gray-500">Upload and analyze your OpenShift health check reports</p>
                        </div>
                        <div className="mt-2 md:mt-0">
                            <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clipRule="evenodd" />
                                </svg>
                                {new Date().toLocaleDateString()}
                            </span>
                        </div>
                    </div>
                </div>

                {/* Upload Section */}
                <UploadSection
                    file={file}
                    loading={loading}
                    error={error}
                    handleFileChange={handleFileChange}
                    handleUpload={handleUpload}
                />

                {reportData && (
                    <>
                        {/* Tab Navigation */}
                        <TabNavigation activeTab={activeTab} setActiveTab={setActiveTab} />

                        {/* Content Area */}
                        {activeTab === 'overview' && <OverviewTab reportData={reportData} />}
                        {activeTab === 'executive' && <ExecutiveSummaryTab reportData={reportData} />}
                        {activeTab === 'remediation' && <RemediationTab reportData={reportData} />}
                    </>
                )}

                {/* Footer Info */}
                <div className="text-center text-xs text-gray-500 mt-8 pb-4">
                    <p>OpenShift Health Check Dashboard • Analyze and visualize health check reports</p>
                    <p>Upload ADOC files to generate insights and remediation recommendations</p>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;