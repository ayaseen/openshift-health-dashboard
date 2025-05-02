import React, { useState } from 'react';
import TabNavigation from './components/TabNavigation';
import UploadSection from './components/UploadSection';
import OverviewTab from './components/OverviewTab';
import ExecutiveSummaryTab from './components/ExecutiveSummaryTab';
import RemediationTab from './components/RemediationTab';

const Dashboard = () => {
    // State for managing tabs
    const [activeTab, setActiveTab] = useState('overview');

    // State for uploaded file and report data
    const [file, setFile] = useState(null);
    const [loading, setLoading] = useState(false);
    const [reportData, setReportData] = useState(null);
    const [error, setError] = useState(null);

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

    // Function to upload and analyze report
    const handleUpload = async () => {
        if (!file) {
            setError('Please select a file first');
            return;
        }

        setLoading(true);
        setError(null);

        try {
            // Create form data for upload
            const formData = new FormData();
            formData.append('report', file);

            // Send to server for parsing
            const response = await fetch('/api/parse-report', {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) {
                throw new Error(`Server error: ${response.statusText}`);
            }

            const data = await response.json();
            setReportData(data);
            setActiveTab('overview');
        } catch (err) {
            console.error('Error uploading report:', err);
            setError(`Failed to upload report: ${err.message}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                {/* Header */}
                <div className="bg-white shadow-sm rounded-lg p-4 mb-6">
                    <div className="flex justify-between items-center">
                        <div>
                            <h1 className="text-2xl font-bold text-gray-800">OpenShift Health Check Dashboard</h1>
                            <p className="text-gray-500">Upload and analyze your OpenShift health check reports</p>
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

                {/* Content Area */}
                {reportData && (
                    <>
                        <TabNavigation activeTab={activeTab} setActiveTab={setActiveTab} />

                        {activeTab === 'overview' && (
                            <OverviewTab reportData={reportData} />
                        )}

                        {activeTab === 'executive' && (
                            <ExecutiveSummaryTab reportData={reportData} />
                        )}

                        {activeTab === 'remediation' && (
                            <RemediationTab reportData={reportData} />
                        )}
                    </>
                )}
            </div>
        </div>
    );
};

export default Dashboard;