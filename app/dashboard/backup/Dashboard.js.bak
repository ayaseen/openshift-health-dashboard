import React, { useState } from 'react';
import { getScoreColor, getScoreRating } from './utils/scoreUtils';

// Circular progress component for overall health
const CircularProgress = ({ value }) => {
    const percentage = value || 0;
    const radius = 70;
    const stroke = 14;
    const normalizedRadius = radius - stroke / 2;
    const circumference = normalizedRadius * 2 * Math.PI;
    const strokeDashoffset = circumference - (percentage / 100) * circumference;
    const rating = getScoreRating(percentage);
    const color = getScoreColor(percentage);

    return (
        <div className="relative flex items-center justify-center my-8">
            <svg height={radius * 2} width={radius * 2} className="transform -rotate-90">
                <circle
                    stroke="#e5e7eb"
                    fill="transparent"
                    strokeWidth={stroke}
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
                <circle
                    stroke={color}
                    fill="transparent"
                    strokeWidth={stroke}
                    strokeDasharray={`${circumference} ${circumference}`}
                    style={{ strokeDashoffset }}
                    strokeLinecap="round"
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
            </svg>
            <div className="absolute flex flex-col items-center justify-center">
                <span className="text-3xl font-bold">{percentage}%</span>
                <span className="text-sm text-gray-500">{rating}</span>
            </div>
        </div>
    );
};

// Progress bar component for category scores
const ProgressBar = ({ name, score }) => {
    const safeScore = score || 0;
    const color = getScoreColor(safeScore);
    return (
        <div className="mb-6">
            <div className="flex justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">{name}</span>
                <span className="text-sm font-medium text-gray-700">{safeScore}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div className="h-2.5 rounded-full" style={{ width: `${safeScore}%`, backgroundColor: color }}></div>
            </div>
        </div>
    );
};

// Action item box
const ActionBox = ({ count, label, color, bgColor }) => {
    return (
        <div className={`p-5 rounded-lg text-center ${bgColor}`}>
            <div className={`text-3xl font-bold ${color}`}>{count}</div>
            <div className={`text-sm ${color}`}>{label}</div>
        </div>
    );
};

// Tab component
const TabButton = ({ active, label, onClick }) => {
    return (
        <button
            className={`px-4 py-2 font-medium rounded-md text-sm transition-all ${
                active
                    ? 'bg-indigo-100 text-indigo-700 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
            }`}
            onClick={onClick}
        >
            {label}
        </button>
    );
};

const Dashboard = () => {
    const [activeTab, setActiveTab] = useState('overview');
    const [file, setFile] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [reportData, setReportData] = useState(null);
    const [uploadSuccess, setUploadSuccess] = useState(false);

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
            const formData = new FormData();
            formData.append('report', file);

            const response = await fetch('/api/parse-report', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(`Server returned ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            setReportData(data);
            setUploadSuccess(true);
            setActiveTab('overview');
        } catch (err) {
            console.error('Error processing file:', err);
            setError(`Failed to process the file: ${err.message || 'Unknown error'}`);
        } finally {
            setLoading(false);
        }
    };

    // Reset form for new upload
    const handleReset = () => {
        setFile(null);
        setReportData(null);
        setUploadSuccess(false);
        setError(null);
    };

    // Render upload section based on state
    const renderUploadSection = () => {
        if (uploadSuccess) {
            return (
                <div className="bg-white rounded-lg shadow p-4 mb-6">
                    <div className="flex justify-between items-center">
                        <h2 className="text-lg font-semibold">Report Uploaded Successfully</h2>
                        <button
                            onClick={handleReset}
                            className="text-indigo-600 hover:text-indigo-800 underline text-sm"
                        >
                            Upload Another Report
                        </button>
                    </div>
                    <div className="mt-2 p-3 bg-green-50 border border-green-100 rounded-md text-sm">
                        <p className="text-green-800">
                            <span className="font-medium">File:</span> {file.name} ({(file.size / 1024).toFixed(1)} KB)
                        </p>
                        <p className="text-green-700 mt-1">
                            Report analysis complete. View the results below.
                        </p>
                    </div>
                </div>
            );
        }

        return (
            <div className="bg-white rounded-lg shadow p-6 mb-6">
                <h2 className="text-xl font-semibold mb-4">Upload Health Check Report</h2>
                <p className="mb-4 text-gray-600">
                    Upload an OpenShift Health Check Report (AsciiDoc format) to visualize the results.
                </p>

                <div className="flex flex-col sm:flex-row space-y-4 sm:space-y-0 sm:space-x-4">
                    <div className="flex-1">
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Select Report File
                        </label>
                        <input
                            type="file"
                            accept=".adoc,.asciidoc"
                            onChange={handleFileChange}
                            className="block w-full text-sm text-gray-500
              file:mr-4 file:py-2 file:px-4
              file:rounded-md file:border-0
              file:text-sm file:font-medium
              file:bg-indigo-50 file:text-indigo-700
              hover:file:bg-indigo-100
              transition-colors duration-300"
                        />
                        {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
                    </div>
                    <div className="flex items-end">
                        <button
                            onClick={handleUpload}
                            disabled={!file || loading}
                            className={`inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white transition-all duration-300
                ${!file || loading ? 'bg-gray-400 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 hover:shadow-md'}`}
                        >
                            {loading ? (
                                <>
                                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Processing...
                                </>
                            ) : (
                                'Upload and Analyze'
                            )}
                        </button>
                    </div>
                </div>

                {file && !error && !loading && (
                    <div className="mt-4 p-3 bg-blue-50 border border-blue-100 rounded-md">
                        <p className="text-sm text-blue-800">
                            <span className="font-medium">Selected file:</span> {file.name} ({(file.size / 1024).toFixed(1)} KB)
                        </p>
                        <p className="text-xs text-blue-700 mt-1">
                            Click "Upload and Analyze" to process the report.
                        </p>
                    </div>
                )}
            </div>
        );
    };

    // Render tabs navigation
    const renderTabNavigation = () => {
        return (
            <div className="bg-white rounded-lg shadow p-2 mb-6">
                <div className="flex space-x-2">
                    <TabButton
                        active={activeTab === 'overview'}
                        label="Overview"
                        onClick={() => setActiveTab('overview')}
                    />
                    <TabButton
                        active={activeTab === 'executive'}
                        label="Executive Summary"
                        onClick={() => setActiveTab('executive')}
                    />
                    <TabButton
                        active={activeTab === 'remediation'}
                        label="Remediation Actions"
                        onClick={() => setActiveTab('remediation')}
                    />
                </div>
            </div>
        );
    };

    // Render overview tab
    const renderOverviewTab = () => {
        if (!reportData) return null;

        const requiredCount = reportData.itemsRequired ? reportData.itemsRequired.length : 0;
        const recommendedCount = reportData.itemsRecommended ? reportData.itemsRecommended.length : 0;
        const advisoryCount = reportData.itemsAdvisory ? reportData.itemsAdvisory.length : 0;
        const noChangeCount = reportData.noChangeCount || 0;

        return (
            <div>
                {/* Overall Health Section */}
                <div className="bg-white rounded-lg shadow p-6 mb-6">
                    <h2 className="text-xl font-semibold mb-2">Overall Health</h2>
                    <div className="flex justify-center">
                        <CircularProgress value={parseFloat(reportData.overallScore)} />
                    </div>
                </div>

                {/* Category Scores Section */}
                <div className="bg-white rounded-lg shadow p-6 mb-6">
                    <h2 className="text-xl font-semibold mb-6">Category Scores</h2>
                    <div>
                        <ProgressBar name="Infrastructure Setup" score={reportData.scoreInfra} />
                        <ProgressBar name="Policy Governance" score={reportData.scoreGovernance} />
                        <ProgressBar name="Compliance Benchmarking" score={reportData.scoreCompliance} />
                        <ProgressBar name="Monitoring" score={reportData.scoreMonitoring} />
                        <ProgressBar name="Build/Deploy Security" score={reportData.scoreBuildSecurity} />
                    </div>
                </div>

                {/* Actions Required Section */}
                <div className="bg-white rounded-lg shadow p-6">
                    <h2 className="text-xl font-semibold mb-6">Actions Required</h2>
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <ActionBox
                            count={requiredCount}
                            label="Required"
                            color="text-red-700"
                            bgColor="bg-red-50"
                        />
                        <ActionBox
                            count={recommendedCount}
                            label="Recommended"
                            color="text-yellow-700"
                            bgColor="bg-yellow-50"
                        />
                        <ActionBox
                            count={advisoryCount}
                            label="Advisory"
                            color="text-blue-700"
                            bgColor="bg-blue-50"
                        />
                        <ActionBox
                            count={noChangeCount}
                            label="No Change"
                            color="text-green-700"
                            bgColor="bg-green-50"
                        />
                    </div>
                </div>
            </div>
        );
    };

    // Render Executive Summary Tab
    const renderExecutiveSummaryTab = () => {
        if (!reportData) return null;

        return (
            <div className="bg-white rounded-lg shadow p-6">
                <h2 className="text-xl font-semibold mb-6">Executive Summary</h2>
                <div className="space-y-6">
                    <div>
                        <h3 className="text-lg font-medium mb-2 text-gray-800">Cluster Overview</h3>
                        <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <p className="text-sm text-gray-500">Cluster Name</p>
                                    <p className="font-medium">{reportData.clusterName || 'N/A'}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-gray-500">Customer</p>
                                    <p className="font-medium">{reportData.customerName || 'N/A'}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-gray-500">Overall Health</p>
                                    <p className="font-medium">{reportData.overallScore || 0}% ({getScoreRating(reportData.overallScore || 0)})</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div>
                        <h3 className="text-lg font-medium mb-2 text-gray-800">Health Assessment</h3>
                        <div className="bg-gray-50 p-4 rounded-lg">
                            <p>The cluster health assessment resulted in an overall score of <span className="font-semibold">{reportData.overallScore || 0}%</span>, which is considered <span className="font-semibold">{getScoreRating(reportData.overallScore || 0)}</span>.</p>

                            <p className="mt-2">
                                {reportData.itemsRequired && reportData.itemsRequired.length > 0
                                    ? `There ${reportData.itemsRequired.length === 1 ? 'is' : 'are'} ${reportData.itemsRequired.length} critical ${reportData.itemsRequired.length === 1 ? 'issue' : 'issues'} requiring immediate attention.`
                                    : 'No critical issues requiring immediate attention were found.'}

                                {reportData.itemsRecommended && reportData.itemsRecommended.length > 0 && ` Additionally, ${reportData.itemsRecommended.length} ${reportData.itemsRecommended.length === 1 ? 'change is' : 'changes are'} recommended to improve cluster health.`}
                            </p>
                        </div>
                    </div>

                    <div>
                        <h3 className="text-lg font-medium mb-2 text-gray-800">Category Breakdown</h3>
                        <div className="bg-gray-50 p-4 rounded-lg grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <p className="text-sm font-medium">Infrastructure Setup</p>
                                <p className="text-sm text-gray-600">{reportData.scoreInfra || 0}% - {getScoreRating(reportData.scoreInfra || 0)}</p>
                                <p className="text-xs text-gray-500 mt-1">{reportData.infraDescription || 'No description available'}</p>
                            </div>
                            <div>
                                <p className="text-sm font-medium">Policy Governance</p>
                                <p className="text-sm text-gray-600">{reportData.scoreGovernance || 0}% - {getScoreRating(reportData.scoreGovernance || 0)}</p>
                                <p className="text-xs text-gray-500 mt-1">{reportData.governanceDescription || 'No description available'}</p>
                            </div>
                            <div>
                                <p className="text-sm font-medium">Compliance</p>
                                <p className="text-sm text-gray-600">{reportData.scoreCompliance || 0}% - {getScoreRating(reportData.scoreCompliance || 0)}</p>
                                <p className="text-xs text-gray-500 mt-1">{reportData.complianceDescription || 'No description available'}</p>
                            </div>
                            <div>
                                <p className="text-sm font-medium">Monitoring</p>
                                <p className="text-sm text-gray-600">{reportData.scoreMonitoring || 0}% - {getScoreRating(reportData.scoreMonitoring || 0)}</p>
                                <p className="text-xs text-gray-500 mt-1">{reportData.monitoringDescription || 'No description available'}</p>
                            </div>
                            <div>
                                <p className="text-sm font-medium">Build/Deploy Security</p>
                                <p className="text-sm text-gray-600">{reportData.scoreBuildSecurity || 0}% - {getScoreRating(reportData.scoreBuildSecurity || 0)}</p>
                                <p className="text-xs text-gray-500 mt-1">{reportData.buildSecurityDescription || 'No description available'}</p>
                            </div>
                        </div>
                    </div>

                    <div className="flex justify-end">
                        <button className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-colors">
                            Download Executive Summary
                        </button>
                    </div>
                </div>
            </div>
        );
    };

    // Render Remediation Actions Tab
    const renderRemediationTab = () => {
        if (!reportData) return null;

        const requiredItems = reportData.itemsRequired || [];
        const recommendedItems = reportData.itemsRecommended || [];
        const advisoryItems = reportData.itemsAdvisory || [];

        return (
            <div className="bg-white rounded-lg shadow p-6">
                <h2 className="text-xl font-semibold mb-6">Remediation Actions</h2>

                <div className="space-y-6">
                    {/* Required Actions */}
                    <div>
                        <h3 className="text-lg font-medium mb-4 text-red-700">Required Actions ({requiredItems.length})</h3>
                        {requiredItems.length > 0 ? (
                            <div className="space-y-3">
                                {requiredItems.map((item, index) => (
                                    <div key={`req-${index}`} className="p-3 bg-red-50 border border-red-100 rounded-md">
                                        <p className="text-red-800 font-medium">{item}</p>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="p-3 bg-green-50 border border-green-100 rounded-md">
                                <p className="text-green-800">No required actions identified. All critical checks passed.</p>
                            </div>
                        )}
                    </div>

                    {/* Recommended Actions */}
                    <div>
                        <h3 className="text-lg font-medium mb-4 text-yellow-700">Recommended Actions ({recommendedItems.length})</h3>
                        {recommendedItems.length > 0 ? (
                            <div className="space-y-3">
                                {recommendedItems.map((item, index) => (
                                    <div key={`rec-${index}`} className="p-3 bg-yellow-50 border border-yellow-100 rounded-md">
                                        <p className="text-yellow-800 font-medium">{item}</p>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="p-3 bg-green-50 border border-green-100 rounded-md">
                                <p className="text-green-800">No recommended actions identified.</p>
                            </div>
                        )}
                    </div>

                    {/* Advisory Actions */}
                    <div>
                        <h3 className="text-lg font-medium mb-4 text-blue-700">Advisory Items ({advisoryItems.length})</h3>
                        {advisoryItems.length > 0 ? (
                            <div className="space-y-3">
                                {advisoryItems.map((item, index) => (
                                    <div key={`adv-${index}`} className="p-3 bg-blue-50 border border-blue-100 rounded-md">
                                        <p className="text-blue-800 font-medium">{item}</p>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="p-3 bg-green-50 border border-green-100 rounded-md">
                                <p className="text-green-800">No advisory items identified.</p>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div className="min-h-screen bg-gray-50 pb-10">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                {/* Header */}
                <div className="bg-white rounded-lg shadow p-4 mb-6">
                    <div className="flex flex-col md:flex-row justify-between items-center">
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
                {renderUploadSection()}

                {/* Content (only shown after file upload) */}
                {reportData && (
                    <>
                        {renderTabNavigation()}

                        {activeTab === 'overview' && renderOverviewTab()}
                        {activeTab === 'executive' && renderExecutiveSummaryTab()}
                        {activeTab === 'remediation' && renderRemediationTab()}
                    </>
                )}

                {/* Footer */}
                <div className="text-center text-xs text-gray-500 mt-8">
                    <p>OpenShift Health Check Dashboard • Analyze and visualize health check reports</p>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;